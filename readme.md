# Lastpass-Controller

### What is this?

This is a very simple app that will sync secrets from your lastpass vault into a kubernetes cluster.

### Is this safe? 

I wouldn't count on it. It downloads and decrypts your entire vault from a single password. I strongly recommend a dedicated lastpass account for this. There is no granular access control or anything like that. Suitable for home lab, but I'd avoid where serious security is a concern.

### How to use.

Just make a `ConfigMap` with the label `lastpass-secret: true`. The data should be names of fields, and the corresponding lastpass account name for the fields. The controller will create a secret with the same name as the config map, and the same field names, with values filled in from lastpass. It will overwrite any existing secret data with that name, so be careful.