## ic get clusters

Get list of clusters

### Synopsis

Get list of clusters.

Supported field names for filters:

name, description, clusterID, clusterType, region, environmentName,
providerName, navisionSubscriptionNumber, navisionCustomerNumber,
navisionCustomerName, resilienceZone, clientVersion, kubernetesVersion


```
ic get clusters [flags]
```

### Examples

```

# get all cluster
ic get clusters

# get clusters in the resilience zone 'platform'
ic get clusters --filter resilienceZone=platform

use: 'ic help filters' for more information on using filters
```

### Options

```
      --filter stringArray   Filter output based on conditions
  -h, --help                 help for clusters
```

### Options inherited from parent commands

```
  -s, --api-server string                            URL for the inventory server. (default "https://api.k8s.netic.dk")
  -d, --debug                                        Debug mode
  -f, --force                                        Force actions
      --log-format string                            Log format (plain|json) (default "plain")
      --log-level string                             Log level (debug|info|warn|error) (default "info")
      --no-color                                     Do not print color
      --no-headers                                   Do not print headers
      --no-input                                     Assume non-interactive mode
      --oidc-auth-bind-addr string                   [authcode-browser] Bind address and port for local server used for OIDC redirect (default "localhost:18000")
      --oidc-client-id string                        OIDC client ID (default "inventory-cli")
      --oidc-grant-type string                       OIDC authorization grant type. One of (authcode-browser|authcode-keyboard) (default "authcode-browser")
      --oidc-issuer-url string                       Issuer URL for the OIDC Provider (default "https://keycloak.netic.dk/auth/realms/mcs")
      --oidc-redirect-uri-authcode-keyboard string   [authcode-keyboard] Redirect URI when using authcode keyboard (default "urn:ietf:wg:oauth:2.0:oob")
      --oidc-redirect-url-hostname string            [authcode-browser] Hostname of the redirect URL (default "localhost")
      --oidc-token-cache-dir string                  Directory used to store cached tokens (default "/Users/kn/Library/Caches/ic/oidc-login")
  -o, --output string                                Output format (default "plain")
```

### SEE ALSO

* [ic get](ic_get.md)	 - Add one or many resources

###### Auto generated by spf13/cobra on 17-Mar-2025
