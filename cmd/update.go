package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

type Cache struct {
	Sources []string `yaml:"sources"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update installed hooks",
	RunE:  updateHooks,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("url", "", "Update hooks from a specific source URL")
	updateCmd.Flags().Bool("all", false, "Update all sources")
}

func updateHooks(cmd *cobra.Command, args []string) error {
	url, _ := cmd.Flags().GetString("url")
	all, _ := cmd.Flags().GetBool("all")

	cache, err := readCache()
	if err != nil {
		return err
	}

	if url != "" {
		return reinstallHooks([]string{url})
	} else if all || len(args) == 0 {
		if cache.Sources == nil {
			return errors.New("no sources to update, use --url instead")
		}
		return reinstallHooks(cache.Sources)
	}

	return errors.New("invalid update parameters")
}

func reinstallHooks(sources []string) error {
	for _, src := range sources {
		fmt.Printf("Updating hooks from: %s\n", src)
		// Call install logic for each stored source
		installHook(src, "")
	}
	return nil
}
