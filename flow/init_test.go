package flow

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/yaml"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowing")
}

func parseYAML(source string, file ...string) yaml.Node {
	if len(file) == 0 {
		file = []string{"test"}
	}
	parsed, err := yaml.Parse(file[0], []byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}
