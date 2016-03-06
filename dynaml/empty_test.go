package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("empty", func() {
	It("evaluates to empty hash", func() {
		Expect(EmptyHashExpr{}).To(EvaluateAs(make(map[string]yaml.Node), FakeBinding{}))
	})
})
