package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("division", func() {
	It("divides both numbers", func() {
		expr := DivisionExpr{
			IntegerExpr{6},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(2, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := DivisionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := DivisionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is zero", func() {
		It("fails", func() {
			expr := DivisionExpr{
				IntegerExpr{2},
				IntegerExpr{0},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side is a CIDR", func() {
		It("divides an IP range", func() {
			expr := DivisionExpr{
				StringExpr{"10.1.2.1/24"},
				IntegerExpr{4},
			}

			Expect(expr).To(EvaluateAs("10.1.2.0/26", FakeBinding{}))
		})
		It("rounds up divisor", func() {
			expr := DivisionExpr{
				StringExpr{"10.1.2.1/24"},
				IntegerExpr{12},
			}

			Expect(expr).To(EvaluateAs("10.1.2.0/28", FakeBinding{}))
		})
		It("fails for too large divisor", func() {
			expr := DivisionExpr{
				StringExpr{"10.1.2.1/24"},
				IntegerExpr{257},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
	Context("floats", func() {
		It("divides floats", func() {
			expr := DivisionExpr{
				FloatExpr{2.2},
				FloatExpr{1.1},
			}

			Expect(expr).To(EvaluateAs(2.0, FakeBinding{}))
		})
		It("divides ints and floats", func() {
			expr := DivisionExpr{
				IntegerExpr{3},
				FloatExpr{0.5},
			}

			Expect(expr).To(EvaluateAs(6.0, FakeBinding{}))
		})
		It("divides floats and ints", func() {
			expr := DivisionExpr{
				FloatExpr{2.2},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs(1.1, FakeBinding{}))
		})
		It("fails for zero", func() {
			expr := DivisionExpr{
				FloatExpr{2.2},
				FloatExpr{0.0},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})
