
# Certificate Generation

This library offers lambda expressions that are able to
maintain and update certificates and keys.

The package name is `utilities.certs`.
Required packages:
- [`utilities.state`](../state/README.md)

The offered functions work together with the _state_ library 
and therefore support the update feature if the state mechanism
is used in the spiff usage scenario. Nevertheless the 
functions can also be used without the state support. In this case
they just generate new keys and certificates for every run.

It is based on the [x509 functions](../../README.md#x509-functions) offered
by _spiff_.

The generated values can be found under the `value` sub-node according
to the rules of the _state_ package. The second result field is `input`
containing the completed certificate/secret specification. The `update` parameter
can be set to `true` is an update should be enforced. This feature only works
when using the functions in a stateful scenario.

## Generate a self signed Certificate for dedicated common name

```
    selfSignedCA(<common name>, <update>=false, <relpath>=[]) -> state
```

The _value_ field provides the fields:
- `key` holding the private key
- `pub` holding the public key
- `cert` holding the certificate for the CA.

## Generate a Key/Certificate Pair

```
    keyCertForCA(<certspec>, <ca>, <update>=false, <relpath>=[]) -> state
```

the certificate specification uses the format for the
function [`x509cert`](../../README.md#-x509certspec-), but without
the key and ca related fields. They are implicity adden by the given `ca` and
the generated key. The `ca` is given just by using a reference to a field
set by the `selfSignedCA` function.

The _value_ field provides the fields:
- `key` holding the private key
- `pub` holding the public key
- `cert` holding the certificate signed by the CA

## Generate a Certificate with an explicitly managed Specification

```
    keyCert(<certspec>, <update>=false, <relpath>=[]) -> state
```

the certificate specification uses the format for the
function [`x509cert`](../../README.md#-x509certspec-). It justed adds
the state support to the bare _spiff_ function.

The _value_ field provides the fields:
- `key` holding the private key
- `pub` holding the public key
- `cert` holding the certificate


## Generate an SSH Key Pair

```
    sshKey(<length>=2048, <update>=false, <relpath>=[])  -> state
```

The _value_ field provides the fields:
- `key` holding the private key
- `pub` holding the public key in ssh format


## Generate a Random Secret with a dedicated Length

```
    secret(<default>, <length>, <update>=false, <relpath>=[])  -> string
```

If no `default` (`~`) is given a random string consisting of alphanumeric
character of given length is generated.

The _value_ field directly contains the secret value.

## Generate a Wireguard Key Pair

```
    wireguardKey(<update>=false, <relpath>=[])  -> state
```

The _value_ field provides the fields:
- `key` holding the private key
- `pub` holding the public key

## Tweaking the state access

By default the old state is always accessed using the `stub()` function
to access the same field containing the state lambda in the stub which
is typically the state yaml. This is handled in the [state](../state/README.md)
library. But this only works correctly if
the state expression directly generates the state fields.

The optional relpath parameter can be used to adjust the stub access
(for accessing old state) in case of generating multiple state instances
with `map`/`sum`  generating implicit intermediate sub structures between the
field containing the lambda expression and the generated state field.

for example, when generating wireguard keys for a dynamic set of names:

```yaml
names:
  - alice
  - bob
state:
  <<: (( &state(merge none) ))
  wireguard: (( map{names|m|-> utilities.certs.wireguardKey(false, [m])} ))
```
