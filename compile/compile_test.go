package compile

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compile", func() {
	Context("wrong types", func() {
		It("rejects unknown type", func() {
			_, err := Compile("test", map[string]interface{}{
				"val": struct{}{},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("val: unknown type (struct {})"))
		})
	})
	Context("Dynaml", func() {
		It("accepts plain yaml", func() {
			source := parseYAML(`
---
val: alice
`)
			_, err := Compile("test", source)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("detects compilation error", func() {
			source := parseYAML(`
---
val: (( blub( ))
`)
			_, err := Compile("test", source)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("val: parse error near symbol 7 - symbol 8: ' '"))
		})

		It("detects nested compilation error", func() {
			source := parseYAML(`
---
val: 
  alice: (( blub( ))
  nested:
    bob: (( map[] ))
`)
			_, err := Compile("test", source)
			Expect(err).To(HaveOccurred())
			Expect(strings.Split(err.Error(), "\n")).To(And(
				ContainElement(`val.nested.bob: parse error near symbol 5 - symbol 6: '['`),
				ContainElement(`val.alice: parse error near symbol 7 - symbol 8: ' '`)))
		})
	})
})
