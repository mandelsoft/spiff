package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionQuiet bool

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version for spiff++",
	Run: func(cmd *cobra.Command, args []string) {
		if versionQuiet {
			fmt.Println(rootCmd.Version)
		} else {
			fmt.Printf("%s version %s\n", rootCmd.Name(), rootCmd.Version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVarP(&versionQuiet, "quiet", "q", false, "print version only")
}
