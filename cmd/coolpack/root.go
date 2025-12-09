package coolpack

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "coolpack",
	Short: "A general purpose build pack for applications",
	Long: `Coolpack is a build pack tool that detects your application type,
generates Dockerfiles, and builds container images.

Currently supports:
  - Node.js (npm, yarn, pnpm, bun)

Environment Variables:
  COOLPACK_INSTALL_CMD     Override install command
  COOLPACK_BUILD_CMD       Override build command
  COOLPACK_START_CMD       Override start command
  COOLPACK_BASE_IMAGE      Override base Docker image (e.g., node:20-alpine)
  COOLPACK_NODE_VERSION    Override Node.js version
  COOLPACK_STATIC_SERVER   Static file server: caddy (default), nginx`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(prepareCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}
