package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("multiplication", func() {
	It("multiplies both numbers", func() {
		expr := MultiplicationExpr{
			IntegerExpr{2},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(6, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := MultiplicationExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := MultiplicationExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side is a CIDR", func() {
		It("shifts the IP range", func() {
			expr := MultiplicationExpr{
				StringExpr{"10.1.2.1/24"},
				IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs("10.1.5.0/24", FakeBinding{}))
		})
	})

	Context("multiplaction", func() {

		It("multiplies floats", func() {
			expr := MultiplicationExpr{
				FloatExpr{1.2},
				FloatExpr{2.2},
			}

			Expect(expr).To(EvaluateAs(2.64, FakeBinding{}))
		})
		It("multiplies ints and floats", func() {
			expr := MultiplicationExpr{
				IntegerExpr{2},
				FloatExpr{2.3},
			}

			Expect(expr).To(EvaluateAs(4.6, FakeBinding{}))
		})
		It("multiplies floats and ints", func() {
			expr := MultiplicationExpr{
				FloatExpr{2.3},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs(4.6, FakeBinding{}))
		})
	})
})
