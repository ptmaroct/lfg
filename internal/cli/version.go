package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ptmaroct/lfg/internal/version"
)

var versionVerbose bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the lfg version",
	Run: func(cmd *cobra.Command, args []string) {
		if versionVerbose {
			fmt.Println(version.Full())
			return
		}
		fmt.Println(version.Short())
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&versionVerbose, "verbose", "v", false,
		"include commit + build date")
	rootCmd.AddCommand(versionCmd)
}
