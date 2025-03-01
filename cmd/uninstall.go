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
}

func uninstallHook(cmd *cobra.Command, args []string) error {
	hooksDir := viper.GetString("omni_hooks_dir")
	if hooksDir == "" {
		return fmt.Errorf("hooks directory not set. Run 'omnihook configure' first")
	}

	hookID, _ := cmd.Flags().GetString("id")
	removeAll, _ := cmd.Flags().GetBool("all")

	if !removeAll && hookID == "" {
		return fmt.Errorf("either --id or --all must be specified")
	}

	if removeAll {
		return uninstallAllHooks(hooksDir)
	} else {
		return uninstallSingleHook(hooksDir, hookID)
	}
}

func uninstallSingleHook(hooksDir, hookID string) error {
	hookPath := filepath.Join(hooksDir, hookID)
	disabledHookPath := hookPath + ".disabled"

	// Determine if it's enabled or disabled
	var targetPath string
	if _, err := os.Stat(hookPath); err == nil {
		targetPath = hookPath
	} else if _, err := os.Stat(disabledHookPath); err == nil {
		targetPath = disabledHookPath
	} else {
		return fmt.Errorf("hook '%s' not found", hookID)
	}

	// Confirm before deletion
	if !confirmAction(fmt.Sprintf("Are you sure you want to remove hook '%s'? (y/N): ", hookID)) {
		fmt.Println("Uninstall cancelled.")
		return nil
	}

	// Remove the hook
	err := os.Remove(targetPath)
	if err != nil {
		return fmt.Errorf("failed to remove hook '%s': %w", hookID, err)
	}

	fmt.Printf("Hook '%s' has been removed.\n", hookID)
	return nil
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
		err := os.Remove(hookPath)
		if err != nil {
			fmt.Printf("Failed to remove hook '%s': %v\n", filepath.Base(hookPath), err)
		} else {
			fmt.Printf("Removed hook '%s'\n", filepath.Base(hookPath))
		}
	}

	fmt.Println("All hooks have been removed.")
	return nil
}

func confirmAction(message string) bool {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
