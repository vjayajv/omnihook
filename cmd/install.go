package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"time"
	"os/exec"
	"path/filepath"
	"github.com/spf13/viper"
	"github.com/lianggaoqiang/progress"
	"slices"
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

		err := installHook(url, filePath)
		if err == nil && url != "" {
			updateCache(url)
		}
		return err
	},
}

type Hook struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Script      string `yaml:"script"`
	ScriptPath  string `yaml:"scriptPath"`
	HookType    string `yaml:"hookType"`
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

	maxHookNameLength := 0
	hookNames := make([]string, len(hooks))
	for i, hook := range hooks {
		hookNames[i] = hook.ID
		if len(hookNames[i]) > maxHookNameLength {
			maxHookNameLength = len(hookNames[i])
		}
	}

	// Create a progress manager
	p := progress.Start()

	// Create progress bars for each hook
	bars := make(map[string]*progress.DefaultBar)
	for _, hookName := range hookNames {
		paddedHookName := fmt.Sprintf("🪝 Installing hook %-*s ", maxHookNameLength, hookName)
		bars[hookName] = progress.NewBar().Custom(progress.BarSetting{
			Total:          15,
			StartText:      paddedHookName,
			EndText:        " ✅",
			NotPassedText:  progress.BlackText("▇"),
			PassedText:     progress.WhiteText("▇"),
		})
		p.AddBar(bars[hookName])
	}

	for _, hook := range hooks {
		if err := validateHook(hook); err != nil {
			for i := 0; i < 100; i += 20 {
				bars[hook.ID].Show()
				bars[hook.ID].Inc()
				bars[hook.ID].Add(1.4)
				bars[hook.ID].Setting.EndText = " ❌"
				bars[hook.ID].Percent(float64(i))
				bars[hook.ID].Hide()
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("invalid hook configuration: %w", err)
		}

		if hook.HookType == "" {
			hook.HookType = "pre-commit"
		}
		hookTypeDir := filepath.Join(hooksDir, hook.HookType)
		if err := os.MkdirAll(hookTypeDir, 0755); err != nil {
			return fmt.Errorf("failed to create hook type directory: %w", err)
		}
		hookFilePath := filepath.Join(hookTypeDir, hook.ID)
		content := fmt.Sprintf("#!/bin/sh\n# %s\n", hook.Description)
		if hook.Script != "" {
			content += fmt.Sprintf("%s\n", hook.Script)
		} else {
			content += fmt.Sprintf("exec %s \"$@\"\n", hook.ScriptPath)
		}


		if err := os.WriteFile(hookFilePath, []byte(content), 0755); err != nil {
			for i := 0; i < 100; i += 20 {
				bars[hook.ID].Show()
				bars[hook.ID].Inc()
				bars[hook.ID].Add(1.4)
				bars[hook.ID].Setting.EndText = " ❌"
				bars[hook.ID].Percent(float64(i))
				bars[hook.ID].Hide()
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("failed to write hook file: %w", err)
		}

		// Explicitly set executable permissions
		if err := os.Chmod(hookFilePath, 0755); err != nil {
			for i := 0; i < 100; i += 20 {
				bars[hook.ID].Show()
				bars[hook.ID].Inc()
				bars[hook.ID].Add(1.4)
				bars[hook.ID].Setting.EndText = " ❌"
				bars[hook.ID].Percent(float64(i))
				bars[hook.ID].Hide()
				time.Sleep(100 * time.Millisecond)
			}
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}

		for i := 0; i < 100; i += 20 {
			bars[hook.ID].Show()
			bars[hook.ID].Inc()
			bars[hook.ID].Add(1.4)
			bars[hook.ID].Setting.EndText = " ✅"
			bars[hook.ID].Percent(float64(i))
			bars[hook.ID].Hide()
			time.Sleep(100 * time.Millisecond)
		}
		bars[hook.ID].Show()
		bars[hook.ID].Percent(100)
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
		if filepath.Base(path) == "omnihook.yml" {
			loadedHooks, err := loadHooksFromFile(path)
			if err != nil {
				return err // Return immediately if a hook file has errors
			}
			for i, hook := range loadedHooks {
				if hook.ScriptPath != "" {
					scriptFullPath := filepath.Join(filepath.Dir(path), hook.ScriptPath)
					scriptContent, err := os.ReadFile(scriptFullPath)
					if err != nil {
						return fmt.Errorf("failed to read script file %s: %w", scriptFullPath, err)
					}
					loadedHooks[i].Script = string(scriptContent)
					loadedHooks[i].ScriptPath = ""
				}
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
	if hook.HookType == "" {
		hook.HookType = "pre-commit"
	}
	if hook.Name == "" {
		return errors.New("hook name is required")
	}
	if hook.Description == "" {
		return errors.New("hook description is required")
	}
	if hook.Script == "" && hook.ScriptPath == "" {
		return errors.New("either script or scriptPath must be provided")
	}
	if hook.Script != "" && hook.ScriptPath != "" {
		return errors.New("hook cannot have both script and scriptPath")
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

func getCacheFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "cache.yml" // Fallback to current directory
	}
	return filepath.Join(homeDir, ".omnihook", "cache.yml")
}

func readCache() (Cache, error) {
	cacheFile := getCacheFilePath()
	cache := Cache{}

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return cache, nil // Return empty cache if file doesn't exist
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return cache, fmt.Errorf("failed to read cache file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cache); err != nil {
		return cache, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return cache, nil
}

func updateCache(url string) error {
	cache, err := readCache()
	if err != nil {
		return err
	}

	if slices.Contains(cache.Sources, url) {
			return nil // URL already exists, no need to update
	}

	cache.Sources = append(cache.Sources, url)
	return writeCache(cache)
}

func writeCache(cache Cache) error {
	cacheFile := getCacheFilePath()
	err := os.MkdirAll(filepath.Dir(cacheFile), 0755)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := yaml.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to serialize cache: %w", err)
	}

	err = os.WriteFile(cacheFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

func init() {
	installCmd.Flags().String("url", "", "Git repository URL of the hooks")
	installCmd.Flags().String("file", "", "Path to the local hook configuration file")
	rootCmd.AddCommand(installCmd)
}