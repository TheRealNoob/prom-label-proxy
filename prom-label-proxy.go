package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var configFile string

var cmdRoot = &cobra.Command{
	Use:   "prom-label-proxy",
	Short: "Run promQL proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the config file if specified via the -c/--config-file flag
		if configFile != "" {
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
		}

		// Unmarshal config and validate input
		var config UserInput
		if err := viper.Unmarshal(&config); err != nil {
			fmt.Println("Error unmarshaling config")
			return err
		}
		if err := config.Validate(); err != nil {
			return err
		}

		// Marshal the config struct to YAML
		yamlOutput, err := yaml.Marshal(&config)
		if err != nil {
			fmt.Println("Error marshaling config to YAML")
			return err
		}
		fmt.Println("Configuration:")
		fmt.Println(string(yamlOutput))

		return Webserver(config)
	},
}

func init() {
	// Define persistent flags on the root command
	cmdRoot.PersistentFlags().StringVarP(&configFile, "config-file", "c", "", "config file path")
	cmdRoot.PersistentFlags().String("upstream-url", "https://example.com", "upstream URL")
	cmdRoot.PersistentFlags().String("upstream-basicAuth-username", "", "basic auth username")
	cmdRoot.PersistentFlags().String("upstream-basicAuth-password", "", "basic auth password")
	cmdRoot.PersistentFlags().Bool("upstream-tls-insecureSkipVerify", false, "skip TLS verification")
	cmdRoot.PersistentFlags().String("upstream-tls-caFile", "", "path to CA file")

	// Bind flags to Viper keys
	viper.BindPFlag("upstream.url", cmdRoot.PersistentFlags().Lookup("upstream-url"))
	viper.BindPFlag("upstream.basicAuth.username", cmdRoot.PersistentFlags().Lookup("upstream-basicAuth-username"))
	viper.BindPFlag("upstream.basicAuth.password", cmdRoot.PersistentFlags().Lookup("upstream-basicAuth-password"))
	viper.BindPFlag("upstream.tls.insecureSkipVerify", cmdRoot.PersistentFlags().Lookup("upstream-tls-insecureSkipVerify"))
	viper.BindPFlag("upstream.tls.caFile", cmdRoot.PersistentFlags().Lookup("upstream-tls-caFile"))

	// Configure environment variables
	viper.SetEnvPrefix("PROMLABELPROXY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("upstream.url", "https://example.com")
	viper.SetDefault("upstream.basicAuth.username", "")
	viper.SetDefault("upstream.basicAuth.password", "")
	viper.SetDefault("upstream.tls.insecureSkipVerify", false)
	viper.SetDefault("upstream.tls.caFile", "")
}

func main() {
	if err := cmdRoot.Execute(); err != nil {
		panic(err)
	}
}
