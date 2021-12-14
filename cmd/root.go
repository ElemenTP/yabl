package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	iptScript string
)

func init() {
	flags := rootCmd.Flags()
	flags.StringVarP(&iptScript, "", "", "", "script file path")
	flags.BoolP("tls", "t", false, "enable tls")
	flags.String("cert", "cert.pem", "TLS certificate")
	flags.String("key", "key.pem", "TLS key")
	flags.StringP("address", "a", "127.0.0.1", "address to listen to")
	flags.StringP("port", "p", "8080", "port to listen to")
}

var rootCmd = &cobra.Command{
	Use:   "yabl",
	Short: "Yet Another Bot Language interpreter",
	Long:  "A yabl interpreter in go, using websocket as interface.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test!")
	},
}
