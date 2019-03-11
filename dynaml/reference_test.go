package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/yaml"
)

var _ = Describe("references", func() {
	Context("when the reference is found", func() {
		It("evaluates to the referenced node", func() {
			expr := ReferenceExpr{[]string{"foo", "bar"}}

			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"foo":     NewNode(nil, nil),
					"foo.bar": NewNode(42, nil),
				},
			}

			Expect(expr).To(EvaluateAs(42, binding))
		})

		Context("and it refers to another expression", func() {
			It("returns itself so the referred node can evaluate first", func() {
				expr := ReferenceExpr{[]string{"foo", "bar"}}

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"foo": NewNode(MergeExpr{}, nil),
					},
				}

				Expect(expr).To(EvaluateAs(expr, binding))
			})
		})
	})

	Context("when the reference is NOT found", func() {
		It("fails", func() {
			expr := ReferenceExpr{[]string{"foo", "bar", "baz"}}

			binding := FakeBinding{}

			Expect(expr).To(FailToEvaluate(binding))
		})
	})
})
