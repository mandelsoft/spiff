package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nested Cascades", func() {
	Context("single level nesting", func() {
		It("provides access to outer binding", func() {
			stub, _ := Flow(parseYAML(`
---
template:
  <<: (( &template ))
  add: 4
  bob: (( person.alice + .add ))
`))
			source := parseYAML(`
---
data:
   person:
     alice: 24
   merged: (( merge(stub(template)) ))
`)
			resolved := parseYAML(`
---
data:
   person:
     alice: 24
   merged:
     add: 4
     bob: 28
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("provides context", func() {
			stub, _ := Flow(parseYAML(`
---
template:
  <<: (( &template ))
  alice: (( __ctx.OUTER.[0].data.person.alice ))
`))
			source := parseYAML(`
---
data:
   person:
     alice: 24
   merged: (( merge(stub(template)) ))
`)
			resolved := parseYAML(`
---
data:
   person:
     alice: 24
   merged:
     alice: 24
`)
			Expect(source).To(FlowAs(resolved, stub))
		})
	})
	Context("multi level nesting", func() {
		It("provides context", func() {
			stub, _ := Flow(parseYAML(`
---
templates:
  level2:
    <<: (( &template ))
    level1: (( __ctx.OUTER.[0].value ))
    alice: (( __ctx.OUTER.[1].data.person.alice ))
  level1:
    <<: (( &template ))
    value: level1
    merged: (( merge(stub(level2)) ))
`))
			source := parseYAML(`
---
data:
   person:
     alice: 24
   merged: (( merge(stub(templates.level1), stub(templates)) ))
`)
			resolved := parseYAML(`
---
data:
   person:
     alice: 24
   merged:
     value: level1
     merged:
       level1: level1
       alice: 24
`)
			Expect(source).To(FlowAs(resolved, stub))
		})
	})
})
