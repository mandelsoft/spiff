package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("mapping expressions", func() {
	It("prints mapping expression", func() {
		desc := MappingExpr{
			ReferenceExpr{Path: []string{"list"}},
			ConcatenationExpr{
				ReferenceExpr{Path: []string{"x"}},
				StringExpr{".*"},
			},
			MapToListContext,
		}.String()
		Expect(desc).To(Equal("map[list|x \".*\"]"))
	})

	It("simplifies lambda mapping expression", func() {
		desc := MappingExpr{
			ReferenceExpr{Path: []string{"list"}},
			LambdaExpr{
				[]Parameter{Parameter{Name: "x"}},
				false,
				ConcatenationExpr{
					ReferenceExpr{Path: []string{"x"}},
					StringExpr{".*"},
				},
			},
			MapToListContext,
		}.String()
		Expect(desc).To(Equal("map[list|x|->x \".*\"]"))
	})
})
