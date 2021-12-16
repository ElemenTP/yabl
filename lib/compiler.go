package lib

import (
	"log"
	"strings"
)

type LexType int

const (
	lex_Variable = iota
	lex_Function
	lex_Keyword
	lex_Constant
)

type LexElem struct {
	lexType LexType
	content string
}
type OpType int

const (
	op_if = iota
	op_fi
	op_loop
	op_pool
	op_invoke
	op_getmsg
	op_postmsg
	op_and
	op_or
	op_not
	op_join
)

type Operation struct {
	opType OpType
	opElem []LexElem
}

var (
	Script     map[string]interface{}
	mainExists bool = false
)

func init() {
	Script = make(map[string]interface{})
}

func Compile() {
	funcTypoCheck()
}

//check functions if exists one and only one main function and if functions are valid.
func funcTypoCheck() {
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
				funcMapStr[k] = strslice
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
