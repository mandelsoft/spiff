package yaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func checkConvert(o, r string) {
	c := StringToExpression(o)
	Expect(c).To(Equal(r))
}

func checkUnescape(o, r string) {
	str, expr := convertToExpression(o, true)
	Expect(expr).To(BeNil())
	Expect(str).To(Not(BeNil()))
	Expect(*str).To(Equal(r))
}

var _ = Describe("Convert String", func() {

	s := "\\("
	_ = s
	Context("no substitution", func() {
		It("handles regular string", func() {
			checkConvert("test", "test")
		})
		It("handles partial subst", func() {
			checkConvert("test(", "test(")
			checkConvert("test((", "test((")
			checkConvert("((test)", "((test)")
			checkConvert("xx((test)", "xx((test)")
			checkConvert("xx((tes\"t))", "xx((tes\"t))")
		})

		Context("substitution", func() {
			It("handles single", func() {
				checkConvert("(( a + b ))", "(( a + b ))")
				checkConvert("start (( a + b )) end", "(( \"start \" a + b \" end\" ))")
				checkConvert("start (( a + b ))", "(( \"start \" a + b ))")
				checkConvert("(( a + b )) end", "(( a + b \" end\" ))")
				checkConvert(" (( a + b )) ", "(( \" \" a + b \" \" ))")
			})
			It("handles multiple", func() {
				checkConvert("start (( a )) middle (( b )) end", "(( \"start \" a \" middle \" b \" end\" ))")
			})

			It("handles brackets", func() {
				checkConvert("a start ((( a ))) end", "(( \"a start (\" a \") end\" ))")
			})

			It("brackets in expression", func() {
				checkConvert("a start (( ( b ( a )) )) end", "(( \"a start \" ( b ( a )) \" end\" ))")
				checkConvert("a start ((( ( b ( a )) ))) end", "(( \"a start (\" ( b ( a )) \") end\" ))")
			})

			It("handles quotes", func() {
				checkConvert("a start (( \"a\" b \"c\" )) end", "(( \"a start \" \"a\" b \"c\" \" end\" ))")
				checkConvert("b start (( \"))\" b )) end", "(( \"b start \" \"))\" b \" end\" ))")
				checkConvert("c start (( \"a(( x ))\" b )) end", "(( \"c start \" \"a(( x ))\" b \" end\" ))")
				checkConvert("d start (( \"\\\"))\\\" \" b )) end", "(( \"d start \" \"\\\"))\\\" \" b \" end\" ))")
			})

			It("handles mask", func() {
				checkConvert("a start \\(( a )) end", "(( \"a start \\\\\" a \" end\" ))")
				checkConvert("b start (\\( a )) end", "b start (\\( a )) end")

				checkConvert("c start (( \"\\\"\" )) end", "(( \"c start \" \"\\\"\" \" end\" ))")
				checkConvert("d start (( \\ )) end", "(( \"d start \" \\ \" end\" ))")
				checkConvert("e start (( \\\\ )) end", "(( \"e start \" \\\\ \" end\" ))")
			})

			It("handles mask at end", func() {
				checkConvert("a start (( a \\))) end", "(( \"a start \" a \\ \") end\" ))")
				checkConvert("b start (( a \\\\))) end", "(( \"b start \" a \\\\ \") end\" ))")
			})

			It("escaped expr", func() {
				checkConvert("a start ((! a \\))) end", "a start ((! a \\))) end")
			})
			It("mixed escaped expr", func() {
				checkConvert("a (( b )) start ((! a )) end", "(( \"a \" b \" start ((! a )) end\" ))")
			})
		})
	})

	Context("unescaping", func() {
		It("unescapes simple expr", func() {
			checkConvert("a start ((! a )) end", "a start ((! a )) end")
			checkUnescape("b start ((! a )) end", "b start (( a )) end")
		})
		It("unescapes double escaped expr", func() {
			checkConvert("a start ((!! a )) end", "a start ((!! a )) end")
			checkUnescape("b start ((!! a )) end", "b start ((! a )) end")
		})
		It("handles incomplete expr", func() {
			checkConvert("a start ((!! a ) end", "a start ((!! a ) end")
			checkUnescape("b start ((!! a ) end", "b start ((!! a ) end")
		})
	})
})
