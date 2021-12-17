package lib

import (
	"log"
	"strings"
)

var (
	Script     map[string]interface{}
	mainExists bool = false
)

func init() {
	Script = make(map[string]interface{})
}

func Compile() {
	for k, v := range Script {
		spiltkey := strings.Fields(k)
		funcName := spiltkey[0] + " " + spiltkey[1]
		if spiltkey[0] == "func" {
			//check if exists one and only one main function.
			if spiltkey[1] == "main" {
				if mainExists {
					compileError(funcName, "duplicate main function.")
				} else {
					mainExists = true
				}
				if len(spiltkey) > 2 {
					compileError(funcName, "main function must not have params.")
				}
			}

			//check if functions are valid.
			switch ifaceslice := v.(type) {
			case []interface{}:
				strslice := make([]string, 0, len(ifaceslice))
				for _, elem := range ifaceslice {
					switch elemTyped := elem.(type) {
					case string:
						strslice = append(strslice, elemTyped)
					default:
						compileError(funcName, "wrong function structure.")
					}
				}

				//compile string script to IL
				tempFuncIL := Function{spiltkey[2:], make([]Operation, 0, len(strslice))}
				for _, s := range strslice {
					spiltstr := strings.Fields(s)
				}
				IL[spiltkey[1]] = tempFuncIL
				compileInfo(funcName, "identified function.")
			default:
				compileError(funcName, "wrong function structure.")
			}
		}
	}
}

//Show a error message of compiler and exit
func compileError(fname, msg string) {
	log.Fatalln("Compiler: Error in", fname, msg)
}

//Show a warning message of compiler
func compileWarning(fname, msg string) {
	log.Println("Compiler: Warning in", fname, msg)
}

//Show a info message of compiler
func compileInfo(fname, msg string) {
	log.Println("Compiler: Info in", fname, msg)
}
