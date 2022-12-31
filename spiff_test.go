package main

import (
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Running spiff", func() {
	Context("merge", func() {
		var merge *Session

		Context("when given a bad file path", func() {
			BeforeEach(func() {
				var err error
				merge, err = Start(exec.Command(spiff, "merge", "foo.yml"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})

			It("says file not found", func() {
				Expect(merge.Wait()).To(Exit(1))
				Expect(merge.Err).To(Say("foo.yml: no such file or directory"))
			})
		})

		Context("when given a single file", func() {
			var basicTemplate *os.File

			BeforeEach(func() {
				var err error

				basicTemplate, err = ioutil.TempFile(os.TempDir(), "basic.yml")
				Expect(err).NotTo(HaveOccurred())
				basicTemplate.Write([]byte(`
---
foo: bar
`))
				merge, err = Start(exec.Command(spiff, "merge", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.Remove(basicTemplate.Name())
			})

			It("resolves the template and prints it out", func() {
				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo: bar`))
			})
		})

		Context("when given values", func() {
			var basicTemplate *os.File
			BeforeEach(func() {
				var err error

				basicTemplate, err = ioutil.TempFile(os.TempDir(), "basic.yml")
				Expect(err).NotTo(HaveOccurred())
				basicTemplate.Write([]byte(`
---
foo: (( values ))
`))
			})

			AfterEach(func() {
				os.Remove(basicTemplate.Name())
			})

			It("resolves the template with flat definition", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues=X", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo: X`))
			})

			It("resolves the template with deep definition", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues.alice=X", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo:
  alice: X`))
			})

			It("resolves the template with multiple deep definitions", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues.alice=X", "-Dvalues.bob=Z", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo:
  alice: X
  bob: Z`))
			})

			It("resolves the template with escaped deep definitions", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues.alice\\.bob=X", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo:
  alice.bob: X`))
			})

			It("resolves the template with escaped dot at end", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues.alice\\..bob=X", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo:
  alice.:
    bob: X`))
			})

			It("resolves the template with escaped \\ deep definitions", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues.alice\\\\.bob=X", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo:
  alice\\:
    bob: X`))
			})

			It("fails for inconsistent definitions", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues=X", "-Dvalues.alice=X", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(1))
				Expect(merge.Err).To(Say(`.*error in value definitions \(-D\): field "values" in values.alice is no map`))
			})

			It("fails for invalid definitions", func() {
				merge, err := Start(exec.Command(spiff, "merge", "-Dvalues..alice=X", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Expect(merge.Wait()).To(Exit(1))
				Expect(merge.Err).To(Say(`.*error in value definitions \(-D\): empty path component in "values..alice"`))
			})
		})

	})
})
