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
	hookTypes := []string{"pre-commit", "prepare-commit-msg", "commit-msg", "pre-push"}
	gitHooksDir := filepath.Join(home, ".git_hooks")

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

	// Ensure hooks sub-directories exist
	for _, hookType := range hookTypes {
		hookTypeDir := filepath.Join(hooksDir, hookType)
		if err := os.MkdirAll(hookTypeDir, 0755); err != nil {
			fmt.Println("Error creating hooks sub-directory:", err)
			return
		}
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

	for _, hookType := range hookTypes {
		hookPath := filepath.Join(gitHooksDir, hookType)
		if err := createGlobalHook(hookPath); err != nil {
			fmt.Println("Error creating global hook: "+hookPath, err)
			return
		}
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

func createGlobalHook(hookPath string) error {
	templateContent := `#!/bin/sh

# Call Omnihook to run managed hooks
if command -v omnihook >/dev/null 2>&1; then
	%s
	if [ $? -ne 0 ]; then
		echo "OmniHook detected an issue. Aborting commit."
		exit 1
	fi
fi

# Also run the repo-local hook if it exists
if [ -f .git/hooks/%s ]; then
	.git/hooks/%s
	if [ $? -ne 0 ]; then
		echo "Repo-local hook failed. Aborting commit."
		exit 1
	fi
fi`

	hookType := filepath.Base(hookPath)
	var omnihookCmd string
	if hookType == "commit-msg" {
		omnihookCmd = "commitmsgfile=$1\n\tcommitmsg=$(cat $commitmsgfile)\n\tomnihook run --commit-msg \"$commitmsg\" --hook-type " + hookType
	} else {
		omnihookCmd = "omnihook run --hook-type " + hookType
	}

	content := fmt.Sprintf(templateContent, omnihookCmd, hookType, hookType)

	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
