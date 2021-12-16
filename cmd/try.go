package cmd

import (
	"io/ioutil"
	"log"
	"yabl/lib"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

//Add version command into the cli app, will show version info and compile time.
func init() {
	flags := tryCmd.Flags()
	flags.StringVarP(&iptScript, "script", "s", "", "script file path")
	rootCmd.AddCommand(tryCmd)
}

var tryCmd = &cobra.Command{
	Use:   "try",
	Short: "Check script validity",
	Long:  "Try to compile the script, check its validity",
	Run: func(cmd *cobra.Command, args []string) {
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

		//compile script
		lib.Compile()
	},
}
