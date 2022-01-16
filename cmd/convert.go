package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/legacy/candiedyaml"
	"github.com/mandelsoft/spiff/yaml"
)

// convertCmd represents the merge command
var convertCmd = &cobra.Command{
	Use:     "convert",
	Aliases: []string{"c"},
	Short:   "Convert template",
	Long:    `A given template file is normalized and converted to json or yaml.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires at one arg")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		convert(false, args[0], asJSON, split, outputPath, selection)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().BoolVar(&asJSON, "json", false, "print output in json format")
	convertCmd.Flags().StringVar(&outputPath, "path", "", "output is taken from given path")
	convertCmd.Flags().BoolVar(&split, "split", false, "if the output is alist it will be split into separate documents")
	convertCmd.Flags().StringArrayVar(&selection, "select", []string{}, "filter dedicated output fields")
}

func convert(stdin bool, templateFilePath string, json, split bool, subpath string, selection []string) {
	var templateFile []byte
	var err error

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
			flowed := templateYAML
			if subpath != "" {
				comps := dynaml.PathComponents(subpath, false)
				node, ok := yaml.FindR(true, flowed, nil, comps...)
				if !ok {
					log.Fatalln(fmt.Sprintf("path %q not found%s", subpath, doc))
				}
				flowed = node
			}

			if len(selection) > 0 {
				new := map[string]yaml.Node{}
				for _, p := range selection {
					comps := dynaml.PathComponents(p, false)
					node, ok := yaml.FindR(true, flowed, nil, comps...)
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
