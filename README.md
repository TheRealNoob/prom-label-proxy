Listens on port 8080 at the moment.

define configuration either via ENV VARS, --cmdline-args, or config file.  Config file can be json, yaml, or etc. other formats.

default config file
```yaml
upstream:
  url: ""
  basicAuth:
    username: ""
    password: ""
  tls:
    insecureSkipVerify: false
    caFile: ""
```
