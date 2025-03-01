package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Omnihook by setting up the global hooks directory",
	Run: func(cmd *cobra.Command, args []string) {
		reset, _ := cmd.Flags().GetBool("reset")
		configureOmnihook(reset)
	},
}

func init() {
	configureCmd.Flags().Bool("reset", false, "Reset the Omnihook configuration")
	rootCmd.AddCommand(configureCmd)
}

func configureOmnihook(reset bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	configDir := filepath.Join(home, ".omnihook")
	configFile := filepath.Join(configDir, "config.yaml")
	hooksDir := filepath.Join(configDir, "hooks")
	gitHooksDir := filepath.Join(home, ".git_hooks")
	preCommitHookPath := filepath.Join(gitHooksDir, "pre-commit")

	if reset {
		if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
			fmt.Println("Error resetting configuration:", err)
			return
		}
		fmt.Println("Omnihook configuration reset.")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("Error creating config directory:", err)
		return
	}

	// Ensure hooks directory exists
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Println("Error creating hooks directory:", err)
		return
	}

	// Ensure Git hooks directory exists
	if err := os.MkdirAll(gitHooksDir, 0755); err != nil {
		fmt.Println("Error creating Git hooks directory:", err)
		return
	}

	// Set global Git hooks path
	if err := setGitHooksPath(gitHooksDir); err != nil {
		fmt.Println("Error setting Git hooks path:", err)
		return
	}

	// Create the pre-commit hook
	if err := createPreCommitHook(preCommitHookPath); err != nil {
		fmt.Println("Error creating pre-commit hook:", err)
		return
	}

	// Save config
	viper.Set("omni_hooks_dir", hooksDir)
	viper.SetConfigFile(configFile)
	if err := viper.WriteConfig(); err != nil {
		fmt.Println("Error writing config file:", err)
		return
	}

	fmt.Println("Omnihook configured successfully.")
}

func setGitHooksPath(gitHooksDir string) error {
	cmd := exec.Command("git", "config", "--global", "core.hooksPath", gitHooksDir)
	return cmd.Run()
}

func createPreCommitHook(preCommitHookPath string) error {
	content := `#!/bin/sh

# Call Omnihook to run managed hooks
if command -v omnihook >/dev/null 2>&1; then
    omnihook run
    if [ $? -ne 0 ]; then
        echo "OmniHook detected an issue. Aborting commit."
        exit 1
    fi
fi

# Also run the repo-local pre-commit if it exists
if [ -f .git/hooks/pre-commit ]; then
    .git/hooks/pre-commit
    if [ $? -ne 0 ]; then
        echo "Repo-local pre-commit hook failed. Aborting commit."
        exit 1
    fi
fi`

	if err := os.WriteFile(preCommitHookPath, []byte(content), 0755); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
