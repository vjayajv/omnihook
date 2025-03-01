package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var enableCmd = &cobra.Command{
	Use:   "enable --id <hook_id>",
	Short: "Enable a previously disabled hook",
	RunE:  enableHook,
}

func init() {
	rootCmd.AddCommand(enableCmd)
	enableCmd.Flags().String("id", "", "ID of the hook to enable")
	enableCmd.MarkFlagRequired("id")
}

func enableHook(cmd *cobra.Command, args []string) error {
	hooksDir := viper.GetString("omni_hooks_dir")
	if hooksDir == "" {
		return fmt.Errorf("hooks directory not set. Run 'omnihook configure' first")
	}

	hookID, _ := cmd.Flags().GetString("id")
	disabledHookPath := filepath.Join(hooksDir, hookID+".disabled")
	enabledHookPath := filepath.Join(hooksDir, hookID)

	// Check if the hook is actually disabled
	if _, err := os.Stat(disabledHookPath); os.IsNotExist(err) {
		if _, err := os.Stat(enabledHookPath); err == nil {
			fmt.Printf("Hook '%s' is already enabled.\n", hookID)
			return nil
		}
		return fmt.Errorf("hook '%s' does not exist", hookID)
	}

	// Rename the disabled hook back to enabled
	err := os.Rename(disabledHookPath, enabledHookPath)
	if err != nil {
		return fmt.Errorf("failed to enable hook '%s': %w", hookID, err)
	}

	fmt.Printf("Hook '%s' has been enabled.\n", hookID)
	return nil
}
