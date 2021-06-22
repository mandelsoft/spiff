package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/features"
)

var _ = Describe("yaml control", func() {
	Context("select", func() {
		It("handles key", func() {
			source := parseYAML(`
---
key: (( x ))
x: test
selected:
  <<switch: (( key ))
  test: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: test
x: test
selected: alice
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
		It("handles default", func() {
			source := parseYAML(`
---
key: (( x ))
x: other
selected:
  <<switch: (( key ))
  test: alice
  <<default: bob
`)
			resolved := parseYAML(`
---
key: other
x: other
selected: bob
`)
			Expect(source).To(FlowAs(resolved).WithFeatures(features.CONTROL))
		})
	})

	Context("select cascade", func() {
		It("handles key", func() {
			source := parseYAML(`
---
key: (( x ))
x: test
selected:
  <<switch: (( key ))
  test: alice
  <<default: bob
`)
			stub := parseYAML(`
---
selected: peter
`)
			resolved := parseYAML(`
---
key: test
x: test
selected: peter
`)
			Expect(source).To(FlowAs(resolved, stub).WithFeatures(features.CONTROL))
		})
	})
})
