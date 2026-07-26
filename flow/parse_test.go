package flow

import (
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parse", func() {
	It("no interpolation", func() {
		node := yaml.NewNode(`test`, "dummy")

		s := yaml.EmbeddedDynaml(node, true)
		Expect(s).To(BeNil())
	})

	It("simple interpolation", func() {
		node := yaml.NewNode(`test (( "test" ))`, "dummy")

		s := yaml.EmbeddedDynaml(node, true)
		Expect(s).NotTo(BeNil())
		Expect(*s).To(Equal("\"test \" ( \"test\" )"))
	})

	It("conditional interpolation", func() {
		node := yaml.NewNode(`a (( (values.name == "") ? "prefix " values.server :"")) b`, "dummy")

		s := yaml.EmbeddedDynaml(node, true)
		Expect(s).NotTo(BeNil())
		Expect(*s).To(Equal("\"a \" ( (values.name == \"\") ? \"prefix \" values.server :\"\" ) \" b\""))
	})

	It("interpolates", func() {
		source := parseYAML(`
---
values: (( &temporary(merge) )) 
data: a (( (values.name == "") ? "prefix " values.server :"" )) b
`)

		stub := parseYAML(`
---
values:
  name: ""
  server: server
`)

		resolved := parseYAML(`
data: a prefix server b
`)
		Expect(source).To(FlowAs(resolved, stub).WithFeatures(features.INTERPOLATION))
	})
})
