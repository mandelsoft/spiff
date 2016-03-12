package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sum expressions", func() {
	It("prints sum expression", func() {
		desc := SumExpr{
			ReferenceExpr{[]string{"list"}},
			IntegerExpr{0},
			ConcatenationExpr{
				ReferenceExpr{[]string{"x"}},
				StringExpr{".*"},
			},
		}.String()
		Expect(desc).To(Equal("sum[list|0|x \".*\"]"))
	})

	It("simplifies lambda sum expression", func() {
		desc := SumExpr{
			ReferenceExpr{[]string{"list"}},
			IntegerExpr{0},
			LambdaExpr{
				[]string{"x"},
				ConcatenationExpr{
					ReferenceExpr{[]string{"x"}},
					StringExpr{".*"},
				},
			},
		}.String()
		Expect(desc).To(Equal("sum[list|0|x|->x \".*\"]"))
	})
})
