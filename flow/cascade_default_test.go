package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cascading YAML templates with defaults", func() {

	Context("for simple values", func() {
		It("defaults map entries", func() {
			source := parseYAML(`
---
data:
  foo: alice
`)
			secondary := parseYAML(`
---
data:
  bar: (( &default("bob") ))
`)
			resolved := parseYAML(`
---
data:
  foo: alice
  bar: bob
`)
			Expect(source).To(CascadeAs(resolved, secondary))
		})

		It("does not overwrite by defaults", func() {
			source := parseYAML(`
---
data:
  foo: alice
  bar: peter
`)
			secondary := parseYAML(`
---
data:
  bar: (( &default("bob") ))
`)
			resolved := parseYAML(`
---
data:
  foo: alice
  bar: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary))
		})

		It("defaults map entries indirectly", func() {
			source := parseYAML(`
---
data:
  foo: alice
`)
			secondary := parseYAML(`
---
data:
  nope: fault
`)
			tertiary := parseYAML(`
---
data:
  bar: (( &default("bob") ))
`)
			resolved := parseYAML(`
---
data:
  foo: alice
  bar: bob
`)
			Expect(source).To(CascadeAs(resolved, secondary, tertiary))
		})

		It("overwrites non-defaults", func() {
			source := parseYAML(`
---
data:
  foo: alice
  bar: claude
`)
			secondary := parseYAML(`
---
data:
  bar: (( &default("bob") ))
`)
			tertiary := parseYAML(`
---
data:
  bar: peter
`)
			resolved := parseYAML(`
---
data:
  foo: alice
  bar: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary, tertiary))
		})

		It("overwrites defaults", func() {
			source := parseYAML(`
---
data:
  foo: alice
`)
			secondary := parseYAML(`
---
data:
  bar: (( &default("bob") ))
`)
			tertiary := parseYAML(`
---
data:
  bar: peter
`)
			resolved := parseYAML(`
---
data:
  foo: alice
  bar: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary, tertiary))
		})
	})

	////////////////////////////////////////////////////////////////////////////////

	Context("for map values", func() {
		It("defaults complete maps", func() {
			source := parseYAML(`
---
data:
`)
			secondary := parseYAML(`
---
data:
  foobar:
    <<: (( &default ))
    foo: bob
`)
			resolved := parseYAML(`
---
data:
  foobar:
    foo: bob
`)
			Expect(source).To(CascadeAs(resolved, secondary))
		})

		It("does not overwrite by defaults", func() {
			source := parseYAML(`
---
data:
  foobar: 
    bar: peter
`)
			secondary := parseYAML(`
---
data:
  foobar:
    <<: (( &default ))
    foo: bob
`)
			resolved := parseYAML(`
---
data:
  foobar:
    bar: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary))
		})

		It("does not overwrite by entries of default map", func() {
			source := parseYAML(`
---
data:
  foobar: 
    foo: alice
    bar: peter
`)
			secondary := parseYAML(`
---
data:
  foobar:
    <<: (( &default ))
    foo: bob
`)
			resolved := parseYAML(`
---
data:
  foobar:
    foo: alice
    bar: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary))
		})

		It("does default entries of default map if these are marked as default", func() {
			source := parseYAML(`
---
data:
  foobar: 
    bar: peter
`)
			secondary := parseYAML(`
---
data:
  foobar:
    <<: (( &default ))
    foo: (( &default("bob") ))
    bar: claude
`)
			resolved := parseYAML(`
---
data:
  foobar:
    foo: bob
    bar: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary))
		})

		It("defaults map entries indirectly", func() {
			source := parseYAML(`
---
data: {}
`)
			secondary := parseYAML(`
---
data:
  nope: fault
`)
			tertiary := parseYAML(`
---
data:
  foobar:
    <<: (( &default ))
    foo: bob
`)
			resolved := parseYAML(`
---
data:
  foobar:
    foo: bob
`)
			Expect(source).To(CascadeAs(resolved, secondary, tertiary))
		})

		It("overwrites non-defaults", func() {
			source := parseYAML(`
---
data:
  foobar:
    bar: claude
`)
			secondary := parseYAML(`
---
data:
  foobar:
    <<: (( &default ))
    foo: bob
`)
			tertiary := parseYAML(`
---
data:
  foobar:
    bar: peter
`)
			resolved := parseYAML(`
---
data:
  foobar:
    bar: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary, tertiary))
		})

		It("overwrites defaults", func() {
			source := parseYAML(`
---
data: {}
`)
			secondary := parseYAML(`
---
data:
  foobar:
    <<: (( &default ))
    foo: bob
`)
			tertiary := parseYAML(`
---
data:
  foobar:
    foo: peter
`)
			resolved := parseYAML(`
---
data:
  foobar:
    foo: peter
`)
			Expect(source).To(CascadeAs(resolved, secondary, tertiary))
		})
	})
})
