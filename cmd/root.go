package cmd

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"yabl/lib"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	iptScript string
)

//Set flags of the application.
func init() {
	flags := rootCmd.Flags()
	flags.StringVarP(&iptScript, "script", "s", "", "script file path")
	flags.StringP("address", "a", "127.0.0.1", "address to listen to")
	flags.StringP("port", "p", "8080", "port to listen to")
}

var rootCmd = &cobra.Command{
	Use:   "yabl",
	Short: "Yet Another Bot Language interpreter",
	Long:  "A yabl interpreter in go, using websocket as interface.",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()

		//exit when no script file was specified.
		if iptScript == "" {
			log.Fatalln("No script file was specified, existing...")
		}

		//read scripts from file
		yamlFile, err := ioutil.ReadFile(iptScript)
		if err != nil {
			log.Fatalln(err)
		}
		err = yaml.Unmarshal(yamlFile, &lib.Script)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Read script from ", iptScript)

		//parse server settings from script and flags.
		var laddr string //listen address
		switch value := lib.Script["address"].(type) {
		case string:
			laddr = value

		default:
			laddr, err = flags.GetString("address")
			if err != nil {
				log.Fatalln(err)
			}
		}
		var lnet string
		if strings.HasPrefix(laddr, "unix:") {
			laddr = laddr[5:]
			lnet = "unix"
		} else {
			var port string //listen port
			switch value := lib.Script["port"].(type) {
			case string:
				port = value
			case int:
				port = strconv.Itoa(value)
			default:
				port, err = flags.GetString("port")
				if err != nil {
					log.Fatalln(err)
				}
			}
			laddr = laddr + ":" + port
			lnet = "tcp"
		}

		//start a websocket server
		ws := lib.NewWsServer(laddr, lnet)
		ws.Start()
	},
}
