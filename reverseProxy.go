package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func Webserver(config UserInput) error {
	targetUrl, err := url.Parse(config.Upstream.URL)
	if err != nil {
		return err
	}
	fmt.Printf("debug targetUrl:\n  Scheme: %s\n  Host: %s\n", targetUrl.Scheme, targetUrl.Host)

	// type ReverseProxy
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// modify request headers
	proxy.Director = func(req *http.Request) {
		// modify header=Host to targetUrl, not default "hm90:8080"
		req.URL.Scheme = targetUrl.Scheme
		req.URL.Host = targetUrl.Host

		// add BasicAuth header
		if config.Upstream.BasicAuth.Username != "" && config.Upstream.BasicAuth.Password != "" {
			auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(config.Upstream.BasicAuth.Username+":"+config.Upstream.BasicAuth.Password))
			req.Header.Set("Authorization", auth)
			fmt.Printf("Authorization header: %s\n", req.Header.Get("Authorization"))
		}

		fmt.Printf("Director debug:\n  req.Host: %s\n\n", req.Host)
	}

	// apply tls.insecureSkipVerify and/or tls.caFile
	if config.Upstream.TLS.InsecureSkipVerify || config.Upstream.TLS.CAFile != "" {
		tlsConfig := &tls.Config{}

		if config.Upstream.TLS.InsecureSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}

		if config.Upstream.TLS.CAFile != "" {
			caCert, err := os.ReadFile(config.Upstream.TLS.CAFile)
			if err != nil {
				return err
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return fmt.Errorf("failed to append CA certificate to pool")
			}
			tlsConfig.RootCAs = caCertPool
		}

		proxy.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	http.HandleFunc("/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
		r.Host = targetUrl.Host
		fmt.Printf("Forwarding request:\n  Scheme: %s\n  Host: %s\n  Method: %s\n  Path: %s\n", r.URL.Scheme, r.Host, r.Method, r.URL.Path)

		// inject labels into the query parameter
		if len(config.InjectLabels) > 0 {
			query := r.URL.Query()
			originalQuery := query.Get("query")
			if originalQuery != "" {
				// Build label injection string
				labelSelector := ""
				for labelName, labelValue := range config.InjectLabels {
					if labelSelector == "" {
						labelSelector = fmt.Sprintf(`%s="%s"`, labelName, labelValue)
					} else {
						labelSelector += fmt.Sprintf(`,%s="%s"`, labelName, labelValue)
					}
				}

				// Inject labels by adding them as a vector selector
				injectedQuery := fmt.Sprintf(`{%s} and on() (%s)`, labelSelector, originalQuery)
				query.Set("query", injectedQuery)
				r.URL.RawQuery = query.Encode()
			}
		}

		proxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "invalid path: %s\n", r.URL.Path)
	})

	fmt.Printf("starting webserver\n\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return err
	}
	return nil
}
