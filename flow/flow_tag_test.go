package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tags", func() {
	Context("dynamic", func() {
		It("handles two arg form", func() {
			source := parseYAML(`
---
data: (( tagdef("tag:alice", 25) ))
alice: (( tag:alice::. ))
`)
			resolved := parseYAML(`
---
data: 25
alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles two arg global form", func() {
			source := parseYAML(`
---
data: (( tagdef("*tag:alice", 25) ))
alice: (( tag:alice::. ))
`)
			resolved := parseYAML(`
---
data: 25
alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles local three arg form", func() {
			source := parseYAML(`
---
data: (( tagdef("tag:alice", 25, "local") ))
alice: (( tag:alice::. ))
`)
			resolved := parseYAML(`
---
data: 25
alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles global three arg form", func() {
			source := parseYAML(`
---
data: (( tagdef("tag:alice", 25, "global") ))
alice: (( tag:alice::. ))
`)
			resolved := parseYAML(`
---
data: 25
alice: 25
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("fails stared three arg form", func() {
			source := parseYAML(`
---
data: (( catch(tagdef("*tag:alice", 25, "global")) ))
`)
			resolved := parseYAML(`
---
data:
  error: "invalid tag name \"*tag:alice\" for tagdef: invalid character \"*\" in tag component"
  valid: false
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("Regular", func() {
		It("handles tag field", func() {
			source := parseYAML(`
---
data:
  nested:
    v: (( tag::c ))
  a:
    b:
      <<: (( &tag:tag ))
      c: (( "value" ))
`)
			resolved := parseYAML(`
---
data:
  a:
    b:
      c: value
  nested:
    v: value

`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles tag", func() {
			source := parseYAML(`
---
data:
  nested:
    v: (( tag::. ))
  a:
    b:
      <<: (( &tag:tag ))
      c: (( "value" ))
`)
			resolved := parseYAML(`
---
data:
  a:
    b:
      c: value
  nested:
    v:
      c: value

`)
			Expect(source).To(FlowAs(resolved))
		})
		It("handles simple value tag", func() {
			source := parseYAML(`
---
data:
  nested:
    v: (( tag::. ))
  a:
    b:
      c: (( &tag:tag("value") ))
`)
			resolved := parseYAML(`
---
data:
  a:
    b:
      c: value
  nested:
    v: value
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("Failure", func() {
		It("unknown tags", func() {
			source := parseYAML(`
---
data:
  nested:
    v: (( catch( tag::c ) ))
  a:
    b:
      c: (( "value" ))
`)
			resolved := parseYAML(`
---
data:
  a:
    b:
      c: value
  nested:
    v:
      error: tag 'tag' not found
      valid: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("reports duplicate tags", func() {
			source := parseYAML(`
---
data:
  a: (( &tag:tag ))
  b: (( &tag:tag ))
`)
			Expect(source).To(FlowToErr(
				`	(( &tag:tag ))	in test	data.b	()	*duplicate tag "tag": data.b <-> data.a`,
			))
		})
	})

	Context("Tag Scopes", func() {
		It("propagates tag content", func() {
			source := parseYAML(`
---
tags:
  - <<: (( &temporary ))
  - <<: (( &tag:lib.alice ))
    data: alice.alice
  - <<: (( &tag:lib.alice.v1 ))
    data: alice.v1
usage:
   data: (( lib::data ))
`)
			resolved := parseYAML(`
---
usage:
  data: alice.alice
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("detects nonuniqe path resolution", func() {
			source := parseYAML(`
---
tags:
  - <<: (( &temporary ))
  - <<: (( &tag:lib.alice ))
    data: alice.alice
  - <<: (( &tag:lib.bob))
    data: bob
usage:
   data: (( catch(lib::data) ))
`)
			resolved := parseYAML(`
---
usage:
  data:
    error: 'ambigious tag resolution for lib::data: lib.alice <-> lib.bob'
    valid: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles context", func() {
			source := parseYAML(`
---
tags:
  - <<: (( &temporary ))
  - <<: (( &tag:lib.alice ))
    func: (( |x|->x * _.multiplier ))
    multiplier: 2
usage:
   data: (( lib::func(2) ))
`)
			resolved := parseYAML(`
---
usage:
  data: 4
`)
			Expect(source).To(FlowAs(resolved))
		})
	})
})
