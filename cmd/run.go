package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"os"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"
	"github.com/lianggaoqiang/progress"
	"github.com/jwalton/gchalk"
)

var runCmd = &cobra.Command{
	Use:   "run --all or --type <hook_type>",
	Short: "Run all installed hooks in parallel",
	RunE:  runHooks,
}

func init() {
	runCmd.Flags().String("commit-msg", "", "Commit message passed from git commit")
	runCmd.Flags().Bool("all", false, "Run all installed hooks")
	runCmd.Flags().String("type", "", "Run all installed hooks of a specific type")
	rootCmd.AddCommand(runCmd)
}

func runHooks(cmd *cobra.Command, args []string) error {
	hookType, _ := cmd.Flags().GetString("type")
	all, _ := cmd.Flags().GetBool("all")
	if !all && hookType == "" {
		return errors.New("must specify either --all or --type")
	}
	commitMsg, _ := cmd.Flags().GetString("commit-msg")

	hooksDir := viper.GetString("omni_hooks_dir")
	if hooksDir == "" {
		return errors.New("hooks directory not set. Run 'omnihook configure' first")
	}

	var hookFiles []string
	var err error
	
	if all {
		// Get all hooks from all subdirectories
		hookFiles, err = filepath.Glob(filepath.Join(hooksDir, "**/*"))
	} else {
		// Get hooks only from specific type subdirectory
		hookFiles, err = filepath.Glob(filepath.Join(hooksDir, hookType, "*"))
	}
	
	if err != nil {
		return fmt.Errorf("failed to list hooks: %w", err)
	}

	// Filter out .disabled hooks and directories
	var activeHooks []string
	for _, hookPath := range hookFiles {
		if strings.HasSuffix(hookPath, ".disabled") {
			continue // Skip disabled hooks
		}
		// Skip if it's a directory
		if info, err := os.Stat(hookPath); err == nil && info.IsDir() {
			continue
		}
		activeHooks = append(activeHooks, hookPath)
	}

	if len(activeHooks) == 0 {
		fmt.Println("No active hooks found.")
		return nil
	}

	// Rest of the code remains the same...
	var wg sync.WaitGroup
	type HookResult struct {
		name     string
		status   string
		errorMsg string
	}
	results := make(chan HookResult, len(activeHooks))

	maxHookNameLength := 0
	hookNames := make([]string, len(activeHooks))
	for i, hookPath := range activeHooks {
		hookNames[i] = filepath.Base(hookPath)
		if len(hookNames[i]) > maxHookNameLength {
			maxHookNameLength = len(hookNames[i])
		}
	}

	p := progress.Start()

	bars := make(map[string]*progress.DefaultBar)
	for _, hookPath := range activeHooks {
		hookName := filepath.Base(hookPath)
		paddedHookName := fmt.Sprintf("🪝 %-*s ", maxHookNameLength, hookName)
		bars[hookName] = progress.NewBar().Custom(progress.BarSetting{
			Total:          15,
			StartText:      paddedHookName,
			EndText:        " ✅",
			NotPassedText:  progress.BlackText("▇"),
			PassedText:     progress.WhiteText("▇"),
		})
		p.AddBar(bars[hookName])
	}

	for _, hookPath := range activeHooks {
		hookName := filepath.Base(hookPath)
		wg.Add(1)

		go func(hookName, hookPath string) {
			defer wg.Done()

			var cmdArgs []string
			if hookType == "commit-msg" && commitMsg != "" {
				cmdArgs = append(cmdArgs, commitMsg)
			}
			cmd := exec.Command(hookPath, cmdArgs...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				for i := 0; i < 100; i += 20 {
					bars[hookName].Show()
					bars[hookName].Inc()
					bars[hookName].Add(1.4)
					bars[hookName].Setting.EndText = " ❌"
					bars[hookName].Percent(float64(i))
					bars[hookName].Hide()
					time.Sleep(100 * time.Millisecond)
				}
				results <- HookResult{hookName, "Failed", string(output)}
			} else {
				for i := 0; i < 100; i += 20 {
					bars[hookName].Show()
					bars[hookName].Inc()
					bars[hookName].Add(1.4)
					bars[hookName].Setting.EndText = " ✅"
					bars[hookName].Percent(float64(i))
					bars[hookName].Hide()
					time.Sleep(100 * time.Millisecond)
				}
				results <- HookResult{hookName, "Ok", ""}
			}
			bars[hookName].Show()
			bars[hookName].Percent(100)
		}(hookName, hookPath)
	}

	wg.Wait()
	close(results)

	fmt.Println()
	failureCount := 0
	for result := range results {
		if result.status == "Failed" {
			fmt.Printf("\n🚧 %s check failed:\n%s\n\n", gchalk.Bold(result.name), gchalk.Red(result.errorMsg))
			failureCount++
		}
	}
	
	if failureCount > 0 {
		cmd.SilenceUsage = true
		return errors.New("one or more pre-commit checks failed")
	}

	return nil
}
