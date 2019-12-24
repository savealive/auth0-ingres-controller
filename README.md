# Auth0 Ingress controller
Updates Auth0 callback URLs and allowed origins

## Build:
```
$ make linux
```
## Usage
* Create auth0 application https://manage.auth0.com/dashboard of type "machine to machine" with permissions to use target "Auth0 Management API" as we're about to update apps properties
* Update charts/auth0-ingress-controller/values.yaml with created app data.
this credentials will be used to update managed apps
```
client:
  clientID: T1k1ne5gc40jE_RVZSt49u5a3AXgHKtq
  clientSecret: F8GpXcsqtKd0iLoQEthvk0KJaBFosZsQTLZNiqU8hf8pAe_vozjr_COWNnsl0SIB
  domain: yourdomain.eu.auth0.com
  apiURL: https://domain.api.url
enableCallbackDeletion: true
creationDelay: 1s
configFilePath: /etc/auth0-ingress-controller/config.yaml
```
* Deploy helm chart (keep in mind helm version and namespace to deploy)
```
helm install auth0 charts/auth0-ingress-controller
```
* In order to update Auth0 app with callback and weborigin URL annotate ingress as following
```
kind: Ingress
metadata:
  annotations:
    auth0.creditplace.com/appid: 6SaohIZkUfrL6tdDCk3TcXxmS4P5CS8C
    auth0.creditplace.com/callbackPath: auth0loginCallback
    auth0.creditplace.com/enabled: "true"
```
Where _auth0.creditplace.com/appid: 6SaohIZkUfrL6tdDCk3TcXxmS4P5CS8C_ is app id of the _managed_ app.
*callbackPath* will be added to ingress path e.g. if ingress was http://my.app.com and callbackPath=auth0loginCallback resuling string will be http://my.app.com/auth0loginCallback. 
WebOrigin will be set to Ingress spec.rules.host (use with caution with more than one host, it's not quite tested)
