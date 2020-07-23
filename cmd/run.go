package cmd

import (
	"errors"
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
)

// runCmd represents the merge command
var processCmd = &cobra.Command{
	Use:     "process",
	Aliases: []string{"r"},
	Short:   "Process a template with merged stubs on a document",
	Long:    `Merge a bunch of template files into one manifest processing a document input, printing it out.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least two args (document and template)")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		run(args[0], args[1], processingOptions, asJSON, split, outputPath, selection, state, args[2:])
	},
}

func init() {
	rootCmd.AddCommand(processCmd)

	processCmd.Flags().BoolVar(&asJSON, "json", false, "print output in json format")
	processCmd.Flags().BoolVar(&debug.DebugFlag, "debug", false, "Print state info")
	processCmd.Flags().BoolVar(&processingOptions.Partial, "partial", false, "Allow partial evaluation only")
	processCmd.Flags().StringVar(&outputPath, "path", "", "output is taken from given path")
	processCmd.Flags().StringVar(&state, "state", "", "select state file to maintain")
	processCmd.Flags().StringArrayVar(&selection, "select", []string{}, "filter dedicated output fields")
	processCmd.Flags().BoolVar(&processingOptions.PreserveEscapes, "preserve-escapes", false, "preserve escaping for escaped expressions and merges")
	processCmd.Flags().BoolVar(&processingOptions.PreserveTemporaray, "preserve-temporary", false, "preserve temporary fields")

}

func run(documentFilePath, templateFilePath string, opts flow.Options, json, split bool,
	subpath string, selection []string, stateFilePath string, stubFilePaths []string) {
	var err error
	var stdin = false
	var documentFile []byte

	if documentFilePath == "-" {
		documentFile, err = ioutil.ReadAll(os.Stdin)
		stdin = true
	} else {
		documentFile, err = ReadFile(documentFilePath)
	}

	documentYAML, err := yaml.Parse(documentFilePath, documentFile)
	if err != nil {
		log.Fatalln(fmt.Sprintf("error parsing template [%s]:", path.Clean(documentFilePath)), err)
	}

	documentYAML = yaml.NewNode(map[string]yaml.Node{"document": documentYAML}, "<"+documentFilePath+">")
	stub := yaml.NewNode(map[string]yaml.Node{"document": yaml.NewNode("(( &temporary &inject (merge) ))", "<document>)")}, "<document>")
	merge(stdin, templateFilePath, opts, json, split, subpath, selection, stateFilePath, []yaml.Node{stub, documentYAML}, stubFilePaths)
}
