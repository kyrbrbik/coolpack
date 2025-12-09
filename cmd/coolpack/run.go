package coolpack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/coollabsio/coolpack/pkg/detector"
	"github.com/spf13/cobra"
)

var (
	runPath      string
	runImageName string
	runTag       string
	runEnvVars   []string
)

var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "Run the built container (DEVELOPMENT ONLY)",
	Long: `
╔══════════════════════════════════════════════════════════════════════════════╗
║                              ⚠️  WARNING ⚠️                                   ║
║                                                                              ║
║  This command is for DEVELOPMENT and TESTING of Coolpack only!              ║
║                                                                              ║
║  DO NOT use this in production environments.                                 ║
║  For production, use proper container orchestration (Docker Compose,        ║
║  Kubernetes, etc.) with appropriate configuration.                          ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

Run the built container image in development mode with:
  - Interactive terminal (-it)
  - Auto-remove on exit (--rm)
  - Port mapping based on detected output type`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringVarP(&runPath, "path", "p", "", "Path to the application (defaults to current directory)")
	runCmd.Flags().StringVarP(&runImageName, "name", "n", "", "Image name (defaults to directory name)")
	runCmd.Flags().StringVarP(&runTag, "tag", "t", "latest", "Image tag")
	runCmd.Flags().StringArrayVarP(&runEnvVars, "env", "e", nil, "Environment variables (KEY=value)")
}

func runRun(cmd *cobra.Command, args []string) error {
	// Print warning
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                              ⚠️  WARNING ⚠️                                   ║")
	fmt.Println("║         This is for DEVELOPMENT ONLY - Do NOT use in production!            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Determine the path to analyze
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	if runPath != "" {
		path = runPath
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

	// Determine image name
	imageName := runImageName
	if imageName == "" {
		imageName = filepath.Base(absPath)
		imageName = strings.ToLower(imageName)
		imageName = strings.ReplaceAll(imageName, " ", "-")
	}

	fullImageName := fmt.Sprintf("%s:%s", imageName, runTag)

	// Run detection to get output type for port
	d := detector.New(absPath)
	plan, err := d.Detect()
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	if plan == nil {
		return fmt.Errorf("no supported application detected")
	}

	// Determine port based on output type
	port := "3000"
	if ot, ok := plan.Metadata["output_type"].(string); ok && ot == "static" {
		port = "80"
	}

	// Build docker run arguments
	dockerArgs := []string{"run", "--rm", "-it", "-p", fmt.Sprintf("%s:%s", port, port)}

	// Add environment variables
	for _, env := range runEnvVars {
		dockerArgs = append(dockerArgs, "-e", env)
	}

	dockerArgs = append(dockerArgs, fullImageName)

	// Run docker
	if len(runEnvVars) > 0 {
		fmt.Printf("Running: docker run --rm -it -p %s:%s", port, port)
		for _, env := range runEnvVars {
			fmt.Printf(" -e %s", env)
		}
		fmt.Printf(" %s\n\n", fullImageName)
	} else {
		fmt.Printf("Running: docker run --rm -it -p %s:%s %s\n\n", port, port, fullImageName)
	}

	dockerCmd := exec.Command("docker", dockerArgs...)
	dockerCmd.Stdin = os.Stdin
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr

	if err := dockerCmd.Run(); err != nil {
		return fmt.Errorf("docker run failed: %w", err)
	}

	return nil
}
