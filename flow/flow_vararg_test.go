package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VarArgs", func() {

	Context("varargs in functions", func() {
		It("provides varargs as list", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func: (( |a,b...|-> [a] b ))

result: (( .data.func(1,2,3) ))
`)
			resolved := parseYAML(`
---
result:
  - 1
  - 2
  - 3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("provides single vararg as list", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func: (( |a,b...|-> [a] b ))

result: (( .data.func(1,2) ))
`)
			resolved := parseYAML(`
---
result:
  - 1
  - 2
`)
			Expect(source).To(FlowAs(resolved))
		})
		It("provides empty varargs as empty list", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  func: (( |a,b...|-> [a] b ))

result: (( .data.func(1) ))
`)
			resolved := parseYAML(`
---
result:
  - 1
`)
			Expect(source).To(FlowAs(resolved))
		})
	})
	Context("argument expansion", func() {
		It("expands arguments for regular functions", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  list:
  - 1
  - 2
  - 3
  func: (( |a,b,c|-> [ a, b, c] ))

result: (( .data.func(.data.list...) ))
`)
			resolved := parseYAML(`
---
result:
  - 1
  - 2
  - 3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("expands arguments in combination with explicit arguments for regular functions", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  list:
  - 1
  - 2
  func: (( |a,b,c|-> [ a, b, c] ))

result: (( .data.func(.data.list..., 3) ))
`)
			resolved := parseYAML(`
---
result:
  - 1
  - 2
  - 3
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("expands arguments in combination with explicit arguments for list literals", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  list:
  - 1
  - 2

result: (( [ 0, .data.list..., 3, .data.list..., 4 ] ))
`)
			resolved := parseYAML(`
---
result:
  - 0
  - 1
  - 2
  - 3
  - 1
  - 2
  - 4
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("expands arguments from function results", func() {
			source := parseYAML(`
---
data:
  <<: (( &temporary ))
  list:
  - 1
  - 2
  func: (( |a,b...|-> [a] b ))

result: (( [ 0, .data.func(.data.list...)..., 3 ] ))
`)
			resolved := parseYAML(`
---
result:
  - 0
  - 1
  - 2
  - 3
`)
			Expect(source).To(FlowAs(resolved))
		})
	})
})
