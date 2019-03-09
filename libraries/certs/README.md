
# Certificate Generation

This library offers lambda expressions that are able to
maintain and update certificates and keys.

The package name is `utilities.certs`.
Required packages:
- utilities.state

The offered functions work together with the _state_ library 
and therefore support the update feature if the state mechanism
is used in the spiff usage scenario. Nevertheless the 
functions can also be used without the state support. In this case
they just generate new keys and certificates for every run.

## Generate a self signed certificate for dediacted common name.

```
    selfSignedCert(common_name)
```

## Generate a random secret with a dedicated length

```
    secret(default, length)  -> string
```

If no default is given a random string of given length is generated.
