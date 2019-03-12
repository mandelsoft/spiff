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
	app.Usage = "YAML in-domain templating processor"
	app.Version = "1.3.0-dev"

	app.Commands = []cli.Command{
		{
			Name:            "merge",
			ShortName:       "m",
			Usage:           "merge stub files into a manifest template",
			SkipFlagParsing: true,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "json",
					Usage: "print output in json format",
				},
				cli.BoolFlag{
					Name:  "debug",
					Usage: "print state info",
				},
				cli.BoolFlag{
					Name:  "partial",
					Usage: "allow partial evaluation only",
				},
				cli.BoolFlag{
					Name:  "split",
					Usage: "if the output is alist it will be split into separate documents",
				},
				cli.StringFlag{
					Name:  "path",
					Usage: "output is taken from given path",
				},
				cli.StringFlag{
					Name:  "state",
					Usage: "select state file to maintain",
				},
				cli.StringSliceFlag{
					Name:  "select",
					Usage: "filter dedicated output fields",
					Value: &cli.StringSlice{},
				},
			},
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					cli.ShowCommandHelp(c, "merge")
					os.Exit(1)
				}
				debug.DebugFlag = c.Bool("debug")
				merge(c.Args()[0], c.Bool("partial"),
					c.Bool("json"), c.Bool("split"),
					c.String("path"), c.StringSlice("select"),
					c.String("state"),
					c.Args()[1:])
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
		{
			Name:  "version",
			Usage: "show verson info",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "q",
					Usage: "print version only",
				},
			},
			Action: func(c *cli.Context) {
				if len(c.Args()) > 0 {
					cli.ShowCommandHelp(c, "version")
					os.Exit(1)
				}

				if c.Bool("q") {
					fmt.Printf("%s\n", app.Version)
				} else {
					fmt.Printf("%s version %s\n", app.Name, app.Version)
				}
			},
		},
	}

	app.Run(os.Args)
}

func keepAll(node yaml.Node) (yaml.Node, flow.CleanupFunction) {
	return node, keepAll
}

func discardNonState(node yaml.Node) (yaml.Node, flow.CleanupFunction) {
	if node.State() {
		return node, keepAll
	}
	return nil, discardNonState
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func merge(templateFilePath string, partial bool, json, split bool,
	subpath string, selection []string, stateFilePath string, stubFilePaths []string) {
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

	var stateData []byte

	if stateFilePath != "" {
		if len(templateYAMLs) > 1 {
			log.Fatalln(fmt.Sprintf("state handling not supported gor multi documents [%s]:", path.Clean(templateFilePath)), err)
		}
		if fileExists(stateFilePath) {
			stateData, err = ioutil.ReadFile(stateFilePath)
		}
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

	if stateData != nil {
		stateYAML, err := yaml.Parse(stateFilePath, stateData)
		if err != nil {
			log.Fatalln(fmt.Sprintf("error parsing state [%s]:", path.Clean(stateFilePath)), err)
		}
		stubs = append(stubs, stateYAML)
	}

	legend := "\nerror classification:\n" +
		" *: error in local dynaml expression\n" +
		" @: dependent of or involved in a cycle\n" +
		" -: depending on a node with an error"

	prepared, err := flow.PrepareStubs(nil, partial, stubs...)
	if !partial && err != nil {
		log.Fatalln("error generating manifest:", err, legend)
	}

	result := [][]byte{}
	count := 0
	for no, templateYAML := range templateYAMLs {
		doc := ""
		if len(templateYAMLs) > 1 {
			doc = fmt.Sprintf(" (document %d)", no+1)
		}
		var bytes []byte
		if templateYAML.Value() != nil {
			count++
			flowed, err := flow.Apply(nil, templateYAML, prepared)
			if !partial && err != nil {
				log.Fatalln(fmt.Sprintf("error generating manifest%s:", doc), err, legend)
			}
			if err != nil {
				flowed = dynaml.ResetUnresolvedNodes(flowed)
			}
			if subpath != "" {
				comps := dynaml.PathComponents(subpath, false)
				node, ok := yaml.FindR(true, flowed, comps...)
				if !ok {
					log.Fatalln(fmt.Sprintf("path %q not found%s", subpath, doc))
				}
				flowed = node
			}
			if stateFilePath != "" {
				state := flow.Cleanup(flowed, discardNonState)
				json := json
				if strings.HasSuffix(stateFilePath, ".yaml") || strings.HasSuffix(stateFilePath, ".yml") {
					json = false
				} else {
					if strings.HasSuffix(stateFilePath, ".json") {
						json = true
					}
				}
				if json {
					bytes, err = yaml.ToJSON(state)
				} else {
					bytes, err = candiedyaml.Marshal(state)
				}
				old := false
				if fileExists(stateFilePath) {
					os.Rename(stateFilePath, stateFilePath+".bak")
					old = true
				}
				err := ioutil.WriteFile(stateFilePath, bytes, 0664)
				if err != nil {
					os.Remove(stateFilePath)
					os.Remove(stateFilePath)
					if old {
						os.Rename(stateFilePath+".bak", stateFilePath)
					}
					log.Fatalln(fmt.Sprintf("cannot write state file %q", stateFilePath))
				}
			}
			if len(selection) > 0 {
				new := map[string]yaml.Node{}
				for _, p := range selection {
					comps := dynaml.PathComponents(p, false)
					node, ok := yaml.FindR(true, flowed, comps...)
					if !ok {
						log.Fatalln(fmt.Sprintf("path %q not found%s", subpath, doc))
					}
					new[comps[len(comps)-1]] = node

				}
				flowed = yaml.NewNode(new, "")
			}
			if split {
				if list, ok := flowed.Value().([]yaml.Node); ok {
					for _, d := range list {
						if json {
							bytes, err = yaml.ToJSON(d)
						} else {
							bytes, err = candiedyaml.Marshal(d)
						}
						if err != nil {
							log.Fatalln(fmt.Sprintf("error marshalling manifest%s:", doc), err)
						}
						result = append(result, bytes)
					}
					continue
				}
			}
			if json {
				bytes, err = yaml.ToJSON(flowed)
			} else {
				bytes, err = candiedyaml.Marshal(flowed)
			}
			if err != nil {
				log.Fatalln(fmt.Sprintf("error marshalling manifest%s:", doc), err)
			}
		}
		result = append(result, bytes)
	}

	for _, bytes := range result {
		if !json && (len(result) > 1 || len(bytes) == 0) {
			fmt.Println("---")
		}
		if bytes != nil {
			fmt.Print(string(bytes))
			if json {
				fmt.Println()
			}
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
				fmt.Printf("No difference in document %d\n", no+1)
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
