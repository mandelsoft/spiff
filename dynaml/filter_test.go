package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("filter expressions", func() {
	Context("lists", func() {
		It("prints filter expression", func() {
			desc := MappingExpr{
				ReferenceExpr{Path: []string{"list"}},
				BooleanExpr{
					true,
				},
				FilterListContext,
			}.String()
			Expect(desc).To(Equal("filter[list|true]"))
		})

		It("simplifies lambda filter expression", func() {
			desc := MappingExpr{
				ReferenceExpr{Path: []string{"list"}},
				LambdaExpr{
					[]Parameter{Parameter{Name: "x"}},
					false,
					ReferenceExpr{Path: []string{"x"}},
				},
				FilterListContext,
			}.String()
			Expect(desc).To(Equal("filter[list|x|->x]"))
		})
	})

	Context("maps", func() {
		It("prints filter expression", func() {
			desc := MappingExpr{
				ReferenceExpr{Path: []string{"map"}},
				BooleanExpr{
					true,
				},
				FilterMapContext,
			}.String()
			Expect(desc).To(Equal("filter{map|true}"))
		})

		It("simplifies lambda filter expression", func() {
			desc := MappingExpr{
				ReferenceExpr{Path: []string{"map"}},
				LambdaExpr{
					[]Parameter{Parameter{Name: "x"}},
					false,
					ReferenceExpr{Path: []string{"x"}},
				},
				FilterMapContext,
			}.String()
			Expect(desc).To(Equal("filter{map|x|->x}"))
		})
	})
})
