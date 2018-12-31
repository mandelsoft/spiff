package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/mandelsoft/spiff/compare"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
)

func main() {
	app := cli.NewApp()
	app.Name = "spiff"
	app.Usage = "BOSH deployment manifest toolkit"
	app.Version = "1.1.0-dev"

	app.Commands = []cli.Command{
		{
			Name:            "merge",
			ShortName:       "m",
			Usage:           "merge stub files into a manifest template",
			SkipFlagParsing: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "debug",
					Usage: "print state info",
				},
				cli.BoolFlag{
					Name:  "partial",
					Usage: "allow partial evaluation only",
				},
			},
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					cli.ShowCommandHelp(c, "merge")
					os.Exit(1)
				}
				debug.DebugFlag = c.Bool("debug")
				merge(c.Args()[0], c.Bool("partial"), c.Args()[1:])
			},
		},
		{
			Name:      "diff",
			ShortName: "d",
			Usage:     "structurally compare two YAML files",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "separator",
					Usage: "separator to print between diffs",
				},
			},
			Action: func(c *cli.Context) {
				if len(c.Args()) < 2 {
					cli.ShowCommandHelp(c, "diff")
					os.Exit(1)
				}

				diff(c.Args()[0], c.Args()[1], c.String("separator"))
			},
		},
	}

	app.Run(os.Args)
}

func merge(templateFilePath string, partial bool, stubFilePaths []string) {
	var templateFile []byte
	var err error
	var stdin = false

	if templateFilePath == "-" {
		templateFile, err = ioutil.ReadAll(os.Stdin)
		stdin = true
	} else {
		templateFile, err = ReadFile(templateFilePath)
	}

	if err != nil {
		log.Fatalln(fmt.Sprintf("error reading template [%s]:", path.Clean(templateFilePath)), err)
	}

	templateYAMLs, err := yaml.ParseMulti(templateFilePath, templateFile)
	if err != nil {
		log.Fatalln(fmt.Sprintf("error parsing template [%s]:", path.Clean(templateFilePath)), err)
	}

	stubs := []yaml.Node{}

	for _, stubFilePath := range stubFilePaths {
		var stubFile []byte
		var err error
		if stubFilePath == "-" {
			if stdin {
				log.Fatalln(fmt.Sprintf("stdin cannot be used twice"))
			}
			stubFile, err = ioutil.ReadAll(os.Stdin)
			stdin = true
		} else {
			stubFile, err = ReadFile(stubFilePath)
		}
		if err != nil {
			log.Fatalln(fmt.Sprintf("error reading stub [%s]:", path.Clean(stubFilePath)), err)
		}

		stubYAML, err := yaml.Parse(stubFilePath, stubFile)
		if err != nil {
			log.Fatalln(fmt.Sprintf("error parsing stub [%s]:", path.Clean(stubFilePath)), err)
		}

		stubs = append(stubs, stubYAML)
	}

	legend := "\nerror classification:\n" +
		" *: error in local dynaml expression\n" +
		" @: dependent of or involved in a cycle\n" +
		" -: depending on a node with an error"

	prepared, err := flow.PrepareStubs(partial, stubs...)
	if !partial && err != nil {
		log.Fatalln("error generating manifest:", err, legend)
	}

	for no, templateYAML := range templateYAMLs {
		doc := ""
		if len(templateYAMLs) > 1 {
			doc = fmt.Sprintf(" (document %d)", no+1)
		}
		if templateYAML.Value() != nil {
			flowed, err := flow.Apply(templateYAML, prepared)
			if !partial && err != nil {
				log.Fatalln(fmt.Sprintf("error generating manifest%s:", doc), err, legend)
			}
			if err != nil {
				flowed = dynaml.ResetUnresolvedNodes(flowed)
			}
			yaml, err := candiedyaml.Marshal(flowed)
			if err != nil {
				log.Fatalln(fmt.Sprintf("error marshalling manifest%s:", doc), err)
			}
			fmt.Println("---")
			fmt.Println(string(yaml))
		} else {
			fmt.Println("---")
		}
	}
}

func diff(aFilePath, bFilePath string, separator string) {
	aFile, err := ReadFile(aFilePath)
	if err != nil {
		log.Fatalln(fmt.Sprintf("error reading a [%s]:", path.Clean(aFilePath)), err)
	}

	aYAMLs, err := yaml.ParseMulti(aFilePath, aFile)
	if err != nil {
		log.Fatalln(fmt.Sprintf("error parsing a [%s]:", path.Clean(aFilePath)), err)
	}

	bFile, err := ReadFile(bFilePath)
	if err != nil {
		log.Fatalln(fmt.Sprintf("error reading b [%s]:", path.Clean(bFilePath)), err)
	}

	bYAMLs, err := yaml.ParseMulti(bFilePath, bFile)
	if err != nil {
		log.Fatalln(fmt.Sprintf("error parsing b [%s]:", path.Clean(bFilePath)), err)
	}

	if len(aYAMLs) != len(bYAMLs) {
		fmt.Printf("Different number of documents (%d != %d)\n", len(aYAMLs), len(bYAMLs))
		return
	}

	ddiffs := make([][]compare.Diff, len(aYAMLs))
	found := false
	for no, aYAML := range aYAMLs {
		bYAML := bYAMLs[no]
		ddiffs[no] = compare.Compare(aYAML, bYAML)
		if len(ddiffs[no]) != 0 {
			found = true
		}
	}
	if !found {
		fmt.Println("no differences!")
		return
	}
	for no := range aYAMLs {
		if len(ddiffs[no]) == 0 {
			if len(aYAMLs) > 1 {
				fmt.Println("No difference in document %d", no+1)
			}
		} else {
			diffs := ddiffs[no]
			doc := ""
			if len(aYAMLs) > 1 {
				doc = fmt.Sprintf("document %d", no+1)
			}
			for _, diff := range diffs {
				fmt.Println("Difference in", doc, strings.Join(diff.Path, "."))

				if diff.A != nil {
					ayaml, err := candiedyaml.Marshal(diff.A)
					if err != nil {
						panic(err)
					}

					fmt.Printf("  %s has:\n    \x1b[31m%s\x1b[0m\n", aFilePath, strings.Replace(string(ayaml), "\n", "\n    ", -1))
				}

				if diff.B != nil {
					byaml, err := candiedyaml.Marshal(diff.B)
					if err != nil {
						panic(err)
					}

					fmt.Printf("  %s has:\n    \x1b[32m%s\x1b[0m\n", bFilePath, strings.Replace(string(byaml), "\n", "\n    ", -1))
				}

				fmt.Printf(separator)
			}
		}
	}
}

func ReadFile(file string) ([]byte, error) {
	if strings.HasPrefix(file, "http:") || strings.HasPrefix(file, "https:") {
		response, err := http.Get(file)
		if err != nil {
			return nil, fmt.Errorf("error getting [%s]: %s", file, err)
		} else {
			defer response.Body.Close()
			return ioutil.ReadAll(response.Body)
		}
	} else {
		return ioutil.ReadFile(file)
	}
}
