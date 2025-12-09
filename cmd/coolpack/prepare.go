package coolpack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coollabsio/coolpack/pkg/detector"
	"github.com/coollabsio/coolpack/pkg/generator"
	"github.com/spf13/cobra"
)

var (
	preparePath         string
	prepareBuildEnvs    []string
	prepareInstallCmd   string
	prepareBuildCmd     string
	prepareStartCmd     string
	prepareStaticServer string
	prepareOutputDir    string
	prepareSPA          bool
	prepareNoSPA        bool
)

var prepareCmd = &cobra.Command{
	Use:   "prepare [path]",
	Short: "Generate Dockerfile and build configuration",
	Long: `Analyze the application at the given path (or current directory),
detect the language, framework, and package manager, then generate
a Dockerfile and related build files in the .coolpack directory.

Environment Variables:
  COOLPACK_INSTALL_CMD     Override install command
  COOLPACK_BUILD_CMD       Override build command
  COOLPACK_START_CMD       Override start command
  COOLPACK_BASE_IMAGE      Override base Docker image (e.g., node:20)
  COOLPACK_NODE_VERSION    Override Node.js version
  COOLPACK_STATIC_SERVER   Static file server: caddy (default), nginx
  COOLPACK_SPA_OUTPUT_DIR  Override static output directory (e.g., dist, build)
  COOLPACK_SPA             Enable SPA mode (serves index.html for all routes)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPrepare,
}

func init() {
	prepareCmd.Flags().StringVarP(&preparePath, "path", "p", "", "Path to the application (defaults to current directory)")
	prepareCmd.Flags().StringArrayVar(&prepareBuildEnvs, "build-env", nil, "Build-time environment variables (KEY=value or KEY to use current env)")
	prepareCmd.Flags().StringVarP(&prepareInstallCmd, "install-cmd", "i", "", "Override install command")
	prepareCmd.Flags().StringVarP(&prepareBuildCmd, "build-cmd", "b", "", "Override build command")
	prepareCmd.Flags().StringVarP(&prepareStartCmd, "start-cmd", "s", "", "Override start command")
	prepareCmd.Flags().StringVar(&prepareStaticServer, "static-server", "", "Static file server: caddy (default), nginx")
	prepareCmd.Flags().StringVar(&prepareOutputDir, "output-dir", "", "Override static output directory (e.g., dist, build, out)")
	prepareCmd.Flags().BoolVar(&prepareSPA, "spa", false, "Enable SPA mode (serves index.html for all routes)")
	prepareCmd.Flags().BoolVar(&prepareNoSPA, "no-spa", false, "Disable SPA mode (overrides auto-detection)")
}

func runPrepare(cmd *cobra.Command, args []string) error {
	// Determine the path to analyze
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	if preparePath != "" {
		path = preparePath
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Run detection
	d := detector.New(absPath)
	plan, err := d.Detect()
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	if plan == nil {
		return fmt.Errorf("no supported application detected")
	}

	// Apply command overrides (CLI > env > detected)
	prepareApplyCommandOverrides(plan, prepareInstallCmd, prepareBuildCmd, prepareStartCmd)

	// Apply static server setting (CLI > env > default)
	prepareApplyStaticServerSetting(plan, prepareStaticServer)

	// Apply SPA setting (CLI > env > auto-detected)
	prepareApplySPASetting(plan, prepareSPA, prepareNoSPA)

	// Apply output directory override (CLI > env > framework default)
	prepareApplyOutputDirSetting(plan, prepareOutputDir)

	// Parse build environment variables
	envMap := prepareParseEnvVars(prepareBuildEnvs)
	if len(envMap) > 0 {
		plan.BuildEnv = envMap
	}

	// Create .coolpack directory
	coolpackDir := filepath.Join(absPath, ".coolpack")
	if err := os.MkdirAll(coolpackDir, 0755); err != nil {
		return fmt.Errorf("failed to create .coolpack directory: %w", err)
	}

	// Generate Dockerfile
	gen := generator.New(plan)
	dockerfile, err := gen.GenerateDockerfile()
	if err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Write Dockerfile
	dockerfilePath := filepath.Join(coolpackDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Printf("Generated files in %s:\n", coolpackDir)
	fmt.Printf("  - Dockerfile\n")

	return nil
}

// prepareParseEnvVars parses environment variable arguments
func prepareParseEnvVars(envArgs []string) map[string]string {
	result := make(map[string]string)
	for _, env := range envArgs {
		if idx := strings.Index(env, "="); idx != -1 {
			key := env[:idx]
			value := env[idx+1:]
			result[key] = value
		} else {
			if value, exists := os.LookupEnv(env); exists {
				result[env] = value
			}
		}
	}
	return result
}

// prepareApplyCommandOverrides applies command overrides from CLI flags or env vars
// Priority: CLI flags > Environment variables > Auto-detected
func prepareApplyCommandOverrides(plan *detector.Plan, installCmd, buildCmd, startCmd string) {
	// Install command: CLI > env > detected
	if installCmd != "" {
		plan.InstallCommand = installCmd
	} else if env := os.Getenv("COOLPACK_INSTALL_CMD"); env != "" {
		plan.InstallCommand = env
	}

	// Build command: CLI > env > detected
	if buildCmd != "" {
		plan.BuildCommand = buildCmd
	} else if env := os.Getenv("COOLPACK_BUILD_CMD"); env != "" {
		plan.BuildCommand = env
	}

	// Start command: CLI > env > detected
	if startCmd != "" {
		plan.StartCommand = startCmd
	} else if env := os.Getenv("COOLPACK_START_CMD"); env != "" {
		plan.StartCommand = env
	}
}

// prepareApplyStaticServerSetting applies static server setting from CLI or env var
// Priority: CLI flag > Environment variable > default (caddy)
func prepareApplyStaticServerSetting(plan *detector.Plan, staticServer string) {
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}

	if staticServer != "" {
		plan.Metadata["static_server"] = staticServer
	} else if env := os.Getenv("COOLPACK_STATIC_SERVER"); env != "" {
		plan.Metadata["static_server"] = env
	}
	// Default is "caddy" which is handled in generator
}

// prepareApplySPASetting applies SPA setting from CLI or env var
// Priority: --no-spa/COOLPACK_NO_SPA > --spa/COOLPACK_SPA > auto-detected
func prepareApplySPASetting(plan *detector.Plan, spa bool, noSPA bool) {
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}

	// --no-spa and COOLPACK_NO_SPA take highest priority
	if noSPA {
		delete(plan.Metadata, "is_spa")
		return
	}
	if env := os.Getenv("COOLPACK_NO_SPA"); env == "true" || env == "1" {
		delete(plan.Metadata, "is_spa")
		return
	}

	if spa {
		plan.Metadata["is_spa"] = true
	} else if env := os.Getenv("COOLPACK_SPA"); env == "true" || env == "1" {
		plan.Metadata["is_spa"] = true
	}
	// Auto-detected value is already in metadata from provider
}

// prepareApplyOutputDirSetting applies output directory override from CLI or env var
// Priority: CLI flag > Environment variable > framework default (handled in generator)
func prepareApplyOutputDirSetting(plan *detector.Plan, outputDir string) {
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}

	if outputDir != "" {
		plan.Metadata["output_dir_override"] = outputDir
	} else if env := os.Getenv("COOLPACK_SPA_OUTPUT_DIR"); env != "" {
		plan.Metadata["output_dir_override"] = env
	}
}
