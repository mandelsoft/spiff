package main

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/mandelsoft/spiff/yaml"
)

var spiff string

func Test(t *testing.T) {
	BeforeSuite(func() {
		var err error
		spiff, err = gexec.Build("github.com/mandelsoft/spiff")
		Ω(err).ShouldNot(HaveOccurred())
		fmt.Printf("executable: %s\n", spiff)
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Executable")
}

func parseYAML(source string) yaml.Node {
	parsed, err := yaml.Parse("test", []byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}
