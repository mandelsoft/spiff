package flow

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

var _ = Describe("Flowing YAML", func() {
	Context("delays resolution until merge succeeded", func() {
		It("handles combination of inline merge and field merge", func() {
			source := parseYAML(`
---
properties:
  <<: (( merge || nil ))
  bar: (( merge ))

foobar:
  - (( "foo." .properties.bar ))
`)
			stub := parseYAML(`
---
properties:
  bar: bar
`)

			resolved := parseYAML(`
---
properties:
  bar: bar
foobar: 
  - foo.bar
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("handles defaulted reference to merged/overridden fields", func() {
			source := parseYAML(`
---
foo:
  <<: (( merge || nil ))
  bar:
    <<: (( merge || nil ))
    alice: alice

props:
  bob: (( foo.bar.bob || "wrong" ))
  alice: (( foo.bar.alice || "wrong" ))
  main: (( foo.foo || "wrong" ))

`)
			stub := parseYAML(`
---
foo: 
  foo: added
  bar:
    alice: overwritten
    bob: added!

`)

			resolved := parseYAML(`
---
foo:
  bar:
    alice: overwritten
    bob: added!
  foo: added
props:
  alice: overwritten
  bob: added!
  main: added

`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("handles defaulted reference to merged/overridden redirected fields", func() {
			source := parseYAML(`
---
foo:
  <<: (( merge alt || nil ))
  bar:
    <<: (( merge || nil ))
    alice: alice

props:
  bob: (( foo.bar.bob || "wrong" ))
  alice: (( foo.bar.alice || "wrong" ))
  main: (( foo.foo || "wrong" ))

`)
			stub := parseYAML(`
---
foo:
  bar:
    alice: wrongly merged
alt:
  foo: added
  bar:
    alice: overwritten
    bob: added!

`)

			resolved := parseYAML(`
---
foo:
  bar:
    alice: overwritten
    bob: added!
  foo: added
props:
  alice: overwritten
  bob: added!
  main: added

`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("replaces a non-merge expression node before expanding", func() {
			source := parseYAML(`
---
alt:
  - wrong
properties: (( alt ))
`)
			stub := parseYAML(`
---
properties:
  - right
`)

			resolved := parseYAML(`
---
alt:
  - wrong
properties:
  - right
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("expands a preferred non-merge expression node before overriding", func() {
			source := parseYAML(`
---
alt:
  - right
properties: (( prefer alt ))
`)
			stub := parseYAML(`
---
properties:
  - wrong
`)

			resolved := parseYAML(`
---
alt:
  - right
properties:
  - right
`)
			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Context("when there are no dynaml nodes", func() {
		It("is a no-op", func() {
			source := parseYAML(`
---
foo: bar
`)

			Expect(source).To(FlowAs(source))
		})
	})

	Context("when there are no dynaml nodes", func() {
		It("is a no-op", func() {
			source := parseYAML(`
---
foo: bar
`)

			Expect(source).To(FlowAs(source))
		})
	})

	Context("when there are no dynaml nodes", func() {
		It("is a no-op", func() {
			source := parseYAML(`
---
foo: bar
`)

			Expect(source).To(FlowAs(source))
		})
	})

	Context("when a value is defined in the template and a stub", func() {
		It("overrides the value with the stubbed value", func() {
			source := parseYAML(`
---
a: ~
b: 1
c: foo
d: 2.5
fizz: buzz
`)

			stub := parseYAML(`
---
a: b
b: 2
c: bar
d: 3.14
`)

			result := parseYAML(`
---
a: b
b: 2
c: bar
d: 3.14
fizz: buzz
`)
			Expect(source).To(FlowAs(result, stub))
		})

		Context("in a list", func() {
			It("does not override the value", func() {
				source := parseYAML(`
---
- 1
- 2
`)

				stub := parseYAML(`
---
- 3
- 4
`)

				Expect(source).To(FlowAs(source, stub))
			})
		})
	})

	Context("when some dynaml nodes cannot be resolved", func() {
		It("returns an error", func() {
			source := parseYAML(`
---
foo: (( auto ))
`)

			_, err := Flow(source)
			Expect(err).To(Equal(dynaml.UnresolvedNodes{
				Nodes: []dynaml.UnresolvedNode{
					{
						Node: yaml.IssueNode(yaml.NewNode(
							dynaml.AutoExpr{Path: []string{"foo"}},
							"test",
						), true, false, yaml.NewIssue("auto only allowed for size entry in resource pools")),
						Context: []string{"foo"},
						Path:    []string{"foo"},
					},
				},
			}))
		})
	})

	Context("when there are ignorable dynaml nodes start with '!'", func() {
		It("ignores nodes", func() {
			source := parseYAML(`
---
foo: ((!template_only.foo))
`)

			resolved := parseYAML(`
---
foo: ((!template_only.foo))
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("when a reference is made to a yet-to-be-resolved node, in a || expression", func() {
		It("eventually resolves to the referenced node", func() {
			source := parseYAML(`
---
properties:
  template_only: (( merge ))
  something: (( template_only.foo || "wrong" ))
`)

			stub := parseYAML(`
---
properties:
  template_only:
    foo: right
`)

			resolved := parseYAML(`
---
properties:
  template_only:
    foo: right
  something: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Context("when a refence is made to an unresolveable node", func() {
		It("fails to flow", func() {
			source := parseYAML(`
---
properties:
  template_only: (( abc ))
  something: (( template_only.foo ))
`)

			_, err := Flow(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when a reference is made to an unresolveable node, in a || expression", func() {
		It("eventually resolves to the referenced node", func() {
			source := parseYAML(`
---
properties:
  template_only: (( merge ))
  something: (( template_only.foo || "right" ))
`)

			stub := parseYAML(`
---
properties:
  template_only:
`)

			resolved := parseYAML(`
---
properties:
  template_only:
  something: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("basic dynaml nodes", func() {
		It("evaluates the nodes", func() {
			source := parseYAML(`
---
foo:
  - (( "hello, world!" ))
  - (( 42 ))
  - (( true ))
  - (( nil ))
`)

			resolved := parseYAML(`
---
foo:
  - hello, world!
  - 42
  - true
  - null
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("reference dynaml nodes", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
foo: (( bar ))
bar: 42
`)

			resolved := parseYAML(`
---
foo: 42
bar: 42
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("follows lexical scoping semantics", func() {
			source := parseYAML(`
---
foo:
  bar:
    baz: (( buzz.fizz ))
  buzz:
    fizz: right
buzz:
  fizz: wrong
`)

			resolved := parseYAML(`
---
foo:
  bar:
    baz: right
  buzz:
    fizz: right
buzz:
  fizz: wrong
`)

			Expect(source).To(FlowAs(resolved))
		})

		Context("when the reference starts with .", func() {
			It("starts from the root of the template", func() {
				source := parseYAML(`
---
foo:
  bar:
    baz: (( .bar.buzz ))
    buzz: 42
bar:
  buzz: 43
`)

				resolved := parseYAML(`
---
foo:
  bar:
    baz: 43
    buzz: 42
bar:
  buzz: 43
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("when the referred node is dynamic", func() {
			It("evaluates with their environment", func() {
				source := parseYAML(`
---
foo:
  bar:
    baz: (( buzz.fizz ))
    quux: wrong
buzz:
  fizz: (( quux ))
  quux: right
`)

				resolved := parseYAML(`
---
foo:
  bar:
    baz: right
    quux: wrong
buzz:
  fizz: right
  quux: right
`)

				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("merging in from stubs", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
foo: (( merge ))
bar: 42
`)

			stub := parseYAML(`
---
foo: merged!
`)

			resolved := parseYAML(`
---
foo: merged!
bar: 42
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("follows through maps in lists by name", func() {
			source := parseYAML(`
---
foo:
- name: x
  value: (( merge ))
`)

			stub := parseYAML(`
---
foo:
- name: y
  value: wrong
- name: x
  value: right
`)

			resolved := parseYAML(`
---
foo:
- name: x
  value: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		// this is a regression test, from when Environment.WithPath
		// used append() for adding the next step.
		//
		// using append() will overwrite previous steps, since it reuses the slice
		//
		// e.g. with inital path A:
		//    append(A, "a")
		//    append(A, "b")
		//
		// would result in all previous A/a paths becoming A/b
		It("can be arbitrarily nested", func() {
			source := parseYAML(`
---
properties:
  something:
    foo:
      key: (( merge ))
      val: (( merge ))
`)

			stub := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: b
`)

			resolved := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: b
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("merging fields", func() {
		It("merges locally referenced fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( bar ))
  other: other
bar:
  alice: alice
  bob: bob
`)

			resolved := parseYAML(`
---
foo:
  alice: alice
  bob: bob
  other: other
bar:
  alice: alice
  bob: bob
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("overwrites locally referenced fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( bar ))
  alice: overwritten
  other: other
bar:
  alice: alice
  bob: bob
`)

			resolved := parseYAML(`
---
foo:
  alice: overwritten
  bob: bob
  other: other
bar:
  alice: alice
  bob: bob
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("merges redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
bar: 42
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: merged!
`)

			resolved := parseYAML(`
---
foo: 
  bob: merged!
bar: 42
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("overwrites redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar: 42
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: added!
  bar: overwritten
`)

			resolved := parseYAML(`
---
foo: 
  bob: added!
  bar: overwritten
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("resolves overwritten redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar: 42
ref:
  bar: (( foo.bar ))
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: added!
  bar: overwritten
`)

			resolved := parseYAML(`
---
foo: 
  bob: added!
  bar: overwritten
ref:
  bar: overwritten
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("deep overwrites redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar:
    alice: alice
    bob: bob
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: added!
  bar:
    alice: overwritten
`)

			resolved := parseYAML(`
---
foo: 
  bar:
    alice: overwritten
    bob: bob
  bob: added!
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("propagates redirection to subsequent merges", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar:
    <<: (( merge ))
    alice: alice
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bar:
    alice: overwritten
    bob: added!
`)

			resolved := parseYAML(`
---
foo: 
  bar:
    alice: overwritten
    bob: added!
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	// replace whole structure instead of deep override
	Describe("replacing nodes from stubs", func() {
		It("does nothing for no direct match", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge replace || nil ))
  bar: 42
`)

			resolved := parseYAML(`
---
foo: 
  bar: 42
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("copies the node", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge replace ))
  bar: 42
`)

			stub := parseYAML(`
---
foo: 
  blah: replaced!
`)

			resolved := parseYAML(`
---
foo: 
  blah: replaced!
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("does not follow through maps in lists by name", func() {
			source := parseYAML(`
---
foo:
- <<: (( merge replace ))
- name: x
  value: v
`)

			stub := parseYAML(`
---
foo:
- name: y
  value: right
- name: z
  value: right
`)

			resolved := parseYAML(`
---
foo:
- name: y
  value: right
- name: z
  value: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("doesn't hamper field value merge", func() {
			source := parseYAML(`
---
foo:
  bar: (( merge replace ))
`)

			stub := parseYAML(`
---
foo:
  bar:
    value: right
`)

			resolved := parseYAML(`
---
foo:
  bar:
    value: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("doesn't hamper list value merge", func() {
			source := parseYAML(`
---
foo:
  bar: (( merge replace ))
`)

			stub := parseYAML(`
---
foo:
  bar:
    - alice
    - bob
`)

			resolved := parseYAML(`
---
foo:
  bar:
    - alice
    - bob
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("replacing map with redirection", func() {
		It("merges with redirected map, but not with original path", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge replace bar ))
  bar:
    alice: alice
    bob: bob
`)

			stub := parseYAML(`
---
foo:
  alice: not merged
bar:
  alice: merged
  bob: merged
`)

			resolved := parseYAML(`
---
foo:
  alice: merged
  bob: merged
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("replacing list with redirection", func() {
		It("merges with redirected map, but not with original path", func() {
			source := parseYAML(`
---
foo: 
  - <<: (( merge replace bar ))
  - bar:
      alice: alice
      bob: bob
`)

			stub := parseYAML(`
---
foo:
  - not
  - merged
bar:
  - alice: merged
  - bob: merged
`)

			resolved := parseYAML(`
---
foo:
  - alice: merged
  - bob: merged
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("resolves references to merges with redirected map", func() {
			source := parseYAML(`
---
foo:
  - <<: (( merge replace bar ))
  - bar:
      alice: alice
      bob: bob
ref: (( foo.[0].alice ))
`)

			stub := parseYAML(`
---
foo:
  - not
  - merged
bar:
  - alice: merged
  - bob: merged
`)

			resolved := parseYAML(`
---
foo:
  - alice: merged
  - bob: merged
ref: merged
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("merging field value", func() {
		It("merges with redirected map, but not with original path", func() {
			source := parseYAML(`
---
foo: (( merge bar ))
`)

			stub := parseYAML(`
---
foo:
  alice: not merged
bar:
  alice: alice
  bob: bob
`)

			resolved := parseYAML(`
---
foo:
  alice: alice
  bob: bob
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges with nothing", func() {
			source := parseYAML(`
---
foo: (( merge nothing || "default" ))
`)

			stub := parseYAML(`
---
foo:
  alice: not merged
`)

			resolved := parseYAML(`
---
foo: default
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("does not override merged values", func() {
			source := parseYAML(`
---
foo: (( (|x|->sum[x|{}|s,k,v|->s { k=v.value }])(merge data.foo) ))
`)

			stub := parseYAML(`
---
data:
  foo:
    alice:
      value: 24
`)

			resolved := parseYAML(`
---
foo:
  alice: 24
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("automatic resource pool sizes", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
resource_pools:
  some_pool:
    size: (( auto ))

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
- name: some_other_job
  resource_pool: some_pool
  instances: 3
- name: yet_another_job
  resource_pool: some_other_pool
  instances: 5
`)

			resolved := parseYAML(`
---
resource_pools:
  some_pool:
    size: 5

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
- name: some_other_job
  resource_pool: some_pool
  instances: 3
- name: yet_another_job
  resource_pool: some_other_pool
  instances: 5
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("static ip population", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
networks:
  some_network:
    type: manual
    subnets:
      - range: 10.10.16.0/20
        name: default_unused
        reserved:
          - 10.10.16.2 - 10.10.16.9
          - 10.10.16.255 - 10.10.16.255
        static:
          - 10.10.16.10 - 10.10.16.254
        gateway: 10.10.16.1
        dns:
          - 10.10.0.2

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
  networks:
  - name: some_network
    static_ips: (( static_ips(0, 4) ))
`)

			resolved := parseYAML(`
---
networks:
  some_network:
    type: manual
    subnets:
      - range: 10.10.16.0/20
        name: default_unused
        reserved:
          - 10.10.16.2 - 10.10.16.9
          - 10.10.16.255 - 10.10.16.255
        static:
          - 10.10.16.10 - 10.10.16.254
        gateway: 10.10.16.1
        dns:
          - 10.10.0.2

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
  networks:
  - name: some_network
    static_ips:
    - 10.10.16.10
    - 10.10.16.14
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates the node with list arguments", func() {
			source := parseYAML(`
---
networks:
  some_network:
    type: manual
    subnets:
      - range: 10.10.16.0/20
        name: default_unused
        reserved:
          - 10.10.16.2 - 10.10.16.9
          - 10.10.16.255 - 10.10.16.255
        static:
          - 10.10.16.10 - 10.10.16.254
        gateway: 10.10.16.1
        dns:
          - 10.10.0.2

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 3
  networks:
  - name: some_network
    static_ips: (( static_ips(0, [[4..1]]) ))
`)

			resolved := parseYAML(`
---
networks:
  some_network:
    type: manual
    subnets:
      - range: 10.10.16.0/20
        name: default_unused
        reserved:
          - 10.10.16.2 - 10.10.16.9
          - 10.10.16.255 - 10.10.16.255
        static:
          - 10.10.16.10 - 10.10.16.254
        gateway: 10.10.16.1
        dns:
          - 10.10.0.2

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 3
  networks:
  - name: some_network
    static_ips:
    - 10.10.16.10
    - 10.10.16.14
    - 10.10.16.13
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates the node with indirection combined with default", func() {
			source := parseYAML(`
---
meta:
  net: "10.10"

networks:
  some_network:
    type: manual
    subnets:
      - range: (( meta.net ".16.0/20" ))
        name: default_unused
        reserved:
          - (( meta.net ".16.2 - " meta.net ".16.9" ))
          - (( meta.net ".16.255 - " meta.net ".16.255" ))
        static:
          - (( meta.net ".16.10 - " meta.net ".16.254" ))
        gateway: (( meta.net ".16.1" ))
        dns:
          - (( meta.net ".0.2" ))

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
  networks:
  - name: some_network
    static_ips: (( static_ips(0, 4) || nil ))
`)

			resolved := parseYAML(`
---
meta:
  net: "10.10"

networks:
  some_network:
    type: manual
    subnets:
      - range: 10.10.16.0/20
        name: default_unused
        reserved:
          - 10.10.16.2 - 10.10.16.9
          - 10.10.16.255 - 10.10.16.255
        static:
          - 10.10.16.10 - 10.10.16.254
        gateway: 10.10.16.1
        dns:
          - 10.10.0.2

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
  networks:
  - name: some_network
    static_ips:
    - 10.10.16.10
    - 10.10.16.14
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("ipset population", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
ranges:
  - 10.0.0.0-10.0.0.255
  - 10.0.2.0/24
ipset: (( ipset(ranges,3,10,12,14,16,18) ))
`)
			resolved := parseYAML(`
---
ranges:
  - 10.0.0.0-10.0.0.255
  - 10.0.2.0/24
ipset:
  - 10.0.0.10
  - 10.0.0.12
  - 10.0.0.14
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates the second range", func() {
			source := parseYAML(`
---
ranges:
  - 10.0.0.0-10.0.0.255
  - 10.0.2.0/24
ipset: (( ipset(ranges,3,[257..270]) ))
`)
			resolved := parseYAML(`
---
ranges:
  - 10.0.0.0-10.0.0.255
  - 10.0.2.0/24
ipset:
  - 10.0.2.1
  - 10.0.2.2
  - 10.0.2.3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("support no indirection", func() {
			source := parseYAML(`
---
ranges:
  - 10.0.0.0-10.0.0.255
  - 10.0.2.0/24
ipset: (( ipset(ranges,3) ))
`)
			resolved := parseYAML(`
---
ranges:
  - 10.0.0.0-10.0.0.255
  - 10.0.2.0/24
ipset:
  - 10.0.0.0
  - 10.0.0.1
  - 10.0.0.2
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("map splicing", func() {
		It("merges one map over another", func() {
			source := parseYAML(`
---
properties:
  something:
    foo:
      <<: (( merge ))
      key: a
      val: b
      some:
        s: stuff
        d: blah
`)

			stub := parseYAML(`
---
properties:
  something:
    foo:
      val: c
      some:
        go: home
`)

			resolved := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: c
      some:
        s: stuff
        d: blah
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges one map over another and resolves inbound references", func() {
			source := parseYAML(`
---
properties:
  something:
    foo:
      <<: (( merge ))
      key: a
      val: b
      some:
        s: stuff
        d: blah
  refkey: (( properties.something.foo.key ))
  refval: (( properties.something.foo.val ))
`)

			stub := parseYAML(`
---
properties:
  something:
    foo:
      val: c
      some:
        go: home
`)

			resolved := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: c
      some:
        s: stuff
        d: blah
  refkey: a
  refval: c
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("list splicing", func() {
		It("merges one list into another", func() {
			source := parseYAML(`
---
properties:
  something:
    - a
    - <<: (( list ))
    - b
  list:
    - c
    - d
`)

			resolved := parseYAML(`
---
properties:
  something:
    - a
    - c
    - d
    - b
  list:
    - c
    - d
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("merges merged map into list", func() {
			source := parseYAML(`
---
properties:
  something:
    - a
    - <<: (( map ))
      foo: bar
    - b
  map:
    alice: bob
`)

			resolved := parseYAML(`
---
properties:
  something:
    - a
    - alice: bob
      foo: bar
    - b
  map:
    alice: bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("merges stub", func() {
			source := parseYAML(`
---
properties:
  something:
    - a
    - <<: (( merge ))
    - b
`)

			stub := parseYAML(`
---
properties:
  something:
    - c
    - d
`)

			resolved := parseYAML(`
---
properties:
  something:
    - a
    - c
    - d
    - b
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("redirects stub", func() {
			source := parseYAML(`
---
properties:
  something:
    - a
    - <<: (( merge alt ))
    - b
`)

			stub := parseYAML(`
---
properties:
  something:
    - e
    - f
alt:
  - c
  - d
`)

			resolved := parseYAML(`
---
properties:
  something:
    - a
    - c
    - d
    - b
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		Context("when names match", func() {
			It("replaces existing entries with matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge ))
    - name: b
      value: 2
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
`)

				Expect(source).To(FlowAs(resolved, stub))
			})

			It("resolves existing entries replaced with matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge ))
    - name: b
      value: 2
ref: (( properties.something.[0].value ))
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
ref: 10
`)

				Expect(source).To(FlowAs(resolved, stub))
			})

			It("replaces existing entries with redirected matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge alt.something ))
    - name: b
      value: 2
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: c
      value: 300

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
`)

				Expect(source).To(FlowAs(resolved, stub))
			})

			It("resolves existing entries replaced with redirected matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge alt.something ))
    - name: b
      value: 2
ref: (( properties.something.a.value ))
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: c
      value: 300

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
ref: 10
`)

				Expect(source).To(FlowAs(resolved, stub))
			})
		})

		It("uses redirected matching names, but not original path", func() {
			source := parseYAML(`
---
properties:
  something: (( merge alt.something ))
`)

			stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: b
      value: 200

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

			resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("avoids override by original path, which occured by traditional redirection", func() {
			source := parseYAML(`
---
alt:
  something: (( merge ))

properties:
  something: (( prefer alt.something ))
`)

			stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: b
      value: 200

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

			resolved := parseYAML(`
---
alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30

properties:
  something:
    - name: a
      value: 100
    - name: c
      value: 30
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges appropriate list entry for lists with explicitly merged maps", func() {
			source := parseYAML(`
---
list:
  - name: alice
    married: bob
    <<: (( merge ))
`)
			stub := parseYAML(`
---
list:
  - name: mary
    married: no
  - name: alice
    married: peter
    age: 25
`)
			resolved := parseYAML(`
---
list:
  - name: alice
    married: peter
    age: 25
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges appropriate list entry for lists with key expressions", func() {
			source := parseYAML(`
---
name: foobar
list:
  - name: (( .name ))
    value: alice
`)
			stub := parseYAML(`
---
list:
  - name: foo
    value: peter
  - name: foobar
    value: bob
`)
			resolved := parseYAML(`
---
name: foobar
list:
  - name: foobar
    value: bob
`)
			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("for list expressions", func() {
		It("evaluates lists", func() {
			source := parseYAML(`
---
foo: (( [ "a", "b" ] ))
`)
			resolved := parseYAML(`
---
foo:
  - a
  - b
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates lists with references", func() {
			source := parseYAML(`
---
a: alice
b: bob
foo: (( [ a, b ] || "failed" ))
`)
			resolved := parseYAML(`
---
a: alice
b: bob
foo:
  - alice
  - bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates for lists with deep references", func() {
			source := parseYAML(`
---
a: alice
b: bob
c: (( b ))
foo: (( [ a, c ] || "failed" ))
`)
			resolved := parseYAML(`
---
a: alice
b: bob
c: bob
foo:
  - alice
  - bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("failes for lists with unresolved references", func() {
			source := parseYAML(`
---
a: alice
foo: (( [ a, b ] || "failed" ))
`)
			resolved := parseYAML(`
---
a: alice
foo: failed
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("for arithmetic expressions", func() {
		Context("addition", func() {
			It("evaluates addition", func() {
				source := parseYAML(`
---
foo: (( 1 + 2 + 3 ))
`)
				resolved := parseYAML(`
---
foo: 6
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 1
b: 2
c: (( b ))
foo: (( a + c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 1
b: 2
c: 2
foo: 3
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 1
b: 2
foo: (( a + c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 1
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("subtraction", func() {
			It("evaluates subtraction", func() {
				source := parseYAML(`
---
foo: (( 6 - 3 - 2 ))
`)
				resolved := parseYAML(`
---
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 3
b: 2
c: (( b ))
foo: (( a - c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 3
b: 2
c: 2
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 3
b: 2
foo: (( a - c || "failed" ))
`)

				resolved := parseYAML(`
---
a: 3
b: 2
foo: failed
`)

				Expect(source).To(FlowAs(resolved))
			})

			It("subtracts IPs", func() {
				source := parseYAML(`
---
foo: (( 10.0.0.1 - 10.0.1.0 ))
`)
				resolved := parseYAML(`
---
foo: -255
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("subtracts IP and integer", func() {
				source := parseYAML(`
---
foo: (( 10.0.0.1 - 2 ))
`)
				resolved := parseYAML(`
---
foo: "9.255.255.255"
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("multiplication", func() {
			It("evaluates multiplication", func() {
				source := parseYAML(`
---
foo: (( 6 * 2 * 3 ))
`)
				resolved := parseYAML(`
---
foo: 36
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 6
b: 2
c: (( b ))
foo: (( a * c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
c: 2
foo: 12
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 6
b: 2
foo: (( a * c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("division", func() {
			It("evaluates division", func() {
				source := parseYAML(`
---
foo: (( 6 / 2 / 3 ))
`)
				resolved := parseYAML(`
---
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("division by zero fails", func() {
				source := parseYAML(`
---
foo: (( 6 / 0 || "failed" ))
`)
				resolved := parseYAML(`
---
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 6
b: 2
c: (( b ))
foo: (( a / c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
c: 2
foo: 3
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 6
b: 2
foo: (( a / c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("modulo", func() {
			It("evaluates modulo", func() {
				source := parseYAML(`
---
foo: (( 13 % ( 2 * 3 )))
`)
				resolved := parseYAML(`
---
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("modulo by zero fails", func() {
				source := parseYAML(`
---
foo: (( 13 % ( 2 - 2 ) || "failed" ))
`)
				resolved := parseYAML(`
---
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 7
b: 2
c: (( b ))
foo: (( a % c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 7
b: 2
c: 2
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 7
b: 2
foo: (( a / c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 7
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("mixed levels", func() {
			It("evaluates multiplication first", func() {
				source := parseYAML(`
---
foo: (( 6 + 2 * 3 ))
`)
				resolved := parseYAML(`
---
foo: 12
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates addition last", func() {
				source := parseYAML(`
---
foo: (( 6 * 2 + 3 ))
`)
				resolved := parseYAML(`
---
foo: 15
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		It("evaluates arithmetic before concatenation", func() {
			source := parseYAML(`
---
foo: (( "prefix" 6 * 2 + 3 "suffix" ))
`)

			resolved := parseYAML(`
---
foo: prefix15suffix
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("concatenates arithmetic values as string", func() {
			source := parseYAML(`
---
foo: ((  6 * 2 + 3 15 ))
`)

			resolved := parseYAML(`
---
foo: "1515"
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("for logical expressions", func() {
		It("evaluates not", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( !foo ))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates and", func() {
			source := parseYAML(`
---
foo: (( 0 ))
bar: (( !foo -and true))
`)
			resolved := parseYAML(`
---
foo: 0
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates or", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( !foo -or true))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates <=", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo <= 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates <", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo < 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates >=", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo >= 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates >", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo > 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates ==", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo == 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates == of lists", func() {
			source := parseYAML(`
---
foo: 
  - alice
  - bob
bar: (( foo == [ "alice","bob" ] ))
`)
			resolved := parseYAML(`
---
foo: 
  - alice
  - bob
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates == of lists to false", func() {
			source := parseYAML(`
---
foo: 
  - alice
  - bob
bar: (( foo == [ "alice","paul" ] ))
`)
			resolved := parseYAML(`
---
foo: 
  - alice
  - bob
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates == of maps", func() {
			source := parseYAML(`
---
foo: 
  a: 1
  b: 2

comp:
  a: 1
  b: 2

bar: (( foo == comp ))
`)
			resolved := parseYAML(`
---
foo: 
  a: 1
  b: 2

comp:
  a: 1
  b: 2

bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates !=", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo != 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when concatenating a map", func() {
		Context("with other maps", func() {
			It("yields a joined map", func() {
				source := parseYAML(`
---
map1:
  alice: a
  bob: b
map2:
  bob: b2
  peter: p

foo: (( map1 map2 ))
`)

				resolved := parseYAML(`
---
map1:
  alice: a
  bob: b
map2:
  bob: b2
  peter: p
foo:
  alice: a
  bob: b2
  peter: p
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("handles empty map constant", func() {
				source := parseYAML(`
---
map1:
  alice: a
  bob: b

foo: (( {} map1 ))
`)

				resolved := parseYAML(`
---
map1:
  alice: a
  bob: b
foo:
  alice: a
  bob: b
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when concatenating a list", func() {
		Context("with incremental expression resolution", func() {
			It("evaluates in case of successfully completed operand resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
c: (( b ))
foo: (( a "+" c || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
c: bob
foo: alice+bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("fails only after failed final resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
foo: (( a "+" c || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with other lists", func() {
			It("yields a joined list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] [ 2 * 3 ] [4,5,6] ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,6,4,5,6]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with an integer", func() {
			It("appends the value to the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] 4 5 ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,4,5]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a string", func() {
			It("appends the value to the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] "foo" "bar" ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,"foo","bar"]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a map", func() {
			It("appends the map to the list", func() {
				source := parseYAML(`
---
bar:
  alice: and bob
foo: (( [1,2,3] bar ))
`)

				resolved := parseYAML(`
---
bar:
  alice: and bob
foo: [1,2,3,{"alice": "and bob"}]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a nested string concatenation", func() {
			It("appends the value to the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] ("foo" "bar") ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,"foobar"]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a nested list concatenation", func() {
			It("joins the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] ([] "bar") ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,"bar"]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when joining", func() {
		It("joins single value", func() {
			source := parseYAML(`
---
foo: (( join( ", ", "alice") ))
`)
			resolved := parseYAML(`
---
foo: alice
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("joins strings and integers", func() {
			source := parseYAML(`
---
foo: (( join( ", ", "alice", "bob", 5) ))
`)
			resolved := parseYAML(`
---
foo: alice, bob, 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("joins elements from lists", func() {
			source := parseYAML(`
---
list:
  - alice
  - bob
foo: (( join( ", ", list, 5) ))
`)
			resolved := parseYAML(`
---
list:
  - alice
  - bob
foo: alice, bob, 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("joins elements from inline list", func() {
			source := parseYAML(`
---
b: bob
foo: (( join( ", ", [ "alice", b ] ) ))
`)
			resolved := parseYAML(`
---
b: bob
foo: alice, bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		Context("with incremental expression resolution", func() {
			It("evaluates in case of successfully completed operand resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
c: (( b ))
foo: (( join( ", ", a, c) || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
c: bob
foo: alice, bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates in case of successfully completed list operand resolution", func() {
				source := parseYAML(`
---
list:
  - alice
  - (( c ))
b: bob
c: (( b ))
foo: (( join( ", ", list) || "failed" ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
b: bob
c: bob
foo: alice, bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates in case of successfully completed list expression resolution", func() {
				source := parseYAML(`
---
b: bob
c: (( b ))
foo: (( join( ", ", [ "alice", c ] ) || "failed" ))
`)
				resolved := parseYAML(`
---
b: bob
c: bob
foo: alice, bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("fails only after failed final resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
foo: (( join( ", ", a, c) || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("fails only after failed final list resolution", func() {
				source := parseYAML(`
---
foo: (( join( ", ", [ "alice", c ] ) || "failed" ))
`)
				resolved := parseYAML(`
---
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when splitting", func() {
		It("splits single value", func() {
			source := parseYAML(`
---
foo: (( split( ",", "alice") ))
`)
			resolved := parseYAML(`
---
foo:
 - alice
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("splits multiple values", func() {
			source := parseYAML(`
---
foo: (( split( ",", "alice,bob") ))
`)
			resolved := parseYAML(`
---
foo:
 - alice
 - bob
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when trimming", func() {
		It("trims strings", func() {
			source := parseYAML(`
---
foo: (( trim( "  alice ") ))
`)
			resolved := parseYAML(`
---
foo: alice
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("trims dedicated characters", func() {
			source := parseYAML(`
---
foo: (( trim( "alice", "ae") ))
`)
			resolved := parseYAML(`
---
foo: lic
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("trims lists", func() {
			source := parseYAML(`
---
foo: (( trim( split(",","alice, bob ")) ))
`)
			resolved := parseYAML(`
---
foo:
  - alice
  - bob
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling element", func() {
		It("extracts fields from maps", func() {
			source := parseYAML(`
---
map:
  alice: 24
  bob: 25

elem: (( element(map,"bob") ))
`)
			resolved := parseYAML(`
---
map:
  alice: 24
  bob: 25

elem: 25
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("extracts dotted fields from maps", func() {
			source := parseYAML(`
---
map:
  foo.bar: 25

elem: (( element(map,"foo.bar") ))
`)
			resolved := parseYAML(`
---
map:
  foo.bar: 25

elem: 25
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("failes for invalid mapkeys", func() {
			source := parseYAML(`
---
map:
  foo.bar: 25

elem: (( element(map,"foo") || "failed" ))
`)
			resolved := parseYAML(`
---
map:
  foo.bar: 25

elem: failed
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("extracts entries from lists", func() {
			source := parseYAML(`
---
list:
  - alice: 24
  - bob: 25

elem: (( element(list,1) ))
`)
			resolved := parseYAML(`
---
list:
  - alice: 24
  - bob: 25

elem:
  bob: 25
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("fails for invalid list index", func() {
			source := parseYAML(`
---
list:
  - alice: 24
  - bob: 25

elem: (( element(list,2) || "failed" ))
`)
			resolved := parseYAML(`
---
list:
  - alice: 24
  - bob: 25

elem: failed
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling stub", func() {
		It("handles reference arg", func() {
			source := parseYAML(`
---
age: (( stub(data.alice) ))
`)
			stub := parseYAML(`
---
data:
  alice: "24"
`)

			resolved := parseYAML(`
---
age: "24"
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("handles string arg", func() {
			source := parseYAML(`
---
age: (( stub("data.alice") ))
`)
			stub := parseYAML(`
---
data:
  alice: 24
`)

			resolved := parseYAML(`
---
age: 24
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("fails on missing stub", func() {
			source := parseYAML(`
---
age: (( stub("data.alice") || "failed" ))
`)

			resolved := parseYAML(`
---
age: failed
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("refers to local path if no arg is given", func() {
			source := parseYAML(`
---
age: (( stub() ))
`)

			stub := parseYAML(`
---
age: 20
`)
			resolved := parseYAML(`
---
age: 20
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("does not prevent merging", func() {
			source := parseYAML(`
---

val: (( prefer stub(data) ))
`)

			stub := parseYAML(`
---
data:
  alice: 24
  bob: 25
val:
  bob: 100
`)
			resolved := parseYAML(`
---
val:
  alice: 24
  bob: 100
`)
			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("when calling uniq", func() {
		It("omits duplicates", func() {
			source := parseYAML(`
---
list:
- a
- b
- a
- c
- a
- b
- 0
- "0"
uniq: (( uniq(list) ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- a
- c
- a
- b
- 0
- "0"
uniq:
- a
- b
- c
- 0
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling compact", func() {
		It("omits empty entries", func() {
			source := parseYAML(`
---
list:
- a
- ~
- ""
- {}
- []
- b

compact: (( compact(list) ))
`)
			resolved := parseYAML(`
---
list:
- a
- ~
- ""
- {}
- []
- b
compact:
- a
- b
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling uniq", func() {
		It("omits duplicates", func() {
			source := parseYAML(`
---
list:
- a
- b
- a
- c
- a
- b
- 0
- "0"
uniq: (( uniq(list) ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- a
- c
- a
- b
- 0
- "0"
uniq:
- a
- b
- c
- 0
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling contains", func() {
		It("finds ints", func() {
			source := parseYAML(`
---
list:
- a
- b
- 0
- c
contains: (( contains(list, "0") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- 0
- c

contains: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("finds string", func() {
			source := parseYAML(`
---
list:
- a
- b
- "0"
- c
contains: (( contains(list, "0") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- "0"
- c

contains: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("works for no match", func() {
			source := parseYAML(`
---
list:
- a
- b
- 0
- c
contains: (( contains(list, "d") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- 0
- c

contains: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string contains", func() {
			source := parseYAML(`
---
contains: (( contains("1234567890123", "0") ))
`)
			resolved := parseYAML(`
---
contains: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string contains with int", func() {
			source := parseYAML(`
---
contains: (( contains("1234567890123", 0) ))
`)
			resolved := parseYAML(`
---
contains: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string contains and fails", func() {
			source := parseYAML(`
---
contains: (( contains("1234567890123", "a") ))
`)
			resolved := parseYAML(`
---
contains: false
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling index", func() {
		It("finds ints", func() {
			source := parseYAML(`
---
list:
- a
- b
- 0
- c
- 0
index: (( index(list, "0") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- 0
- c
- 0

index: 2
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("finds string", func() {
			source := parseYAML(`
---
list:
- a
- b
- "0"
- c
- "0"
index: (( index(list, "0") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- "0"
- c
- "0"
index: 2
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("works for no match", func() {
			source := parseYAML(`
---
list:
- a
- b
- 0
- c
index: (( index(list, "d") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- 0
- c

index: -1
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string index", func() {
			source := parseYAML(`
---
index: (( index("12345678901230", "0") ))
`)
			resolved := parseYAML(`
---
index: 9
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string index with int", func() {
			source := parseYAML(`
---
index: (( index("12345678901230", 0) ))
`)
			resolved := parseYAML(`
---
index: 9
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string index and fails", func() {
			source := parseYAML(`
---
index: (( index("1234567890123", "a") ))
`)
			resolved := parseYAML(`
---
index: -1
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling lastindex", func() {
		It("finds ints", func() {
			source := parseYAML(`
---
list:
- a
- b
- 0
- c
- 0
index: (( lastindex(list, "0") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- 0
- c
- 0

index: 4
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("finds string", func() {
			source := parseYAML(`
---
list:
- a
- b
- "0"
- c
- "0"
index: (( lastindex(list, "0") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- "0"
- c
- "0"
index: 4
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("works for no match", func() {
			source := parseYAML(`
---
list:
- a
- b
- 0
- c
index: (( lastindex(list, "d") ))
`)
			resolved := parseYAML(`
---
list:
- a
- b
- 0
- c

index: -1
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string index", func() {
			source := parseYAML(`
---
index: (( lastindex("12345678901230", "0") ))
`)
			resolved := parseYAML(`
---
index: 13
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string index with int", func() {
			source := parseYAML(`
---
index: (( lastindex("12345678901230", 0) ))
`)
			resolved := parseYAML(`
---
index: 13
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles string index and fails", func() {
			source := parseYAML(`
---
index: (( lastindex("1234567890123", "a") ))
`)
			resolved := parseYAML(`
---
index: -1
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when replacing", func() {
		It("replaces unlimited", func() {
			source := parseYAML(`
---
result: (( replace("foobar","o", "u") ))
`)
			resolved := parseYAML(`
---
result: fuubar
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("replaces limited", func() {
			source := parseYAML(`
---
result: (( replace("foobar","o", "u", 1) ))
`)
			resolved := parseYAML(`
---
result: fuobar
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when matching regexps", func() {
		It("matches strings", func() {
			source := parseYAML(`
---
match: (( match("^f.*r$","foobar") ))
`)
			resolved := parseYAML(`
---
match:
  - foobar
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("matches non-matching strings", func() {
			source := parseYAML(`
---
match: (( match("^f.*r$","foobal") ))
`)
			resolved := parseYAML(`
---
match: []
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("matches sub expressions strings", func() {
			source := parseYAML(`
---
match: (( match("^(f.*)(b.*)$","foobar") ))
`)
			resolved := parseYAML(`
---
match:
  - foobar
  - foo
  - bar
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("calling length", func() {
		It("calculates string length", func() {
			source := parseYAML(`
---
foo: (( length( "alice") ))
`)
			resolved := parseYAML(`
---
foo: 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("calculates list length", func() {
			source := parseYAML(`
---
foo: (( length( ["alice","bob"]) ))
`)
			resolved := parseYAML(`
---
foo: 2
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("calculates map length", func() {
			source := parseYAML(`
---
map:
  alice: 25
  bob: 24

foo: (( length( map) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
  bob: 24
foo: 2
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when reevaluating an expression", func() {
		It("resolves indirect fields", func() {
			source := parseYAML(`
---
alice:
  bob: married

foo: alice
bar: bob

status: (( eval( foo "." bar ) ))
`)
			resolved := parseYAML(`
---
alice:
  bob: married

foo: alice
bar: bob

status: married
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("defaults evaluation errors", func() {
			source := parseYAML(`
---
alice:
  bob: married

foo: alice

status: (( eval( foo "." bar ) || "failed" ))
`)
			resolved := parseYAML(`
---
alice:
  bob: married

foo: alice

status: failed
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when resing from the environment", func() {
		os.Setenv("TEST1", "alice")
		os.Setenv("TEST2", "bob")
		dynaml.ReloadEnv()

		It("resolves a single variable", func() {
			source := parseYAML(`
---
alice: (( env("TEST1") ))
`)
			resolved := parseYAML(`
---
alice: alice
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("defaults a non-existing single variable", func() {
			source := parseYAML(`
---
alice: (( env("TEST3") || "default" ))
`)
			resolved := parseYAML(`
---
alice: default
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("resolves a two variables to a map", func() {
			source := parseYAML(`
---
env: (( env("TEST1","TEST2") ))
`)
			resolved := parseYAML(`
---
env:
  TEST1: alice
  TEST2: bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("resolves a list to a map", func() {
			source := parseYAML(`
---
list:
  - TEST1
  - TEST2
env: (( env(list) ))
`)
			resolved := parseYAML(`
---
list:
  - TEST1
  - TEST2
env:
  TEST1: alice
  TEST2: bob
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when formatting a string", func() {
		It("formats strings and integers", func() {
			source := parseYAML(`
---
int: 5
str: string
msg: (( format("%s %d", str, int) ))
`)
			resolved := parseYAML(`
---
int: 5
str: string
msg: string 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("formats maps", func() {
			source := parseYAML(`
---
map:
  alice: 25
msg: (( format("%s", map) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
msg: |+
  alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("formats lists", func() {
			source := parseYAML(`
---
list:
  - alice
  - bob
msg: (( format("%s", list) ))
`)
			resolved := parseYAML(`
---
list:
  - alice
  - bob
msg: |+
  - alice
  - bob
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when transforming a list to a map", func() {
		It("handles standard key", func() {
			source := parseYAML(`
---
list:
  - name: alice
    age: 24
  - name: bob
    age: 30

map: (( list_to_map(list) ))

`)
			resolved := parseYAML(`
---
list:
  - name: alice
    age: 24
  - name: bob
    age: 30

map:
  alice:
    age: 24
  bob:
    age: 30
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles inline key", func() {
			source := parseYAML(`
---
list:
  - key:key: alice
    age: 24
  - key: bob
    age: 30

map: (( list_to_map(list) ))

`)
			resolved := parseYAML(`
---
list:
  - key: alice
    age: 24
  - key: bob
    age: 30

map:
  alice:
    age: 24
  bob:
    age: 30
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles explicit key", func() {
			source := parseYAML(`
---
list:
  - key: alice
    age: 24
  - key: bob
    age: 30

map: (( list_to_map(list,"key") ))

`)
			resolved := parseYAML(`
---
list:
  - key: alice
    age: 24
  - key: bob
    age: 30

map:
  alice:
    age: 24
  bob:
    age: 30
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when making a map", func() {
		It("handles entries given by a list", func() {
			source := parseYAML(`
---
list:
  - key: alice
    value: 24
  - key: bob 
    value: 25
  - key: 5
    value: 25

map: (( makemap(list) ))

`)
			resolved := parseYAML(`
---
list:
  - key: alice
    value: 24
  - key: bob 
    value: 25
  - key: 5
    value: 25

map:
  "5": 25
  alice: 24
  bob: 25
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles a entries given by arguments", func() {
			source := parseYAML(`
---
map: (( makemap("peter", 23, "paul", 22) ))
`)
			resolved := parseYAML(`
---
map:
  paul: 22
  peter: 23
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles map literals", func() {
			source := parseYAML(`
---
peter:
  name: peter
  age: 23
paul:
  name: paul
  age: 22
map: (( { peter.name=peter.age, paul.name=paul.age } ))
`)
			resolved := parseYAML(`
---
peter:
  name: peter
  age: 23
paul:
  name: paul
  age: 22
map:
  paul: 22
  peter: 23
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles nested map literals", func() {
			source := parseYAML(`
---
name: peter
age: 23
map: (( { "alice" = {}, name = age } ))
`)
			resolved := parseYAML(`
---
name: peter
age: 23
map:
  alice: {}
  peter: 23
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when doing a mapping", func() {
		Context("for a list", func() {
			It("maps simple expression", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
mapped: (( map[list|x|->x] ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
mapped:
  - alice
  - bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("filters nil values", func() {
				source := parseYAML(`
---
list:
  - alice
  - ~
mapped: (( map[list|x|->x] ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - ~
mapped:
  - alice
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps index expression", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
mapped: (( map[list|y,x|->y x] ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
mapped:
  - 0alice
  - 1bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps concatenation expression", func() {
				source := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped: (( map[list|x|->x ":" port] ))
`)
				resolved := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped:
  - alice:4711
  - bob:4711
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps reference expression", func() {
				source := parseYAML(`
---
list:
  - name: alice
    age: 25
  - name: bob
    age: 24
names: (( map[list|x|->x.name] ))
`)
				resolved := parseYAML(`
---
list:
  - name: alice
    age: 25
  - name: bob
    age: 24
names:
  - alice
  - bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps concatenation expression without failure", func() {
				source := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped: (( map[list|x|->x ":" port] || "failed" ))
`)
				resolved := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped:
  - alice:4711
  - bob:4711
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps concatenation expression with failure", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
mapped: (( map[list|x|->x ":" port] || "failed" ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
mapped: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("works with nested expressions", func() {
				source := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
joined: (( join( ", ", map[list|x|->x ":" port] ) || "failed" ))
`)
				resolved := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
joined: alice:4711, bob:4711
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("works with nested failing expressions", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
joined: (( join( ", ", map[list|x|->x ":" port] ) || "failed" ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
joined: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps with referenced expression", func() {
				source := parseYAML(`
---
map: '|x|->x'
list:
  - alice
  - bob
mapped: (( map[list|lambda map] ))
`)
				resolved := parseYAML(`
---
map: '|x|->x'
list:
  - alice
  - bob
mapped:
  - alice
  - bob
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("for a map", func() {
			It("maps simple expression", func() {
				source := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped: (( map[map|x|->x] ))
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped:
  - 25
  - 24
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("filters empty expression", func() {
				source := parseYAML(`
---
map:
  alice: 25
  bob: ~
mapped: (( map[map|x|->x] ))
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  bob: ~
mapped:
  - 25
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps key expression", func() {
				source := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped: (( map[map|y,x|->y x] ))
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped:
  - alice25
  - bob24
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when doing a sum", func() {
		Context("for a list", func() {
			It("sums simple expression", func() {
				source := parseYAML(`
---
list:
  - 1
  - 2
sum: (( sum[list|0|s,x|->s + x] ))
`)
				resolved := parseYAML(`
---
list:
  - 1
  - 2
sum: 3
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("sums provides index and value", func() {
				source := parseYAML(`
---
list:
  - 1
  - 2
  - 3
sum: (( sum[list|0|s,i,x|->s + i * x] ))
`)
				resolved := parseYAML(`
---
list:
  - 1
  - 2
  - 3
sum: 8
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("works with failing expressions", func() {
				source := parseYAML(`
---
list:
  - 1
  - 2
sum: (( sum[list|0|s,x|->s + x + y] || "failed" ))
`)
				resolved := parseYAML(`
---
list:
  - 1
  - 2
sum: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps with referenced expression", func() {
				source := parseYAML(`
---
map: "|s,x|->s + x"
list:
  - 1
  - 2
sum: (( sum[list|0|lambda map] ))
`)
				resolved := parseYAML(`
---
map: "|s,x|->s + x"
list:
  - 1
  - 2
sum: 3
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("for a map", func() {
			It("sums simple expression", func() {
				source := parseYAML(`
---
map:
  alice: 1
  bob: 2
sum: (( sum[map|0|s,x|->s + x] ))
`)
				resolved := parseYAML(`
---
map:
  alice: 1
  bob: 2
sum: 3
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("sums provides access to key", func() {
				source := parseYAML(`
---
factors:
  alice: 2
  bob: 3
map:
  alice: 1
  bob: 2
sum: (( sum[map|0|s,k,x|->s + eval("factors." k) * x] ))
`)
				resolved := parseYAML(`
---
factors:
  alice: 2
  bob: 3
map:
  alice: 1
  bob: 2
sum: 8
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("using templates", func() {
		Context("direct usage in list", func() {
			It("uses usage context", func() {
				source := parseYAML(`
---
templ: (( &template ( a + 1 ) ))
foo:
  a: 2
  bar: (( *templ ))
`)

				resolved, _ := Flow(parseYAML(`
---
templ: (( &template ( a + 1 ) ))
foo:
  a: 2
  bar: 3
`))
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("direct usage in list", func() {
			It("uses usage context", func() {
				source := parseYAML(`
---
verb: hates

foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))

use:
  verb: loves
  subst: (( *foo.bar ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))
use:
  subst:
    - loves alice
  verb: loves
verb: hates
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("uses usage context without falling back to default", func() {
				source := parseYAML(`
---
verb: hates

foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))

use:
  verb: loves
  subst: (( *foo.bar || "failed" ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))
use:
  subst:
    - loves alice
  verb: loves
verb: hates
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("handles independent usage", func() {
				source := parseYAML(`
---
verb: hates

foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))

use1:
  verb: loves
  subst: (( *foo.bar || "failed" ))
use2:
  verb: works
  subst: (( *foo.bar || "failed" ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))
use1:
  subst:
    - loves alice
  verb: loves
use2:
  subst:
    - works alice
  verb: works
verb: hates
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("defaults failures", func() {
				source := parseYAML(`
---
foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))

use:
  subst: (( *foo.bar || "failed" ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    - <<: (( &template ))
    - (( verb " alice" ))
use:
  subst: failed
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("defaults deep failures", func() {
				source := parseYAML(`
---
verbs: (( merge || nil ))
foo:
  bar:
    - <<: (( &template ))
    - (( verbs.verb " alice" ))

use:
  subst: (( *foo.bar || "failed" ))
`)
				resolved, _ := Flow(parseYAML(`
---
verbs: ~
foo:
  bar:
    - <<: (( &template ))
    - (( verbs.verb " alice" ))
use:
  subst: failed
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("merges list templates", func() {
				source := parseYAML(`
---
foo:
  bar:
  - <<: (( &template ))
  - (( verb " alice" ))

use:
  verb: loves
  subst:
  - a
  - <<: (( *foo.bar || "failed" ))
  - b
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
  - <<: (( &template ))
  - (( verb " alice" ))
use:
  verb: loves
  subst:
  - a
  - loves alice
  - b
`))
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("list template overriding", func() {
			It("overrides template substitution expression", func() {
				source := parseYAML(`
---
templ:
  - <<: (( &template ))
  - name: foo
    attr: 24
  - name: bar
    attr: 25

inst: (( *templ ))
`)
				stub := parseYAML(`
---
inst:
  - name: foo
    add: all
    attr: 34
  - name: alice
    attr: 35
`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  - <<: (( &template ))
  - name: foo
    attr: 24
  - name: bar
    attr: 25

inst:
  - name: foo
    add: all
    attr: 34
  - name: alice
    attr: 35
`))
				Expect(source).To(FlowAs(resolved, stub))
			})

			It("overrides template substitution", func() {
				source := parseYAML(`
---
templ:
  - <<: (( &template ))
  - name: foo
    attr: 24
  - name: bar
    attr: 25

inst: (( prefer *templ ))
`)
				stub := parseYAML(`
---
inst:
  - name: foo
    add: all
    attr: 34
  - name: alice
    attr: 35

`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  - <<: (( &template ))
  - name: foo
    attr: 24
  - name: bar
    attr: 25

inst:
  - name: foo
    attr: 34
  - name: bar
    attr: 25
`))
				Expect(source).To(FlowAs(resolved, stub))
			})

			It("inserts into template substitution", func() {
				source := parseYAML(`
---
templ:
  - <<: (( &template ))
  - name: foo
    <<: (( merge ))
    attr: 24
  - name: bar
    attr: 25

inst: (( prefer *templ ))
`)
				stub := parseYAML(`
---
inst:
  - name: foo
    add: all
    attr: 34
  - name: alice
    attr: 35l

`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  - <<: (( &template ))
  - name: foo
    <<: (( merge ))
    attr: 24
  - name: bar
    attr: 25

inst:
  - name: foo
    add: all
    attr: 34
  - name: bar
    attr: 25
`))
				Expect(source).To(FlowAs(resolved, stub))
			})

			It("supports extended marker", func() {
				source := parseYAML(`
---
templ:
  - <<: (( &template (merge) ))
  - name: foo
    attr: 24
  - name: bar
    attr: 25

inst: (( prefer *templ ))
`)
				stub := parseYAML(`
---
inst:
  - name: foo
    add: all
    attr: 34
  - name: alice
    attr: 35

`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  - <<: (( &template (merge) ))
  - name: foo
    attr: 24
  - name: bar
    attr: 25

inst:
  - name: alice
    attr: 35
  - name: foo
    attr: 34
  - name: bar
    attr: 25
`))
				Expect(source).To(FlowAs(resolved, stub))
			})
		})

		Context("direct usage for map", func() {
			It("uses usage context", func() {
				source := parseYAML(`
---
verb: hates

foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))


use:
  verb: loves
  subst: (( *foo.bar ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))
use:
  subst:
    alice: alice
    bob: loves alice
  verb: loves
verb: hates
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("uses usage context without falling back to default", func() {
				source := parseYAML(`
---
verb: hates

foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))


use:
  verb: loves
  subst: (( *foo.bar || "failed" ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))
use:
  subst:
    alice: alice
    bob: loves alice
  verb: loves
verb: hates
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("handles independent usage", func() {
				source := parseYAML(`
---
verb: hates

foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))


use1:
  verb: loves
  subst: (( *foo.bar ))
use2:
  verb: works with
  subst: (( *foo.bar ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))
use1:
  subst:
    alice: alice
    bob: loves alice
  verb: loves
use2:
  subst:
    alice: alice
    bob: works with alice
  verb: works with
verb: hates
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("defaults failures", func() {
				source := parseYAML(`
---
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))

use:
  subst: (( *foo.bar || "failed" ))
`)
				resolved, _ := Flow(parseYAML(`
---
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))
use:
  subst: failed
`))
				Expect(source).To(FlowAs(resolved))
			})

			It("defaults deep failures", func() {
				source := parseYAML(`
---
verbs: (( merge || nil ))
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verbs.verb " " alice ))

use:
  subst: (( *foo.bar || "failed" ))
`)
				resolved, _ := Flow(parseYAML(`
---
verbs: ~
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verbs.verb " " alice ))
use:
  subst: failed
`))
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("map template overriding", func() {
			It("overrides template substitution expression", func() {
				source := parseYAML(`
---
templ:
  <<: (( &template ))
  foo:
    bar: x

inst: (( *templ ))
`)
				stub := parseYAML(`
---
inst:
  bar: a 
  foo:
    bar: b 
    add: all

`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  <<: (( &template ))
  foo:
    bar: x

inst:
  bar: a 
  foo:
    bar: b 
    add: all
`))
				Expect(source).To(FlowAs(resolved, stub))
			})

			It("overrides template substitution", func() {
				source := parseYAML(`
---
templ:
  <<: (( &template ))
  foo:
    bar: x

inst: (( prefer *templ ))
`)
				stub := parseYAML(`
---
inst:
  bar: a 
  foo:
    bar: b 
    add: all

`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  <<: (( &template ))
  foo:
    bar: x

inst:
  foo:
    bar: b 
`))
				Expect(source).To(FlowAs(resolved, stub))
			})

			It("inserts into template substitution", func() {
				source := parseYAML(`
---
templ:
  <<: (( &template ))
  foo:
    <<: (( merge ))
    bar: x

inst: (( prefer *templ ))
`)
				stub := parseYAML(`
---
inst:
  bar: a 
  foo:
    bar: b 
    add: all

`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  <<: (( &template ))
  foo:
    <<: (( merge ))
    bar: x

inst:
  foo:
    add: all
    bar: b 
`))
				Expect(source).To(FlowAs(resolved, stub))
			})

			It("supports extended marker", func() {
				source := parseYAML(`
---
templ:
  <<: (( &template (merge) ))
  foo:
    bar: x

inst: (( prefer *templ ))
`)
				stub := parseYAML(`
---
inst:
  bar: a 
  foo:
    bar: b 
    add: all

`)
				resolved, _ := Flow(parseYAML(`
---
templ:
  <<: (( &template (merge) ))
  foo:
    bar: x

inst:
  bar: a 
  foo:
    bar: b 
`))
				Expect(source).To(FlowAs(resolved, stub))
			})
		})

	})

	Describe("merging lists with specified key", func() {
		Context("no merge", func() {
			It("clean up key tag", func() {
				source := parseYAML(`
---
list:
  - key:address: a
    attr: b
  - address: c
    attr: d
`)
				resolved := parseYAML(`
---
list:
  - address: a
    attr: b
  - address: c
    attr: d
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("auto merge with key tag", func() {
			It("overrides matching key entries", func() {
				source := parseYAML(`
---
list:
  - key:address: a
    attr: b
  - address: c
    attr: d
`)
				stub := parseYAML(`
---
list:
  - address: c
    attr: stub
  - address: e
    attr: f
`)
				resolved := parseYAML(`
---
list:
  - address: a
    attr: b
  - address: c
    attr: stub
`)
				Expect(source).To(FlowAs(resolved, stub))
			})
		})

		Context("explicit merge with key tag", func() {
			It("overrides matching key entries", func() {
				source := parseYAML(`
---
list:
  - <<: (( merge on address ))
  - address: a
    attr: b
  - address: c
    attr: d
`)
				stub := parseYAML(`
---
list:
  - address: c
    attr: stub
  - address: e
    attr: f
`)
				resolved := parseYAML(`
---
list:
  - address: e
    attr: f
  - address: a
    attr: b
  - address: c
    attr: stub
`)
				Expect(source).To(FlowAs(resolved, stub))
			})
		})
	})

	Describe("accessing the evaluation context", func() {
		It("resolves context variables", func() {
			source := parseYAML(`
---
foo:
  bar:
    path: (( __ctx.PATH ))
    str: (( __ctx.PATHNAME ))
    file: (( __ctx.FILE ))
    dir: (( __ctx.DIR ))
`)
			resolved := parseYAML(`
---
foo:
  bar:
    dir: .
    file: test
    path:
    - foo
    - bar
    - path
    str: foo.bar.str
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("valid values", func() {
		It("fails for undefined values", func() {
			source := parseYAML(`
---
foo: (( valid(bar) ))
`)

			resolved := parseYAML(`
---
foo: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("fails for nil values", func() {
			source := parseYAML(`
---
foo: (( valid(bar) ))
bar: ~
`)

			resolved := parseYAML(`
---
foo: false
bar: ~
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("fails for empty values", func() {
			source := parseYAML(`
---
foo: (( valid(bar) ))
bar:
`)

			resolved := parseYAML(`
---
foo: false
bar:
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("succeeds for empty maps", func() {
			source := parseYAML(`
---
foo: (( valid(bar) ))
bar: {}
`)

			resolved := parseYAML(`
---
foo: true
bar: {}
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("succeeds for empty arrays", func() {
			source := parseYAML(`
---
foo: (( valid(bar) ))
bar: []
`)

			resolved := parseYAML(`
---
foo: true
bar: []
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("fails for nil value", func() {
			source := parseYAML(`
---
foo: (( valid(~) ))
`)

			resolved := parseYAML(`
---
foo: false
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("require values", func() {
		It("checks for undefined values", func() {
			source := parseYAML(`
---
foo: (( require(bar) || "not set" ))
`)

			resolved := parseYAML(`
---
foo: not set
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("checks for nil values", func() {
			source := parseYAML(`
---
foo: (( require(bar) || "not set" ))
bar: ~
`)

			resolved := parseYAML(`
---
foo: not set
bar: ~
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("checks for no values", func() {
			source := parseYAML(`
---
foo: (( require(bar) || "not set" ))
bar:
`)

			resolved := parseYAML(`
---
foo: not set
bar:
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("passes values", func() {
			source := parseYAML(`
---
foo: (( require(bar) || "not set" ))
bar: x
`)

			resolved := parseYAML(`
---
foo: x
bar: x
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("meets docu", func() {
			source := parseYAML(`
---
foo: ~
bob: (( foo || "default" ))
alice: (( require(foo) || "default" ))
`)

			resolved := parseYAML(`
---
foo: ~
bob: ~
alice: default
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("undefined values", func() {
		It("eliminates undefined entries", func() {
			source := parseYAML(`
---
foo:
  alice: 24
  bob: (( ~~ ))
`)

			resolved := parseYAML(`
---
foo:
  alice: 24
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("eliminates entries evaluating to undefined", func() {
			source := parseYAML(`
---
foo:
  alice: 24
  bob: (( bar || ~~ ))
`)

			resolved := parseYAML(`
---
foo:
  alice: 24
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("meets docu", func() {
			source := parseYAML(`
---
foo: (( ~~ ))
bob: (( foo || ~~ ))
alice: (( bob || "default"))
`)

			resolved := parseYAML(`
---
alice: default
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when a dynamic index", func() {
		Context("for integer index", func() {
			It("it indexes an array", func() {
				source := parseYAML(`
---
index: 0

value: (( data.bob.[index].foo ))

data:
  bob:
    - foo: bar
`)
				resolved := parseYAML(`
---
index: 0

value: bar

data:
  bob:
    - foo: bar
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it indexes an array for deep evaluation", func() {
				source := parseYAML(`
---
fill: 0
index: (( fill ))

value: (( data.bob.[index].foo || "none" ))

data:
  bob:
    - foo: bar
`)
				resolved := parseYAML(`
---
fill: 0
index: 0

value: bar

data:
  bob:
    - foo: bar
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("for string index", func() {
			It("it accesses a map entry", func() {
				source := parseYAML(`
---
name: alice

value: (( data.[name].foo ))

data:
  alice:
    foo: bar
`)
				resolved := parseYAML(`
---
name: alice

value: bar

data:
  alice:
    foo: bar
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it accesses a deep map entry", func() {
				source := parseYAML(`
---
name:
  - foo
  - bar

value: (( data.[name] ))

data:
  foo:
    bar: alice
`)
				resolved := parseYAML(`
---
name:
  - foo
  - bar

value: alice

data:
  foo:
    bar: alice
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("for range literal", func() {
			It("handles positive increasing indices", func() {
				source := parseYAML(`
---
value: (( [1..3] ))
`)
				resolved := parseYAML(`
---
value:
  - 1
  - 2
  - 3
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("handled mixed increasing indices", func() {
				source := parseYAML(`
---
value: (( [-1..1] ))
`)
				resolved := parseYAML(`
---
value:
  - -1
  - 0
  - 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("handled mixed decreasing indices", func() {
				source := parseYAML(`
---
value: (( [1..-1] ))
`)
				resolved := parseYAML(`
---
value:
  - 1
  - 0
  - -1
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("for range index", func() {
			It("it extracts a slice for non-negative range", func() {
				source := parseYAML(`
---
value: (( data.[1..2] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - b
  - c

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts a complete slice for non-negative range", func() {
				source := parseYAML(`
---
value: (( data.[0..2] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - a
  - b
  - c

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts a slice for non-negative start range", func() {
				source := parseYAML(`
---
value: (( data.[1..] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - b
  - c

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts a slice for non-negative end range", func() {
				source := parseYAML(`
---
value: (( data.[..1] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - a
  - b

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts a slice for negative range", func() {
				source := parseYAML(`
---
value: (( data.[-2..-1] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - b
  - c

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts a complete slice for negative range", func() {
				source := parseYAML(`
---
value: (( data.[-3..-1] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - a
  - b
  - c

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts a slice for negative start range", func() {
				source := parseYAML(`
---
value: (( data.[-1..] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - c

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts a slice for negative end range", func() {
				source := parseYAML(`
---
value: (( data.[..-1] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value:
  - a
  - b
  - c

data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it extracts an empty slice", func() {
				source := parseYAML(`
---
value: (( data.[1..0] ))

data:
  - a
  - b
  - c
`)
				resolved := parseYAML(`
---
value: [ ]


data:
  - a
  - b
  - c
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when projecting", func() {
		Context("a list", func() {
			It("it handles an identity projection", func() {
				source := parseYAML(`
---
list:
  - name: a
    value: aValue
  - name: b
    value: bValue
  - name: c
    value: cValue

projection: (( .list.[*] ))
`)
				resolved := parseYAML(`
---
list:
  - name: a
    value: aValue
  - name: b
    value: bValue
  - name: c
    value: cValue
projection:
  - name: a
    value: aValue
  - name: b
    value: bValue
  - name: c
    value: cValue
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it handles a field projection", func() {
				source := parseYAML(`
---
list:
  - name: a
    value: aValue
  - name: b
    value: bValue
  - name: c
    value: cValue

projection: (( .list.[*].value ))
`)
				resolved := parseYAML(`
---
list:
  - name: a
    value: aValue
  - name: b
    value: bValue
  - name: c
    value: cValue
projection:
  - aValue
  - bValue
  - cValue
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it handles a field projection for a slice", func() {
				source := parseYAML(`
---
list:
  - name: a
    value: aValue
  - name: b
    value: bValue
  - name: c
    value: cValue

projection: (( .list.[1..2].value ))
`)
				resolved := parseYAML(`
---
list:
  - name: a
    value: aValue
  - name: b
    value: bValue
  - name: c
    value: cValue
projection:
  - bValue
  - cValue
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("a map", func() {
			It("it handles a value projection", func() {
				source := parseYAML(`
---
map:
  zz:
    name: a
    value: aValue
  xx:
    name: b
    value: bValue
  yy:
    name: c
    value: cValue

projection: (( .map.[*] ))
`)
				resolved := parseYAML(`
---
map:
  zz:
    name: a
    value: aValue
  xx:
    name: b
    value: bValue
  yy:
    name: c
    value: cValue
projection:
  - name: b
    value: bValue
  - name: c
    value: cValue
  - name: a
    value: aValue
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it handles a field value projection", func() {
				source := parseYAML(`
---
map:
  zz:
    name: a
    value: aValue
  xx:
    name: b
    value: bValue
  yy:
    name: c
    value: cValue

projection: (( .map.[*].value ))
`)
				resolved := parseYAML(`
---
map:
  zz:
    name: a
    value: aValue
  xx:
    name: b
    value: bValue
  yy:
    name: c
    value: cValue
projection:
  - bValue
  - cValue
  - aValue
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("in combination", func() {
			It("it handles chained projections", func() {
				source := parseYAML(`
---
map:
  zz:
    name: a
    value: aValue
  xx:
    name: b
    value: bValue
  yy:
    name: c
    value: cValue

projection: (( (.map.[*]).[1..2] ))
`)
				resolved := parseYAML(`
---
map:
  zz:
    name: a
    value: aValue
  xx:
    name: b
    value: bValue
  yy:
    name: c
    value: cValue
projection:
  - name: c
    value: cValue
  - name: a
    value: aValue
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("it handles nested projections", func() {
				source := parseYAML(`
---
list:
- zz:
    name: a
    value: aValue
- xx:
    name: b
    value: bValue
- yy:
    name: c
    value: cValue

projection: (( .list.[1..2].[*].value ))
`)
				resolved := parseYAML(`
---
list:
- zz:
    name: a
    value: aValue
- xx:
    name: b
    value: bValue
- yy:
    name: c
    value: cValue
projection:
  - - bValue
  - - cValue
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when merging inline maps", func() {
		It("it overrides field", func() {
			source := parseYAML(`
---
map1:
  alice: 24
  bob: 25
map2:
  alice: 26
  peter: 8
result: (( merge(map1,map2) ))
`)
			resolved := parseYAML(`
---
map1:
  alice: 24
  bob: 25
map2:
  alice: 26
  peter: 8
result:
  alice: 26
  bob: 25
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("it handles dynaml expressions", func() {
			source := parseYAML(`
---
map1:
  alice: 24
  bob: 25

map2:
  alice: 26
  peter: 8

result: (( merge(map1, map2, { "bob"="(( carl ))", "carl"=100 }) ))

`)
			resolved := parseYAML(`
---
map1:
  alice: 24
  bob: 25
map2:
  alice: 26
  peter: 8
result:
  alice: 26
  bob: 100
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("it handles templates", func() {
			source := parseYAML(`
---
data:
  alice: bob
  sub:
    foo: bar

template:
  <<: (( &template ))
  alice: (( merge sub ))
  ref: (( alice ))

result: (( merge(template, data) ))
`)
			resolved, _ := Flow(parseYAML(`
---
data:
  alice: bob
  sub:
    foo: bar
template:
  <<: (( &template ))
  alice: (( merge sub ))
  ref: (( alice ))
result:
  alice:
    foo: bar
  ref:
    foo: bar
`))
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when shifting network ranges", func() {
		Context("with arithmetic operator", func() {
			It("splits and shifts", func() {
				source := parseYAML(`
---
subnet: (( "10.1.2.1/24" / 12 ))
next: (( "10.1.2.1/24" / 12 * 2 ))
`)
				resolved := parseYAML(`
---
subnet: 10.1.2.0/28
next: 10.1.2.32/28
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("regression test for fixed errors", func() {
		It("nexted expressions for template markers", func() {
			source := parseYAML(`
---
template:
    <<: (( &template ( { } ( true ? {} :{} ) ) ))
data: (( *template ))
`)
			resolved, _ := Flow(parseYAML(`
---
template:
    <<: (( &template ( { } ( true ? {} :{} ) ) ))
data: {}
`))
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling base64", func() {
		Context("doing encoding", func() {
			It("it encodes a string", func() {
				source := parseYAML(`
---
value: (( base64("test") ))
`)
				resolved := parseYAML(`
---
value: dGVzdA==
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
		Context("doing decoding", func() {
			It("it decodes a string", func() {
				source := parseYAML(`
---
value: (( base64_decode("dGVzdA==") ))
`)
				resolved := parseYAML(`
---
value: test
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when calling md5", func() {
		It("it encodesgenerates md5 hash of a string", func() {
			source := parseYAML(`
---
value: (( md5("test") ))
`)
			resolved := parseYAML(`
---
value: 098f6bcd4621d373cade4e832627b4f6
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling bcrypt", func() {
		It("it crypts and validates a password", func() {
			source := parseYAML(`
---
value: (( bcrypt_check("test", bcrypt("test", 10)) ))
`)
			resolved := parseYAML(`
---
value: true
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when calling substr", func() {
		Context("with 2 args", func() {
			It("it handles positive start index", func() {
				source := parseYAML(`
---
value: (( substr("test",1) ))
`)
				resolved := parseYAML(`
---
value: est
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("it handles negative start index", func() {
				source := parseYAML(`
---
value: (( substr("test",-1) ))
`)
				resolved := parseYAML(`
---
value: t
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
		Context("with 3 args", func() {
			It("it handles positive start index", func() {
				source := parseYAML(`
---
value: (( substr("test",1,3) ))
`)
				resolved := parseYAML(`
---
value: es
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("it handles negative start index", func() {
				source := parseYAML(`
---
value: (( substr("test",-2,3) ))
`)
				resolved := parseYAML(`
---
value: s
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("it handles positive start index with negative end index", func() {
				source := parseYAML(`
---
value: (( substr("test",1,-1) ))
`)
				resolved := parseYAML(`
---
value: es
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("it handles negative start index with negative end index", func() {
				source := parseYAML(`
---
value: (( substr("test",-2,-1) ))
`)
				resolved := parseYAML(`
---
value: s
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("yaml and json", func() {
		Context("parsing", func() {
			It("it parses json", func() {
				source := parseYAML(`
---
json: |
    { "alice": 25 }

result: (( parse( json ) ))
`)
				resolved := parseYAML(`
---
json: |
    { "alice": 25 }
result:
    alice: 25
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("it transforms json", func() {
				source := parseYAML(`
---
data:
    alice: 25

result: (( asjson( data ) ))
`)
				resolved := parseYAML(`
---
data:
    alice: 25
result: '{"alice":25}'
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("it transforms yaml", func() {
				source := parseYAML(`
---
data:
    alice: 25

result: (( asyaml( data ) ))
`)
				resolved := parseYAML(`
---
data:
    alice: 25
result: |+
    alice: 25
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("catch", func() {
		Context("failed expressions", func() {
			It("provide error message", func() {
				source := parseYAML(`
---
fail: (( catch( 1 / 0 ) ))
`)
				resolved := parseYAML(`
---
fail:
    error: division by zero
    valid: false
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
		Context("valid expressions", func() {
			It("provide value", func() {
				source := parseYAML(`
---
fail: (( catch( 5 * 5 ) ))
`)
				resolved := parseYAML(`
---
fail:
    error: ""
    valid: true
    value: 25
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("sync", func() {
		Context("succeeded", func() {
			It("yields value", func() {
				source := parseYAML(`
---
data:
  alice: 25
result: (( sync( data, defined(value.alice), value.alice) ))
`)
				resolved := parseYAML(`
---
data:
  alice: 25
result: 25
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
		Context("timeout", func() {
			It("stops for succeeded evaluation", func() {
				source := parseYAML(`
---
data:
  alice: 25
result: (( catch(sync( data, defined(value.bob), value.bob, 1)) ))
`)
				resolved := parseYAML(`
---
data:
  alice: 25
result:
  error: sync timeout reached
  valid: false
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("stops for failed evaluation", func() {
				source := parseYAML(`
---
data:
  alice: 25
result: (( catch(sync( data.bob, defined(value.bob), value.bob, 1)) ))
`)
				resolved := parseYAML(`
---
data:
  alice: 25
result:
  error: "'data.bob' not found"
  valid: false
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})
	Describe("scoped expressions", func() {
		Context("in normal expressions", func() {
			It("accepts empty scopes", func() {
				source := parseYAML(`
---
alice: 1
bob: 2
scoped: (( () alice + bob ))
`)
				resolved := parseYAML(`
---
alice: 1
bob: 2
scoped: 3
`)
				Expect(source).To(FlowAs(resolved))
			})
			It("resolve scope fields", func() {
				source := parseYAML(`
---
alice: 1
bob: 2
scoped: (( ( $alice = 25, "bob" = 26 ) alice + bob ))
`)
				resolved := parseYAML(`
---
alice: 1
bob: 2
scoped: 51
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
		Context("in template expressions", func() {
			It("resolve scope fields in map templates", func() {
				source := parseYAML(`
---
alice: 1
template:
  <<: (( &template ))
  sum: (( alice + bob ))
scoped: (( ( $alice = 25, "bob" = 26 ) *template ))
`)
				resolved, _ := Flow(parseYAML(`
---
alice: 1
template:
  <<: (( &template ))
  sum: (( alice + bob ))
scoped:
  sum: 51
`))
				Expect(source).To(FlowAs(resolved))
			})
			It("resolve scope fields in value templates", func() {
				source := parseYAML(`
---
alice: 1
template: (( &template ( alice + bob ) ))
scoped: (( ( $alice = 25, "bob" = 26 ) *template ))
`)
				resolved, _ := Flow(parseYAML(`
---
alice: 1
template: (( &template ( alice + bob ) ))
scoped: 51
`))
				Expect(source).To(FlowAs(resolved))
			})
			It("resolve scope fields in list templates", func() {
				source := parseYAML(`
---
alice: 1
template: 
 - <<: (( &template ))
 - (( alice + bob ))
scoped: (( ( $alice = 25, "bob" = 26 ) *template ))
`)
				resolved, _ := Flow(parseYAML(`
---
alice: 1
template:
template: 
 - <<: (( &template ))
 - (( alice + bob ))
scoped:
  - 51
`))
				Expect(source).To(FlowAs(resolved))
			})
		})
	})
})
