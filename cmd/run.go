package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"
	"github.com/lianggaoqiang/progress"
	"github.com/jwalton/gchalk"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all installed hooks in parallel",
	RunE:  runHooks,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runHooks(cmd *cobra.Command, args []string) error {
	hooksDir := viper.GetString("omni_hooks_dir")
	if hooksDir == "" {
		return errors.New("hooks directory not set. Run 'omnihook configure' first")
	}

	hookFiles, err := filepath.Glob(filepath.Join(hooksDir, "*"))
	if err != nil {
		return fmt.Errorf("failed to list hooks: %w", err)
	}

	// Filter out .disabled hooks
	var activeHooks []string
	for _, hookPath := range hookFiles {
		if strings.HasSuffix(hookPath, ".disabled") {
			continue // Skip disabled hooks
		}
		activeHooks = append(activeHooks, hookPath)
	}

	if len(activeHooks) == 0 {
		fmt.Println("No active hooks found.")
		return nil
	}

	var wg sync.WaitGroup
	type HookResult struct {
		name     string
		status   string
		errorMsg string
	}
	results := make(chan HookResult, len(activeHooks))

	// Find max hook name length for alignment
	maxHookNameLength := 0
	hookNames := make([]string, len(activeHooks))
	for i, hookPath := range activeHooks {
		hookNames[i] = filepath.Base(hookPath)
		if len(hookNames[i]) > maxHookNameLength {
			maxHookNameLength = len(hookNames[i])
		}
	}

	// Create a progress manager
	p := progress.Start()

	// Create progress bars for each hook
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

	// Start running hooks in parallel
	for _, hookPath := range activeHooks {
		hookName := filepath.Base(hookPath)
		wg.Add(1)

		go func(hookName, hookPath string) {
			defer wg.Done()

			cmd := exec.Command(hookPath)
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

	// Wait for all hooks to finish
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
