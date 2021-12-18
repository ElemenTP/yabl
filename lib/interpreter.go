package lib

import (
	"log"
	"yabl/stack"

	"github.com/gorilla/websocket"
)

var (
	IL map[string]Function
)

func init() {
	IL = make(map[string]Function)
}

type FuncField struct {
	PCp         int               //PC pointer
	branchStack *stack.Stack      //stack to handle branch
	cycleStack  *stack.Stack      //stack to handle cycle
	localVar    map[string]string //local variables
}

//a invoker of yabl functions
func funcInvoker(funcName string, params *map[string]string, conn *websocket.Conn) {
	n, ok := IL[funcName]
	if !ok {
		interpreteError(funcName, "can not find a function with name this name.")
	}
	f := FuncField{0, stack.NewStack(), stack.NewStack(), make(map[string]string)}
	for _, k := range n.params {
		v, ok := (*params)[k]
		if !ok {
			interpreteError(funcName, "can not fetch the param "+k+".")
		}
		f.localVar[k] = v
	}

}

//Show a error message of compiler and exit
func interpreteError(fname, msg string) {
	log.Fatalln("[Interpreter] Error in func", fname, ":", msg)
}

//Show a warning message of compiler
func interpreteWarning(fname, msg string) {
	log.Println("[Interpreter] Warning in func", fname, ":", msg)
}

//Show a info message of compiler
func interpreteInfo(fname, msg string) {
	log.Println("[Interpreter] Info in func", fname, ":", msg)
}
