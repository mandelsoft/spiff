package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Reporting issues for unresolved nodes", func() {

	It("reports unknown nodes", func() {
		source := parseYAML(`
---
node: (( ref ))
`)
		Expect(source).To(FlowToErr(
			`	(( ref ))	in test	node	()	*'ref' not found`,
		))
	})

	It("reports addition errors", func() {
		source := parseYAML(`
---
a: true
node: (( a + 1 ))
`)
		Expect(source).To(FlowToErr(
			`	(( a + 1 ))	in test	node	()	*first argument of PLUS must be IP address or integer`,
		))
	})

	It("reports subtraction errors", func() {
		source := parseYAML(`
---
a: true
node: (( a - 1 ))
`)
		Expect(source).To(FlowToErr(
			`	(( a - 1 ))	in test	node	()	*first argument of MINUS must be IP address or integer`,
		))
	})

	It("reports division by zero", func() {
		source := parseYAML(`
---
a: 1
node: (( a / 0 ))
`)
		Expect(source).To(FlowToErr(
			`	(( a / 0 ))	in test	node	()	*division by zero`,
		))
	})

	It("requires integer for second arith operand", func() {
		source := parseYAML(`
---
a: 1
node: (( a / true ))
`)
		Expect(source).To(FlowToErr(
			`	(( a / true ))	in test	node	()	*integer operand required`,
		))
	})

	It("reports merge failure", func() {
		source := parseYAML(`
---
node: (( merge ))
`)
		Expect(source).To(FlowToErr(
			`	(( merge ))	in test	node	(node)	*'node' not found in any stub`,
		))
	})

	It("reports merge redirect failure", func() {
		source := parseYAML(`
---
node: (( merge other.node))
`)
		Expect(source).To(FlowToErr(
			`	(( merge other.node ))	in test	node	(other.node)	*'other.node' not found in any stub`,
		))
	})

	It("reports join failure", func() {
		source := parseYAML(`
---
list:
  - a: true
node: (( join( ",", list.[0] ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( join(",", list.[0]) ))	in test	node	()	*argument 1 to join must be simple value or list`,
		))
	})

	It("reports join failure", func() {
		source := parseYAML(`
---
list:
  - a: true
node: (( join( [], "a" ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( join([], "a") ))	in test	node	()	*first argument for join must be a string`,
		))
	})

	It("reports join failure", func() {
		source := parseYAML(`
---
list:
  - a: true
node: (( join( ",", list ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( join(",", list) ))	in test	node	()	*elements of list(arg 1) to join must be simple values`,
		))
	})

	It("reports ip_min", func() {
		source := parseYAML(`
---
node: (( min_ip( "10" ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( min_ip("10") ))	in test	node	()	*CIDR argument required`,
		))
	})

	It("reports ip_min", func() {
		source := parseYAML(`
---
a:
- a
node: (( "." a ))
`)
		Expect(source).To(FlowToErr(
			`	(( "." a ))	in test	node	()	*type 'list' cannot be concatenated with type 'string'`,
		))
	})

	It("reports length", func() {
		source := parseYAML(`
---

node: (( length( 5 ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( length(5) ))	in test	node	()	*invalid type for function length`,
		))
	})

	It("reports list error in select{}", func() {
		source := parseYAML(`
---

node: (( select{[5]|x|->x} ))
`)
		Expect(source).To(FlowToErr(
			`	(( select{[5]|x|->x} ))	in test	node	()	*list value not supported for select mapping`,
		))
	})

	It("reports list error in map{}", func() {
		source := parseYAML(`
---

node: (( map{[5]|x|->x} ))
`)
		Expect(source).To(FlowToErr(
			`	(( map{[5]|x|->x} ))	in test	node	()	*list value not supported for map mapping`,
		))
	})

	It("reports unparseable", func() {
		source := parseYAML(`
---
node: (( a "." ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( a "." ) ))	in test	node	()	*parse error near symbol 7 - symbol 8: " "`,
		))
	})

	It("reports unparseable list insert operator", func() {
		source := parseYAML(`
---
node:
  - <<: (( a "." ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( a "." ) ))	in test	node.[0].<<	()	*parse error near symbol 7 - symbol 8: " "`,
		))
	})

	It("reports unparseable map insert operator", func() {
		source := parseYAML(`
---
node:
  <<: (( a "." ) ))
`)
		Expect(source).To(FlowToErr(
			`	(( a "." ) ))	in test	node.<<	()	*parse error near symbol 7 - symbol 8: " "`,
		))
	})

	It("reports unparseable map insert operator in multi line expression", func() {
		source := parseYAML(`
---
node:
  <<: |-
    ((
    a "." )
    ))
`)
		Expect(source).To(FlowToErr(
			`	((
	a "." )
	))	in test	node.<<	()	*parse error near line 2 symbol 6 - line 2 symbol 7: " "`,
		))
	})
})
