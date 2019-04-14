package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/mandelsoft/spiff/dynaml/passwd"
	"github.com/mandelsoft/spiff/yaml"
)

var decrypt bool

// encryptCmd represents the diff command
var encryptCmd = &cobra.Command{
	Use:     "encrypt <file> [<password>] [<method>]",
	Aliases: []string{"e"},
	Short:   "Encrypt/Decrypt yaml document",
	Long:    ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 3 {
			return errors.New("requires one, two or three args")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		encrypt(decrypt, args)
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.Flags().BoolVarP(&decrypt, "decrypt", "d", false, "decrypt content")
}

func encrypt(decrypt bool, args []string) {
	var file []byte
	var err error

	filePath := args[0]

	if filePath == "-" {
		file, err = ioutil.ReadAll(os.Stdin)
	} else {
		file, err = ReadFile(filePath)
	}

	if err != nil {
		log.Fatalln(fmt.Sprintf("error reading data [%s]:", path.Clean(filePath)), err)
	}

	key := os.Getenv("SPIFF_ENCRYPTION_KEY")
	method := passwd.TRIPPLEDES
	v := ""
	if len(args) > 1 {
		v = args[1]
	}

	switch len(args) {
	case 2:
		if passwd.GetEncoding(v) != nil {
			method = v
		} else {
			key = v
		}
	case 3:
		method = args[2]
	}

	if key == "" {
		log.Fatalln("invalid empty encyption key")
	}

	e := passwd.GetEncoding(method)
	if e == nil {
		log.Fatalf("invalid encyption method %q", method)
	}

	if key == "" {
		log.Fatalf("invalid empty encyption key")
	}

	if decrypt {
		result, err := e.Decode(string(file), key)
		if err != nil {
			log.Fatalln(fmt.Sprintf("error decoding data [%s]:", path.Clean(filePath)), err)
		}
		fmt.Printf("%s\n", result)
	} else {
		_, err := yaml.Parse(filePath, file)
		if err != nil {
			log.Fatalln(err)
		}
		result, err := e.Encode(string(file), key)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%s\n", result)
	}
}
