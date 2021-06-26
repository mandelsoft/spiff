package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mandelsoft/spiff/flow"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "spiff",
	Short:   "YAML in-domain templating processor",
	Version: flow.VERSION,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}

func ReadFile(file string) ([]byte, error) {
	if strings.HasPrefix(file, "http:") || strings.HasPrefix(file, "https:") {
		response, err := http.Get(file)
		if err != nil {
			return nil, fmt.Errorf("error getting [%s]: %s", file, err)
		} else {
			if response.StatusCode != http.StatusOK {
				defer response.Body.Close()
				msg, _ := ioutil.ReadAll(response.Body)
				return nil, fmt.Errorf("[status %d]: %s", response.StatusCode, msg)
			}
			return ioutil.ReadAll(response.Body)
		}
	} else {
		return ioutil.ReadFile(file)
	}
}
