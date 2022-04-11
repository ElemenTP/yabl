package cmd

import (
	"io/ioutil"
	"log"
	"yabl/lib"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//Add version command into the cli app, will show version info and compile time.
func init() {
	rootCmd.AddCommand(tryCmd)
}

var tryCmd = &cobra.Command{
	Use:   "try [script]",
	Short: "Check script validity",
	Long:  "Try to compile the script, check its validity",
	Args:  scriptArg,
	Run: func(_ *cobra.Command, args []string) {
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

		log.Println("[Server] Info : script successfully compiled")
	},
}
