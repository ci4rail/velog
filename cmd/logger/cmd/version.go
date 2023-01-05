package cmd

import (
	"fmt"

	"github.com/ci4rail/velog/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information and quit",
	Long:  `Print version information and quit`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version: %s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
