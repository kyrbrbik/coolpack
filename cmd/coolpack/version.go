package coolpack

import (
	"fmt"

	"github.com/coollabsio/coolpack/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("coolpack %s\n", version.Version)
		if version.Commit != "none" {
			fmt.Printf("  commit: %s\n", version.Commit)
		}
		if version.Date != "unknown" {
			fmt.Printf("  built:  %s\n", version.Date)
		}

		// Check for updates
		version.CheckForUpdate()
	},
}
