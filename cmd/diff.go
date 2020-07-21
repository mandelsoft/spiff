package cmd

import (
	"errors"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mandelsoft/spiff/compare"
	"github.com/mandelsoft/spiff/legacy/candiedyaml"
	"github.com/mandelsoft/spiff/yaml"
)

var separator string

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:     "diff",
	Aliases: []string{"d"},
	Short:   "Structurally compare two YAML files",
	Long: `Show structural differences between two deployment manifests.
Here streams with multiple documents are supported, also. To indicate 
no difference the number of documents in both streams must be identical
and each document in the first stream must have no difference compared
to the document with the same index in the second stream. Found differences
are shown for each document separately.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("requires two args")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		diff(args[0], args[1], separator)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&separator, "separator", "", "Separator to print between diffs")
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
