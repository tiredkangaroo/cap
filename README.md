readme coming soon

env variables:

- PROXY_CACERT: relative (from the execution point) or absolute path of a certificate authority certificate. The public cert is expected to be a PEM-encoded X509 certificate.

- PROXY_CAKEY: relative (from the execution point) or absolute path of the certificate authority private key. This will be used to sign the MTIM certificates generated. The private key is expected to be a PEM-encoded PKCS8 private key.

- PROXY_CONFIG_FILE: relative (from the execution point) or absolute path of the file containing the proxy configuration. This
file doesn't have to exist yet, it will be created if it doesn't exist. It is recommended to add the .json extension to the file name.
