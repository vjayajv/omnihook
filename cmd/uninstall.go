package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall a specific hook or all hooks",
	RunE:  uninstallHook,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().String("id", "", "ID of the hook to uninstall")
	uninstallCmd.Flags().Bool("all", false, "Remove all installed hooks")
	uninstallCmd.Flags().String("type", "", "Remove all installed hooks of a specific type")
}

func uninstallHook(cmd *cobra.Command, args []string) error {
	hooksDir := viper.GetString("omni_hooks_dir")
	if hooksDir == "" {
		return fmt.Errorf("hooks directory not set. Run 'omnihook configure' first")
	}

	hookType, _ := cmd.Flags().GetString("type")
	hookID, _ := cmd.Flags().GetString("id")
	removeAll, _ := cmd.Flags().GetBool("all")

	if removeAll {
		return uninstallAllHooks(hooksDir)
	}

	if hookID != "" && hookType == "" {
		return fmt.Errorf("--type is required when --id is specified")
	}

	if hookType == "" && !removeAll {
		return fmt.Errorf("either --type, --id (with --type), or --all must be specified")
	}

	if hookID != "" {
		return uninstallSingleHook(hooksDir, hookType, hookID)
	}

	return uninstallHookType(hooksDir, hookType)
}

func uninstallAllHooks(hooksDir string) error {
	hookFiles, err := filepath.Glob(filepath.Join(hooksDir, "*"))
	if err != nil {
		return fmt.Errorf("failed to list hooks: %w", err)
	}

	if len(hookFiles) == 0 {
		fmt.Println("No hooks installed.")
		return nil
	}

	// Confirm before deleting all hooks
	if !confirmAction("Are you sure you want to remove ALL hooks? This action cannot be undone. (y/N): ") {
		fmt.Println("Uninstall cancelled.")
		return nil
	}

	// Remove each hook
	for _, hookPath := range hookFiles {
		err := os.RemoveAll(hookPath)
		if err != nil {
			fmt.Printf("Failed to remove hook '%s': %v\n", filepath.Base(hookPath), err)
		} else {
			fmt.Printf("Removed hook '%s'\n", filepath.Base(hookPath))
		}
	}

	fmt.Println("All hooks have been removed.")
	return nil
}

func uninstallHookType(hooksDir, hookType string) error {
	typeDir := filepath.Join(hooksDir, hookType)
	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		return fmt.Errorf("hook type directory '%s' not found", hookType)
	}

	if !confirmAction(fmt.Sprintf("Are you sure you want to remove all hooks of type '%s'? (y/N): ", hookType)) {
		fmt.Println("Uninstall cancelled.")
		return nil
	}

	if err := os.RemoveAll(typeDir); err != nil {
		return fmt.Errorf("failed to remove hook type directory '%s': %w", hookType, err)
	}

	fmt.Printf("All hooks of type '%s' have been removed.\n", hookType)
	return nil
}

func uninstallSingleHook(hooksDir, hookType, hookID string) error {
	hookPath := filepath.Join(hooksDir, hookType, hookID)
	disabledHookPath := hookPath + ".disabled"

	var targetPath string
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		targetPath = hookPath
	} else if _, err := os.Stat(disabledHookPath); !os.IsNotExist(err) {
		targetPath = disabledHookPath
	} else {
		return fmt.Errorf("hook '%s' of type '%s' not found", hookID, hookType)
	}

	if !confirmAction(fmt.Sprintf("Are you sure you want to remove hook '%s' of type '%s'? (y/N): ", hookID, hookType)) {
		fmt.Println("Uninstall cancelled.")
		return nil
	}

	if err := os.Remove(targetPath); err != nil {
		return fmt.Errorf("failed to remove hook '%s': %w", hookID, err)
	}

	fmt.Printf("Hook '%s' of type '%s' has been removed.\n", hookID, hookType)
	return nil
}

func confirmAction(message string) bool {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
