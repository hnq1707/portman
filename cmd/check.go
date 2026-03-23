package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/config"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "✅ Check ports defined in .portman.yml",
	Long:  "Read .portman.yml from the current directory and verify all ports match their expected state.",
	Example: `  portman check
  
  # Example .portman.yml:
  # name: my-app
  # ports:
  #   - port: 3000
  #     name: Frontend
  #     expect: listening
  #   - port: 8080
  #     name: Backend
  #     expect: listening`,
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen)
		greenBold := color.New(color.FgGreen, color.Bold)
		red := color.New(color.FgRed)
		redBold := color.New(color.FgRed, color.Bold)
		cyan := color.New(color.FgCyan, color.Bold)

		cwd, _ := os.Getwd()
		cfg, err := config.LoadConfig(cwd)
		if err != nil {
			redBold.Printf("\n  ✗ %v\n\n", err)
			return
		}

		results, err := config.CheckConfig(cfg)
		if err != nil {
			redBold.Printf("\n  ✗ Failed to check: %v\n\n", err)
			return
		}

		projectName := cfg.Name
		if projectName == "" {
			projectName = "project"
		}

		cyan.Printf("\n  ✅ %s — checking %d port(s):\n\n", projectName, len(results))

		allOK := true
		for _, r := range results {
			label := r.Spec.Name
			if label == "" {
				label = fmt.Sprintf("port %d", r.Spec.Port)
			}

			if r.OK {
				green.Printf("    ✓ :%d %s — %s\n", r.Spec.Port, label, r.Message)
			} else {
				red.Printf("    ✗ :%d %s — %s\n", r.Spec.Port, label, r.Message)
				allOK = false
			}
		}

		fmt.Println()
		if allOK {
			greenBold.Println("  ✓ All ports OK!")
		} else {
			redBold.Println("  ✗ Some ports failed")
		}
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
