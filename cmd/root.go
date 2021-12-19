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
	rootCmd   = &cobra.Command{
		Use:   "yabl",
		Short: "Yet Another Bot Language interpreter",
		Long:  "A yabl interpreter in go, using websocket as interface.",
		Run: func(cmd *cobra.Command, args []string) {
			flags := cmd.Flags()

			//Exit when no script file was specified.
			if iptScript == "" {
				log.Fatalln("[Server] Error : no script file was specified, existing...")
			}

			//Read scripts from file.
			yamlFile, err := ioutil.ReadFile(iptScript)
			if err != nil {
				log.Fatalln(err)
			}
			err = yaml.Unmarshal(yamlFile, &lib.Script)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("[Server] Info : read script from", iptScript)

			//Compile script.
			lib.Compile()

			//Parse server settings from script and flags.
			var laddr string //listen address
			laddr, err = flags.GetString("address")
			if err != nil {
				log.Println(err)
			}
			if laddr == "" {
				switch value := lib.Script["address"].(type) {
				case string:
					laddr = value
				default:
					laddr = "127.0.0.1"
				}
			}
			var lnet string
			if strings.HasPrefix(laddr, "unix:") {
				laddr = laddr[5:]
				lnet = "unix"
			} else {
				var port string //listen port
				port, err = flags.GetString("port")
				if err != nil {
					log.Println(err)
				}
				if port == "" {
					switch value := lib.Script["port"].(type) {
					case string:
						port = value
					case int:
						port = strconv.Itoa(value)
					default:
						port = "8080"
					}
				}
				laddr = laddr + ":" + port
				lnet = "tcp"
			}

			//Start a websocket server.
			ws := lib.NewWsServer(laddr, lnet)
			ws.Start()
		},
	}
)

//Execute executes the commands.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

//Set flags of the application.
func init() {
	flags := rootCmd.Flags()
	flags.StringVarP(&iptScript, "script", "s", "", "script file path")
	flags.StringP("address", "a", "", "server listen address (default 127.0.0.1)")
	flags.StringP("port", "p", "", "server listen port (default 8080)")
}
