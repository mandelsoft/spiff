package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NamedArgs", func() {

	Context("named args in functions", func() {
		It("handles positional substitution", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( .data.func(a=1) ))
`)
			resolved := parseYAML(`
---
result:
  a: 1
  b: 1
  c: 2
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles positional and leading optional", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( .data.func(a=1, b=2) ))
`)
			resolved := parseYAML(`
---
result:
  a: 1
  b: 2
  c: 2
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles trailing optional argument", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( .data.func(c=3, 2) ))
`)
			resolved := parseYAML(`
---
result:
  a: 2
  b: 1
  c: 3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles indexed argiment", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( .data.func(3=3, 2) ))
`)
			resolved := parseYAML(`
---
result:
  a: 2
  b: 1
  c: 3
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("named args in currying", func() {
		It("handles currying last optional", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))
  curry:  (( data.func*(c=5) ))

result: (( .data.curry(1,2) ))
`)
			resolved := parseYAML(`
---
result:
  a: 1
  b: 2
  c: 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles currying non-optional by name", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))
  curry:  (( data.func*(a=5) ))

result: (( .data.curry(2,3) ))
`)
			resolved := parseYAML(`
---
result:
  a: 5
  b: 2
  c: 3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles mixed currying", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))
  curry:  (( data.func*(b=5, 4) ))

result: (( .data.curry(3) ))
`)
			resolved := parseYAML(`
---
result:
  a: 4
  b: 5
  c: 3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles named vararg currying", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b...|->{$a=a, $b=b } ))
  curry:  (( data.func*(b=[5], 4) ))

result: (( .data.curry() ))
`)
			resolved := parseYAML(`
---
result:
  a: 4
  b: 
  - 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles named non-vararg currying", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b...|->{$a=a, $b=b } ))
  curry:  (( data.func*(a=4) ))

result: (( .data.curry(5,6) ))
`)
			resolved := parseYAML(`
---
result:
  a: 4
  b: 
  - 5
  - 6
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Context("named args with errors", func() {
		It("handles invalid name", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( catch(.data.func(d=5)) ))
`)
			resolved := parseYAML(`
---
result:
  valid: false
  error: no lambda parameter found for named argument d
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("handles invalid index", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( catch(.data.func(4=5)) ))
`)
			resolved := parseYAML(`
---
result:
  valid: false
  error: argument index 4 too large for 3 parameters
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("too many arguments", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( catch(.data.func(1,2,3,4)) ))
`)
			resolved := parseYAML(`
---
result:
  valid: false
  error: found 4 argument(s), but expects 3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("too less arguments", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func:  (( |a,b=1,c=2|->{$a=a, $b=b, $c=c } ))

result: (( catch(.data.func()) ))
`)
			resolved := parseYAML(`
---
result:
  valid: false
  error: expected at least 1 arguments (2 optional), but found 0
`)
			Expect(source).To(FlowAs(resolved))
		})
	})
})
