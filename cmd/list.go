package cmd

import (
	"fmt"
	"os"
	"github.com/vjayajv/omnihook/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Git hooks",
	Run: func(cmd *cobra.Command, args []string) {
		hooksDir := utils.ExpandPath(viper.GetString("omni_hooks_dir"))

		if hooksDir == "" {
			fmt.Println("Error: Hooks directory not set. Run 'omnihook configure' first.")
			os.Exit(1)
		}

		files, err := os.ReadDir(hooksDir)
		if err != nil {
			fmt.Printf("Error reading hooks directory: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No hooks installed.")
			return
		}

		fmt.Println("Installed hooks:")
		for _, file := range files {
			if !file.IsDir() {
				fmt.Println("-", file.Name())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
