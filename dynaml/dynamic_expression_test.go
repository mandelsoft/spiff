package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("dynamic references", func() {
	Context("when a dynamic string reference is found", func() {
		It("evaluates to the map entry", func() {
			ref := ReferenceExpr{[]string{"foo"}}
			idx := StringExpr{"bar"}
			expr := DynamicExpr{ref, idx}

			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"foo": node(map[string]yaml.Node{
						"bar": node(42, nil),
					}, nil),
				},
			}

			Expect(expr).To(EvaluateAs(42, binding))
		})
	})

	Context("when a dynamic array refernce is found", func() {
		It("evaluates to the indexed array entry", func() {
			ref := ReferenceExpr{[]string{"foo"}}
			idx := IntegerExpr{1}
			expr := DynamicExpr{ref, idx}
			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"foo": node([]yaml.Node{node(1, nil), node(42, nil)}, nil),
				},
			}

			Expect(expr).To(EvaluateAs(42, binding))
		})
	})
})
