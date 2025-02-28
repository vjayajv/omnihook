package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"github.com/spf13/viper"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a new hook from a Git repository or file",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		filePath, _ := cmd.Flags().GetString("file")

		if url == "" && filePath == "" {
			return errors.New("either --url or --file must be provided")
		}
		if url != "" && filePath != "" {
			return errors.New("cannot use both --url and --file at the same time")
		}

		return installHook(url, filePath)
	},
}

func init() {
	installCmd.Flags().String("url", "", "Git repository URL of the hooks")
	installCmd.Flags().String("file", "", "Path to the local hook configuration file")
	rootCmd.AddCommand(installCmd)
}

type Hook struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Script      string `yaml:"script,omitempty"`
	ScriptPath  string `yaml:"script_path,omitempty"`
}

type OmniHook struct {
	Hooks []Hook `yaml:"hooks"`
}

func installHook(url, filePath string) error {
	hooksDir := getHooksDir()
	if hooksDir == "" {
		return errors.New("hooks directory not set. Run 'omnihook configure' first")
	}

	var hooks []Hook
	var err error

	if url != "" {
		hooks, err = fetchHooksFromGitRepo(url)
	} else {
		hooks, err = loadHooksFromFile(filePath)
	}

	if err != nil {
		return fmt.Errorf("failed to load hooks: %w", err)
	}

	for _, hook := range hooks {
		if err := validateHook(hook); err != nil {
			return fmt.Errorf("invalid hook configuration: %w", err)
		}

		hookFilePath := filepath.Join(hooksDir, hook.ID)
		content := fmt.Sprintf("#!/bin/sh\n# %s\n", hook.Description)
		if hook.Script != "" {
			content += fmt.Sprintf("%s\n", hook.Script)
		} else {
			content += fmt.Sprintf("exec %s \"$@\"\n", hook.ScriptPath)
		}

		if err := os.WriteFile(hookFilePath, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to write hook file: %w", err)
		}

		// Explicitly set executable permissions
		if err := os.Chmod(hookFilePath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	return nil
}

func fetchHooksFromGitRepo(repoURL string) ([]Hook, error) {
	tempDir, err := os.MkdirTemp("", "omnihook-clone-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("git", "clone", repoURL, tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %s: %w", string(output), err)
	}

	var hooks []Hook
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "hook.yml" || filepath.Base(path) == "omnihook.yml" {
			loadedHooks, err := loadHooksFromFile(path)
			if err != nil {
				return err // Return immediately if a hook file has errors
			}
			hooks = append(hooks, loadedHooks...)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning repository files: %w", err)
	}

	if len(hooks) == 0 {
		return nil, errors.New("no valid hook configurations found in repository")
	}

	return hooks, nil
}

func loadHooksFromFile(filePath string) ([]Hook, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try parsing as OmniHook first
	var omniHook OmniHook
	if err := yaml.Unmarshal(data, &omniHook); err == nil {
		if len(omniHook.Hooks) > 0 {
			return omniHook.Hooks, nil
		}
	}

	// Reset error and try parsing as single Hook
	var hook Hook
	if err := yaml.Unmarshal(data, &hook); err == nil && hook.ID != "" {
		return []Hook{hook}, nil
	}

	return nil, errors.New("failed to parse YAML: file may be improperly formatted")
}

func validateHook(hook Hook) error {
	if hook.ID == "" {
		return errors.New("hook ID is required")
	}
	if hook.Name == "" {
		return errors.New("hook name is required")
	}
	if hook.Description == "" {
		return errors.New("hook description is required")
	}
	if hook.Script == "" && hook.ScriptPath == "" {
		return errors.New("either script or script_path must be provided")
	}
	if hook.Script != "" && hook.ScriptPath != "" {
		return errors.New("hook cannot have both script and script_path")
	}
	return nil
}

func getHooksDir() string {
	hooksDir := viper.GetString("omni_hooks_dir")
	if hooksDir == "" {
		fmt.Println("Hooks directory is not configured. Run 'omnihook configure' first.")
		return ""
	}

	// Ensure the hooks directory exists
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		fmt.Printf("Hooks directory '%s' does not exist. Run 'omnihook configure' to set it up.\n", hooksDir)
		return ""
	}

	return hooksDir
}

func init() {
	rootCmd.AddCommand(installCmd)
}
