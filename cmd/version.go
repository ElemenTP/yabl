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

//Add version command into the cli app, will show version info and compile time.
func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("yabl %s %s %s with %s %s\n", Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), BuildTime)
		},
	})
}
