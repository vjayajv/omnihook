package cmd

import (
	"fmt"
	"os"
	"github.com/vjayajv/omnihook/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listAll  bool
	hookType string
)

// Valid Git hook types
var validHookTypes = []string{
	"pre-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-push",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Git hooks",
	Long:  "List installed Git hooks. Use --all to list all hooks or --type to list specific hook type.",
	Run: func(cmd *cobra.Command, args []string) {
		hooksDir := utils.ExpandPath(viper.GetString("omni_hooks_dir"))

		if hooksDir == "" {
			fmt.Println("Error: Hooks directory not set. Run 'omnihook configure' first.")
			os.Exit(1)
		}

		if listAll || hookType == "all" {
			listAllHooks(hooksDir)
			return
		}

		// Validate hook type if specified
		if hookType != "" && !isValidHookType(hookType) {
			fmt.Printf("Invalid hook type: %s\n", hookType)
			fmt.Println("Valid hook types:")
			for _, t := range validHookTypes {
				fmt.Printf("  - %s\n", t)
			}
			return
		}

		// List specific hook type
		if hookType != "" {
			listHookType(hooksDir, hookType)
			return
		}

		cmd.Help()

	},
}

func listHookType(hooksDir, hookType string) {
	typeDir := fmt.Sprintf("%s/%s", hooksDir, hookType)
	files, err := os.ReadDir(typeDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No hooks installed for type: %s\n", hookType)
			return
		}
		fmt.Printf("Error accessing hooks directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No hooks installed for type: %s\n", hookType)
		return
	}

	fmt.Printf("Installed %s hooks:\n", hookType)
	for _, file := range files {
		if !file.IsDir() {
			fmt.Println("  -", file.Name())
		}
	}
}


func listAllHooks(hooksDir string) {
	dirs, err := os.ReadDir(hooksDir)
	if err != nil {
		fmt.Printf("Error reading hooks directory: %v\n", err)
		os.Exit(1)
	}

	if len(dirs) == 0 {
		fmt.Println("No hooks installed.")
		return
	}

	
	hookCount := 0
	for _, dir := range dirs {
		if dir.IsDir() {
			files, err := os.ReadDir(fmt.Sprintf("%s/%s", hooksDir, dir.Name()))
			if err != nil || len(files) == 0 {
				continue
			}
			hookCount++
			fmt.Println("└──", dir.Name())
			for _, file := range files {
				if !file.IsDir() {
					fmt.Println("    ├──", file.Name())
				}
			}
		}
	}
	if hookCount == 0 {
		fmt.Println("No hooks installed.")
	}
}

func isValidHookType(hookType string) bool {
	for _, valid := range validHookTypes {
		if valid == hookType {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listAll, "all", false, "List all installed hooks")
	listCmd.Flags().StringVar(&hookType, "type", "", "List hooks of specific type")
}
