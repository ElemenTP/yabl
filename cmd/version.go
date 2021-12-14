package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "unknown version"
	BuildTime = "unknown time"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Clash %s %s %s with %s %s\n", Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), BuildTime)
		},
	})
}
