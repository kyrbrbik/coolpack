package coolpack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/coollabsio/coolpack/pkg/detector"
	"github.com/spf13/cobra"
)

var (
	planOutputJSON bool
	planPath       string
)

var planCmd = &cobra.Command{
	Use:   "plan [path]",
	Short: "Detect and plan the build for an application",
	Long: `Analyze the application at the given path (or current directory),
detect the language, framework, and package manager, then output a build plan.

Environment Variables:
  COOLPACK_BASE_IMAGE      Override base Docker image
  COOLPACK_NODE_VERSION    Override Node.js version`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPlan,
}

func init() {
	planCmd.Flags().BoolVar(&planOutputJSON, "json", false, "Output plan as JSON")
	planCmd.Flags().StringVarP(&planPath, "path", "p", "", "Path to the application (defaults to current directory)")
}

func runPlan(cmd *cobra.Command, args []string) error {
	// Determine the path to analyze
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	if planPath != "" {
		path = planPath
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
		fmt.Println("No supported application detected")
		return nil
	}

	// Output the plan
	if planOutputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(plan)
	}

	// Pretty print the plan
	printPlan(plan)
	return nil
}

func printPlan(plan *detector.Plan) {
	fmt.Println("=== Coolpack Build Plan ===")
	fmt.Println()
	fmt.Printf("Provider:                %s\n", plan.Provider)
	fmt.Printf("Language:                %s\n", plan.Language)
	if plan.LanguageVersion != "" {
		fmt.Printf("Language Version:        %s\n", plan.LanguageVersion)
	}
	if plan.Framework != "" {
		fmt.Printf("Framework:               %s\n", plan.Framework)
	}
	if plan.FrameworkVersion != "" {
		fmt.Printf("Framework Version:       %s\n", plan.FrameworkVersion)
	}
	if plan.PackageManager != "" {
		fmt.Printf("Package Manager:         %s\n", plan.PackageManager)
	}
	if plan.PackageManagerVersion != "" {
		fmt.Printf("Package Manager Version: %s\n", plan.PackageManagerVersion)
	}
	if plan.InstallCommand != "" {
		fmt.Printf("Install Command:         %s\n", plan.InstallCommand)
	}
	if plan.BuildCommand != "" {
		fmt.Printf("Build Command:           %s\n", plan.BuildCommand)
	}
	if plan.StartCommand != "" {
		fmt.Printf("Start Command:           %s\n", plan.StartCommand)
	}
	if len(plan.DetectedFiles) > 0 {
		fmt.Println()
		fmt.Println("Detected Files:")
		for _, f := range plan.DetectedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}
	if len(plan.Metadata) > 0 {
		fmt.Println()
		fmt.Println("Metadata:")
		// Sort keys for consistent output
		keys := make([]string, 0, len(plan.Metadata))
		for k := range plan.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %s: %v\n", k, plan.Metadata[k])
		}
	}
}
