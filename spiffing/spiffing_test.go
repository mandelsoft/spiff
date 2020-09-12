package spiffing

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spiffing", func() {

	Context("Simple processing", func() {
		ctx, err := New().WithValues(map[string]interface{}{
			"values": map[string]interface{}{
				"alice": 25,
				"bob":   26,
			},
		})
		Expect(err).To(Succeed())

		It("Handles value document", func() {
			templ, err := ctx.Unmarshal("test", []byte("(( values.alice + values.bob ))"))
			Expect(err).To(Succeed())
			result, err := ctx.Cascade(templ, nil)
			Expect(err).To(Succeed())
			data, err := ctx.Marshal(result)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("51\n"))
		})
		It("Handles dynaml expression", func() {
			result, err := EvaluateDynamlExpression(ctx, "values.alice + values.bob")
			Expect(err).To(Succeed())
			Expect(string(result)).To(Equal("51"))
		})

		It("Handles complex dynaml expression", func() {
			result, err := EvaluateDynamlExpression(ctx, "[values.alice, values.bob]")
			Expect(err).To(Succeed())
			Expect(string(result)).To(Equal("- 25\n- 26\n"))
		})
	})
})
