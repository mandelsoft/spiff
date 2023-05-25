package spiffing

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/dynaml"
)

var _ = Describe("Spiffing", func() {

	Context("with functions", func() {
		It("handles nil set", func() {
			ctx := New().WithFunctions(nil)
			templ, err := ctx.Unmarshal("test", []byte("(( \"testvalue\" ))"))
			Expect(err).To(Succeed())
			result, err := ctx.Cascade(templ, nil)
			Expect(err).To(Succeed())
			data, err := ctx.Marshal(result)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("testvalue\n"))
		})
		It("calls a function", func() {
			f := func(arguments []interface{}, binding dynaml.Binding) (interface{}, dynaml.EvaluationInfo, bool) {
				return "testvalue", dynaml.EvaluationInfo{}, true
			}
			funcs := NewFunctions()
			funcs.RegisterFunction("testFunc", f)
			ctx := New().WithFunctions(funcs)
			templ, err := ctx.Unmarshal("test", []byte("(( testFunc() ))"))
			Expect(err).To(Succeed())
			result, err := ctx.Cascade(templ, nil)
			Expect(err).To(Succeed())
			data, err := ctx.Marshal(result)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal("testvalue\n"))
		})
	})

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

	Context("Bindings", func() {
		ctx, err := New().WithValues(map[string]interface{}{
			"values": map[string]interface{}{
				"alice": 25,
				"bob":   26,
			},
		})
		Expect(err).To(Succeed())

		It("Handles simple bindings", func() {
			templ, err := ctx.Unmarshal("test", []byte(`
data: (( values.alice ))
`))
			Expect(err).To(Succeed())
			result, err := ctx.Cascade(templ, nil)
			Expect(err).To(Succeed())
			data, err := ctx.Marshal(result)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal(
				`data: 25
`))
		})

		It("Handles override bindings", func() {
			templ, err := ctx.Unmarshal("test", []byte(`
values: other
data: (( ___.values.alice ))
orig: (( values ))
`))
			Expect(err).To(Succeed())
			result, err := ctx.Cascade(templ, nil)
			Expect(err).To(Succeed())
			data, err := ctx.Marshal(result)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal(
				`data: 25
orig: other
values: other
`))
		})
	})
})
