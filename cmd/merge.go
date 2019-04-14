package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
	"github.com/spf13/cobra"
)

var asJSON bool
var partial bool
var outputPath string
var selection []string
var split bool
var state string

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:     "merge",
	Aliases: []string{"m"},
	Short:   "Merge stub files into a manifest template",
	Long:    `Merge a bunch of template files into one manifest, printing it out.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one arg")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		merge(args[0], partial, asJSON, split, outputPath, selection, state, args[1:])
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().BoolVar(&asJSON, "json", false, "print output in json format")

	mergeCmd.Flags().BoolVar(&debug.DebugFlag, "debug", false, "Print state info")

	mergeCmd.Flags().BoolVar(&partial, "partial", false, "Allow partial evaluation only")

	mergeCmd.Flags().StringVar(&outputPath, "path", "", "output is taken from given path")

	mergeCmd.Flags().BoolVar(&split, "split", false, "if the output is alist it will be split into separate documents")

	mergeCmd.Flags().StringVar(&state, "state", "", "select state file to maintain")

	mergeCmd.Flags().StringArrayVar(&selection, "select", []string{}, "filter dedicated output fields")
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
