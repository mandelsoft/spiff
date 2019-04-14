package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flowing YAML for validation", func() {
	Context("cidr", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("1.2.3.4/20", "cidr") ))
`)
			resolved := parseYAML(`
---
val: 1.2.3.4/20
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("1.2.3.4/200", "cidr")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: "condition 1 failed: is no CIDR: invalid CIDR address: 1.2.3.4/200"
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("ip", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("1.2.3.4", "ip") ))
`)
			resolved := parseYAML(`
---
val: 1.2.3.4
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("1.2.3.4.5", "ip")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: "condition 1 failed: is no ip address: 1.2.3.4.5"
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("wildcarddnsdomain", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("*.mandelsoft.org", "wildcarddnsdomain") ))
`)
			resolved := parseYAML(`
---
val: "*.mandelsoft.org"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("mandelsoft.org", "wildcarddnsdomain")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is no wildcard dns domain: [a wildcard DNS-1123 subdomain
  	    must start with ''*.'', followed by a valid DNS subdomain, which must consist
  	    of lower case alphanumeric characters, ''-'' or ''.'' and end with an alphanumeric
  	    character (e.g. ''*.example.com'', regex used for validation is ''\*\.[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'')]'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("dnsdomain", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("spiff.mandelsoft.org", "dnsdomain") ))
`)
			resolved := parseYAML(`
---
val: "spiff.mandelsoft.org"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("*.mandelsoft.org", "dnsdomain")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is no dns domain: [a DNS-1123 subdomain must consist
  	    of lower case alphanumeric characters, ''-'' or ''.'', and must start and end
  	    with an alphanumeric character (e.g. ''example.com'', regex used for validation
  	    is ''[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'')]'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("dnslabel", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("alice-bob", "dnslabel") ))
`)
			resolved := parseYAML(`
---
val: "alice-bob"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("alice+bob", "dnslabel")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is no dns label: [a DNS-1123 label must consist of lower
  	    case alphanumeric characters or ''-'', and must start and end with an alphanumeric
  	    character (e.g. ''my-name'',  or ''123-abc'', regex used for validation is ''[a-z0-9]([-a-z0-9]*[a-z0-9])?'')]'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("dnsname", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val1: (( validate("spiff.mandelsoft.org", "dnsname") ))
val2: (( validate("*.mandelsoft.org", "dnsname") ))
`)
			resolved := parseYAML(`
---
val1: "spiff.mandelsoft.org"
val2: "*.mandelsoft.org"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("alice+bob", "dnsname")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is no dns name: [a DNS-1123 subdomain must consist of
  	    lower case alphanumeric characters, ''-'' or ''.'', and must start and end with
  	    an alphanumeric character (e.g. ''example.com'', regex used for validation is
  	    ''[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'')]'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("type", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("spiff.mandelsoft.org", [ "type", "string" ]) ))
`)
			resolved := parseYAML(`
---
val: "spiff.mandelsoft.org"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate([], [ "type", "string" ])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is not of type string'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("match", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("alice", [ "match", "a.*e" ]) ))
`)
			resolved := parseYAML(`
---
val: "alice"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("bob", [ "match", "a.*e" ])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: invalid value "bob"'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("valueset", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("alice", [ "valueset", [ "alice", "bob" ] ]) ))
`)
			resolved := parseYAML(`
---
val: "alice"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("peter", [ "valueset",  [ "alice", "bob" ] ])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: invalid value "peter"'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("empty", func() {
		It("accepts string", func() {
			source := parseYAML(`
---
val: (( validate("", "empty") ))
`)
			resolved := parseYAML(`
---
val: ""
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects string", func() {
			source := parseYAML(`
---
val: (( catch(validate("foobar", "empty")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is not empty'
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("accepts map", func() {
			source := parseYAML(`
---
val: (( validate({}, "empty") ))
`)
			resolved := parseYAML(`
---
val: {}
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects map", func() {
			source := parseYAML(`
---
val: (( catch(validate({ $foo="bar"}, "empty")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is not empty'
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("accepts list", func() {
			source := parseYAML(`
---
val: (( validate([], "empty") ))
`)
			resolved := parseYAML(`
---
val: []
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects list", func() {
			source := parseYAML(`
---
val: (( catch(validate(["foobar"], "empty")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is not empty'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("privatekey", func() {
		It("accepts", func() {
			source := parseYAML(`
---
key: |
  -----BEGIN RSA PRIVATE KEY-----
  MD4CAQACCQDKkkX0E4sNbQIDAQABAggqWiaxn0tEwQIFAP6cjEUCBQDLrRMJAgRW
  w4g1AgQEpjbBAgUA0oHQfw==
  -----END RSA PRIVATE KEY-----
val: (( validate(key, "privatekey") ))
`)
			resolved := parseYAML(`
---
key: |
  -----BEGIN RSA PRIVATE KEY-----
  MD4CAQACCQDKkkX0E4sNbQIDAQABAggqWiaxn0tEwQIFAP6cjEUCBQDLrRMJAgRW
  w4g1AgQEpjbBAgUA0oHQfw==
  -----END RSA PRIVATE KEY-----
val: |
  -----BEGIN RSA PRIVATE KEY-----
  MD4CAQACCQDKkkX0E4sNbQIDAQABAggqWiaxn0tEwQIFAP6cjEUCBQDLrRMJAgRW
  w4g1AgQEpjbBAgUA0oHQfw==
  -----END RSA PRIVATE KEY-----
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("peter", "privatekey")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is no private key: invalid private key format (expected
  	    pem block)'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("publickey", func() {
		It("accepts", func() {
			source := parseYAML(`
---
key: |
  -----BEGIN RSA PUBLIC KEY-----
  MBACCQC60m+vYsCt7wIDAQAB
  -----END RSA PUBLIC KEY-----
val: (( validate(key, "publickey") ))
`)
			resolved := parseYAML(`
---
key: |
  -----BEGIN RSA PUBLIC KEY-----
  MBACCQC60m+vYsCt7wIDAQAB
  -----END RSA PUBLIC KEY-----
val: |
  -----BEGIN RSA PUBLIC KEY-----
  MBACCQC60m+vYsCt7wIDAQAB
  -----END RSA PUBLIC KEY-----
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("peter", "publickey")) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is no public key: invalid public key format (expected
  	    pem block)'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("or", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("alice", [ "or", "empty", "dnslabel" ]) ))
`)
			resolved := parseYAML(`
---
val: "alice"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("alice", [ "or", "empty", "!dnslabel" ])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: (is not empty and is dns label)'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("and", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate("alice", [ "and", "!empty", "dnslabel" ]) ))
`)
			resolved := parseYAML(`
---
val: "alice"
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate("alice", [ "and", "empty", "dnslabel" ])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is not empty'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("mapfield", func() {
		It("accept exists", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "mapfield", "alice" ]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("accept validated", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "mapfield", "alice", [ "type", "int" ]]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("fail exists", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( catch(validate(map, [ "mapfield", "bob" ])) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  valid: false
  error: 'condition 1 failed: has no field "bob"'
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("fail validator", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( catch(validate(map, [ "mapfield", "alice", [ "type", "string"]])) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  valid: false
  error: 'condition 1 failed: map entry "alice" is not of type string'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("optionalfield", func() {
		It("accept exists", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "optionalfield", "alice" ]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("accept validated", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "optionalfield", "alice", [ "type", "int" ]]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("accept not exists", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "optionalfield", "bob" ]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("fail validator", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( catch(validate(map, [ "optionalfield", "alice", [ "type", "string"]])) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  valid: false
  error: 'condition 1 failed: map entry "alice" is not of type string'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("map", func() {
		It("accepts map", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, "map") ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("accept keys", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "map", "dnslabel", ~ ]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("accept values", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "map", ["type", "int"]  ]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("accept key and value", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( validate(map, [ "map", "dnslabel", ["type", "int"]  ]) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("rejects non-map", func() {
			source := parseYAML(`
---
map: []
val: (( catch(validate(map, "map")) ))
`)
			resolved := parseYAML(`
---
map: []
val:
  valid: false
  error: 'condition 1 failed: is no map'
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("rejects keys", func() {
			source := parseYAML(`
---
map:
 alice: 25
val: (( catch(validate(map, [ "map", "ip", ~ ])) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  valid: false
  error: 'condition 1 failed: map key "alice" is no ip address: alice'
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("reject values", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( catch(validate(map, [ "map", ["type", "string"]  ])) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  valid: false
  error: 'condition 1 failed: map entry "alice" is not of type string'
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("reject key and value", func() {
			source := parseYAML(`
---
map:
  alice: 25
val: (( catch(validate(map, [ "map", "dnslabel", ["type", "string"]  ])) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
val:
  valid: false
  error: 'condition 1 failed: map entry "alice" is not of type string'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("lambda", func() {
		It("accepts", func() {
			source := parseYAML(`
---
val: (( validate(5, |x|-> x >= 5) ))
`)
			resolved := parseYAML(`
---
val: 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("rejects", func() {
			source := parseYAML(`
---
val: (( catch(validate(4, |x|-> x >= 5)) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: lambda|x|->x >= 5 failed'
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("accepts with arg", func() {
			source := parseYAML(`
---
val: (( validate(5, [ |x,m|-> x >= m, 5 ]) ))
`)
			resolved := parseYAML(`
---
val: 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("rejects with arg", func() {
			source := parseYAML(`
---
val: (( catch(validate(4, [ |x,m|-> x >= m, 5 ])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: lambda|x,m|->x >= m failed'
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("accepts with arg and message", func() {
			source := parseYAML(`
---
val: (( validate(5, [ |x,m|-> [x >= m, "is larger than or equal to " m], 5 ]) ))
`)
			resolved := parseYAML(`
---
val: 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("rejects with arg and message", func() {
			source := parseYAML(`
---
val: (( catch(validate(5, [ "not" ,[|x,m|-> [ x >= m, "is larger than or equal to " m] , 5 ]])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is larger than or equal to 5'
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("accepts with arg and messages", func() {
			source := parseYAML(`
---
val: (( validate(5, [ |x,m|-> [x >= m, "is larger than or equal to " m, "is less than " m], 5 ]) ))
`)
			resolved := parseYAML(`
---
val: 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("rejects with arg and messages 1", func() {
			source := parseYAML(`
---
val: (( catch(validate(5, [ "!" , [ |x,m|-> [x >= m, "is larger than or equal to " m, "is less than " m], 5 ]])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is larger than or equal to 5'
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("rejects with arg and message 2", func() {
			source := parseYAML(`
---
val: (( catch(validate(4, [|x,m|-> [ x >= m, "is larger than or equal to " m, "is less than " m] , 5 ])) ))
`)
			resolved := parseYAML(`
---
val:
  valid: false
  error: 'condition 1 failed: is less than 5'
`)
			Expect(source).To(FlowAs(resolved))
		})
	})
})
