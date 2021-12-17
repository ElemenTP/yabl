package lib

import "log"

var (
	IL map[string]Function
)

func init() {
	IL = make(map[string]Function)
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
