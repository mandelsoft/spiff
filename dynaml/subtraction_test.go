package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("subtraction", func() {
	It("subtracts both numbers", func() {
		expr := SubtractionExpr{
			IntegerExpr{7},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(4, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := SubtractionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := SubtractionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side is an IP address", func() {
		It("subtracts from the IP address without carry", func() {
			expr := SubtractionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{1},
			}

			Expect(expr).To(EvaluateAs("10.10.10.9", FakeBinding{}))
		})

		It("adds to the IP address with single byte carry", func() {
			expr := SubtractionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{257},
			}

			Expect(expr).To(EvaluateAs("10.10.9.9", FakeBinding{}))
		})
	})

	Context("floats", func() {
		It("subtracts floats", func() {
			expr := SubtractionExpr{
				FloatExpr{1.25},
				FloatExpr{2.125},
			}

			Expect(expr).To(EvaluateAs(-0.875, FakeBinding{}))
		})
		It("subtracts ints and floats", func() {
			expr := SubtractionExpr{
				IntegerExpr{1},
				FloatExpr{2.25},
			}

			Expect(expr).To(EvaluateAs(-1.25, FakeBinding{}))
		})
		It("subtracts floats and ints", func() {
			expr := SubtractionExpr{
				FloatExpr{2.25},
				IntegerExpr{1},
			}

			Expect(expr).To(EvaluateAs(1.25, FakeBinding{}))
		})
	})
	Context("IPs", func() {
		It("subtracts ips and ips", func() {
			expr := SubtractionExpr{
				StringExpr{"10.0.0.10"},
				StringExpr{"10.0.0.1"},
			}

			Expect(expr).To(EvaluateAs(9, FakeBinding{}))
		})
		It("subtracts cidr and ips", func() {
			expr := SubtractionExpr{
				StringExpr{"10.0.0.10/24"},
				StringExpr{"10.0.0.1"},
			}

			Expect(expr).To(EvaluateAs(9, FakeBinding{}))
		})
		It("subtracts cidr and cidrs", func() {
			expr := SubtractionExpr{
				StringExpr{"10.0.0.10/24"},
				StringExpr{"10.0.0.1/30"},
			}

			Expect(expr).To(EvaluateAs(9, FakeBinding{}))
		})
		It("subtracts cidr and int", func() {
			expr := SubtractionExpr{
				StringExpr{"10.0.0.10/24"},
				IntegerExpr{9},
			}

			Expect(expr).To(EvaluateAs("10.0.0.1/24", FakeBinding{}))
		})

		It("adds ip and int", func() {
			expr := AdditionExpr{
				StringExpr{"10.0.0.10"},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs("10.0.0.12", FakeBinding{}))
		})
		It("adds cidr and int", func() {
			expr := AdditionExpr{
				StringExpr{"10.0.0.10/24"},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs("10.0.0.12/24", FakeBinding{}))
		})
	})
})
