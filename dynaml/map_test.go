package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/yaml"
)

var _ = Describe("empty", func() {
	It("evaluates to empty hash", func() {
		Expect(CreateMapExpr{}).To(EvaluateAs(make(map[string]yaml.Node), FakeBinding{}))
	})
})
