package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/yaml"
)

var _ = Describe("or", func() {
	Context("when both sides fail", func() {
		It("fails", func() {
			expr := ValidOrExpr{
				FailingExpr{},
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side fails", func() {
		It("returns the right-hand side", func() {
			expr := ValidOrExpr{
				FailingExpr{},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{}))
		})
	})

	Context("when the right-hand side fails", func() {
		It("returns the left-hand side", func() {
			expr := ValidOrExpr{
				IntegerExpr{1},
				FailingExpr{},
			}

			Expect(expr).To(EvaluateAs(1, FakeBinding{}))
		})
	})

	Context("when the left-hand side is nil", func() {
		It("returns the right-hand side", func() {
			expr := ValidOrExpr{
				NilExpr{},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{}))
		})
		It("fails if right-hand side fails", func() {
			expr := ValidOrExpr{
				NilExpr{},
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right side is nil and the left fails", func() {
		It("returns the left-hand side", func() {
			expr := ValidOrExpr{
				FailingExpr{},
				NilExpr{},
			}

			Expect(expr).To(EvaluateAs(nil, FakeBinding{}))
		})
	})

	Context("when the left side evaluates to itself (i.e. reference)", func() {
		It("fails assuming the left hand side cannot be determined yet", func() {
			expr := ValidOrExpr{
				ReferenceExpr{Path: []string{"foo", "bar"}},
				NilExpr{},
			}

			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"foo": NewNode(MergeExpr{}, nil),
				},
			}

			Expect(expr).To(FailToEvaluate(binding))
		})
	})
})
