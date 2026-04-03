package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/nay-kia/portman/internal/config"
	"github.com/nay-kia/portman/internal/profile"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "📋 Manage port profiles",
	Long:  "Save, load, check, and manage port profiles for your projects.",
}

var profileSaveCmd = &cobra.Command{
	Use:     "save <name>",
	Short:   "Save current ports as a profile",
	Example: "  portman profile save myapp",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen, color.Bold)
		red := color.New(color.FgRed, color.Bold)
		dim := color.New(color.FgHiBlack)

		name := args[0]
		p, err := profile.Save(name)
		if err != nil {
			red.Printf("\n  ✗ Failed to save profile: %v\n\n", err)
			return
		}

		green.Printf("\n  ✓ Profile '%s' saved with %d port(s)\n\n", name, len(p.Ports))

		for _, entry := range p.Ports {
			dim.Printf("    ● :%d %s (%s)\n", entry.Port, entry.ProcessName, entry.Proto)
		}
		fmt.Println()
	},
}

var profileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all saved profiles",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan, color.Bold)
		yellow := color.New(color.FgYellow)
		dim := color.New(color.FgHiBlack)

		names, err := profile.List()
		if err != nil || len(names) == 0 {
			yellow.Println("\n  ⚠ No profiles saved yet.")
			dim.Println("  Use: portman profile save <name>")
			fmt.Println()
			return
		}

		cyan.Printf("\n  📋 Saved profiles (%d):\n\n", len(names))
		for _, name := range names {
			fmt.Printf("    ● %s\n", name)
		}
		fmt.Println()
	},
}

var profileCheckCmd = &cobra.Command{
	Use:   "check [name]",
	Short: "Check if ports are active (profile or .portman.yml)",
	Long: `Check port health from a saved profile or .portman.yml.

If a profile name is given, checks that profile's ports.
If no name is given, reads .portman.yml from the current directory.`,
	Example: `  portman profile check myapp      # Check saved profile
  portman profile check             # Check .portman.yml in current dir`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No name → check .portman.yml
			runConfigCheck()
		} else {
			// Name given → check saved profile
			runProfileCheck(args[0])
		}
	},
}

func runProfileCheck(name string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan, color.Bold)
	redBold := color.New(color.FgRed, color.Bold)

	results, err := profile.Check(name)
	if err != nil {
		redBold.Printf("\n  ✗ %v\n\n", err)
		return
	}

	cyan.Printf("\n  📋 Profile '%s' — checking %d port(s):\n\n", name, len(results))

	allOK := true
	for _, r := range results {
		if r.Active {
			green.Printf("    ✓ :%d %s — active (%s)\n", r.Entry.Port, r.Entry.ProcessName, r.Current)
		} else {
			red.Printf("    ✗ :%d %s — not listening\n", r.Entry.Port, r.Entry.ProcessName)
			allOK = false
		}
	}

	fmt.Println()
	if allOK {
		green.Println("  ✓ All ports are active!")
	} else {
		red.Println("  ✗ Some ports are not active")
	}
	fmt.Println()
}

func runConfigCheck() {
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
}

var profileDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Short:   "Delete a saved profile",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen, color.Bold)
		red := color.New(color.FgRed, color.Bold)

		name := args[0]
		if err := profile.Delete(name); err != nil {
			red.Printf("\n  ✗ Failed to delete: %v\n\n", err)
			return
		}

		green.Printf("\n  ✓ Profile '%s' deleted\n\n", name)
	},
}

func init() {
	profileCmd.AddCommand(profileSaveCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCheckCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	rootCmd.AddCommand(profileCmd)
}
