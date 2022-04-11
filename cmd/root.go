package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"yabl/lib"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	rootCmd = &cobra.Command{
		Use:   "yabl [script]",
		Short: "Yet Another Bot Language interpreter",
		Long:  "A yabl interpreter in go, using websocket as interface.",
		Args:  scriptArg,
		Run: func(cmd *cobra.Command, args []string) {
			flags := cmd.Flags()

			//Read scripts from file.
			yamlFile, err := ioutil.ReadFile(args[0])
			if err != nil {
				log.Fatalln(err)
			}
			err = yaml.Unmarshal(yamlFile, &lib.Script)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("[Server] Info : read script from", args[0])

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
			go func() {
				ws := lib.NewWsServer(laddr, lnet)
				ws.Start()
			}()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			<-sigCh
			fmt.Println("Got interrupt message, existing...")
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
	flags.StringP("address", "a", "", "server listen address (default 127.0.0.1)")
	flags.StringP("port", "p", "", "server listen port (default 8080)")
}

func scriptArg(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("[Server] Error : no script file was specified")
	}
	if len(args) > 1 {
		return fmt.Errorf("[Server] Error : more than one script file was specified")
	}
	return nil
}
