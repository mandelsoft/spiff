package compile

import (
	"bytes"
	"testing"

	"github.com/cloudfoundry-incubator/candiedyaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compiling")
}

func parseYAML(source string) interface{} {
	r := bytes.NewBuffer([]byte(source))
	d := candiedyaml.NewDecoder(r)
	for d.HasNext() {
		var parsed interface{}
		err := d.Decode(&parsed)
		if err != nil {
			panic(err)
		}
		return parsed
	}
	return nil
}
