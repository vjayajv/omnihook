package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var disableCmd = &cobra.Command{
	Use:   "disable --id <hook_id> --type <hook_type>",
	Short: "Disable a specified hook",
	RunE:  disableHook,
}

func init() {
	rootCmd.AddCommand(disableCmd)
	disableCmd.Flags().String("id", "", "ID of the hook to disable")
	disableCmd.Flags().String("type", "", "Type of the hook to disable")
	disableCmd.MarkFlagRequired("id")
	disableCmd.MarkFlagRequired("type")
}

func disableHook(cmd *cobra.Command, args []string) error {
	hooksDir := viper.GetString("omni_hooks_dir")
	if hooksDir == "" {
		return errors.New("hooks directory not set. Run 'omnihook configure' first")
	}

	hookID, _ := cmd.Flags().GetString("id")
	hookType, _ := cmd.Flags().GetString("type")
	hookPath := filepath.Join(hooksDir, hookType, hookID)
	disabledHookPath := hookPath + ".disabled"
	
	if _, err := os.Stat(disabledHookPath); err == nil {
		fmt.Printf("Hook '%s' is already disabled.\n", hookID)
		return nil
	}
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return fmt.Errorf("hook '%s' not found", hookID)
	}

	if err := os.Rename(hookPath, disabledHookPath); err != nil {
		return fmt.Errorf("failed to disable hook '%s': %w", hookID, err)
	}

	fmt.Printf("Hook '%s' has been disabled.\n", hookID)
	return nil
}
