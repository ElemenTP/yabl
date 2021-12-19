package lib

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"yabl/stack"

	"github.com/gorilla/websocket"
)

var (
	IL map[string]Function //compiled script in IL for interpreter
)

func init() {
	IL = make(map[string]Function)
}

//Field of a function, contains PC pointer, local variables
//, a stack to handle branch and a stack to handle cycle.
type FuncField struct {
	PCp         int               //PC pointer
	BranchStack *stack.Stack      //stack to handle branch
	CycleStack  *stack.Stack      //stack to handle cycle
	LocalVar    map[string]string //local variables
}

//An invoker of yabl functions.
func FuncInvoker(funcName string, params *[]string, conn *websocket.Conn) string {
	n, ok := IL[funcName]
	if !ok {
		InterpreteError(funcName, "can not find a function with name this name.")
	}
	provideLen, neededLen := len(*params), len(n.Params)
	if provideLen > neededLen {
		if neededLen == 0 {
			InterpreteError(funcName, "no param is needed, but params are provided.")
		} else {
			InterpreteWarning(funcName, "more params than needed is provided, ignoring.")
		}
	}
	f := FuncField{0, stack.NewStack(), stack.NewStack(), make(map[string]string)}
	for i, k := range n.Params {
		if i >= provideLen {
			InterpreteError(funcName, "can not fetch the param "+k+".")
		} else {
			f.LocalVar[k] = (*params)[i]
		}
	}
	funcLen := len(n.FuncElem)
	for {
		tOp := &n.FuncElem[f.PCp]
		switch tOp.OT {
		case op_null:
			if tOp.IsAssign {
				switch tOp.OpElem[1].LT {
				case lex_Constant:
					f.LocalVar[tOp.OpElem[0].Content] = tOp.OpElem[1].Content
				case lex_Identifier:
					res, ok := f.LocalVar[tOp.OpElem[1].Content]
					if !ok {
						InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
					} else {
						f.LocalVar[tOp.OpElem[0].Content] = res
					}
				}
			}

		case op_if:
			res := false
			if tOp.IsPC {
				res = len(tOp.PCValue) > 0
			} else {
				a, ok := f.LocalVar[tOp.OpElem[1].Content]
				if !ok {
					InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
				} else {
					res = len(a) > 0
				}
			}
			f.BranchStack.Push(op_if)
			if !res {
				curpos := f.BranchStack.Len()
				f.PCp += 1
			find1:
				for {
					switch n.FuncElem[f.PCp].OT {
					case op_if:
						f.BranchStack.Push(op_if)
					case op_else:
						if f.BranchStack.Len() == curpos {
							break find1
						}
					case op_elif:
						if f.BranchStack.Len() == curpos {
							ress := false
							tOpp := &n.FuncElem[f.PCp]
							if tOpp.IsPC {
								ress = len(tOpp.PCValue) > 0
							} else {
								a, ok := f.LocalVar[tOpp.OpElem[1].Content]
								if !ok {
									InterpreteError(funcName, "can not find variable "+tOpp.OpElem[1].Content)
								} else {
									ress = len(a) > 0
								}
							}
							if ress {
								break find1
							}
						}
					case op_fi:
						f.BranchStack.Pop()
						if f.BranchStack.Len() == curpos-1 {
							break find1
						}
					}
					f.PCp += 1
				}
			}

		case op_else:
			curpos := f.BranchStack.Len()
			f.PCp += 1
		find2:
			for {
				switch n.FuncElem[f.PCp].OT {
				case op_if:
					f.BranchStack.Push(op_if)
				case op_fi:
					f.BranchStack.Pop()
					if f.BranchStack.Len() == curpos-1 {
						break find2
					}
				}
				f.PCp += 1
			}

		case op_elif:
			curpos := f.BranchStack.Len()
			f.PCp += 1
		find3:
			for {
				switch n.FuncElem[f.PCp].OT {
				case op_if:
					f.BranchStack.Push(op_if)
				case op_fi:
					f.BranchStack.Pop()
					if f.BranchStack.Len() == curpos-1 {
						break find3
					}
				}
				f.PCp += 1
			}

		case op_fi:

		case op_loop:
			f.CycleStack.Push(f.PCp)

		case op_pool:
			switch value := f.CycleStack.Peek().(type) {
			case int:
				f.PCp = value
			}

		case op_continue:
			switch value := f.CycleStack.Peek().(type) {
			case int:
				f.PCp = value
			}

		case op_break:
			curpos := f.CycleStack.Len()
			f.PCp += 1
		find4:
			for {
				switch n.FuncElem[f.PCp].OT {
				case op_loop:
					f.CycleStack.Push(op_loop)
				case op_pool:
					f.CycleStack.Pop()
					if f.CycleStack.Len() == curpos-1 {
						break find4
					}
				}
				f.PCp += 1
			}

		case op_return:
			if len(tOp.OpElem) == 1 {
				return ""
			} else {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, ok := f.LocalVar[tOp.OpElem[1].Content]
					if !ok {
						InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
					} else {
						res = a
					}
				}
				return res
			}

		case op_equal:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, b := "", ""
					switch tOp.OpElem[1].LT {
					case lex_Constant:
						a = tOp.OpElem[1].Content
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[1].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
						}
						a = ta
					}
					switch tOp.OpElem[3].LT {
					case lex_Constant:
						b = tOp.OpElem[3].Content
					case lex_Identifier:
						tb, ok := f.LocalVar[tOp.OpElem[3].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
						}
						b = tb
					}
					if a == b {
						res = "true"
					}
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_and:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, b := "", ""
					switch tOp.OpElem[1].LT {
					case lex_Constant:
						a = tOp.OpElem[1].Content
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[1].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
						}
						a = ta
					}
					switch tOp.OpElem[3].LT {
					case lex_Constant:
						b = tOp.OpElem[3].Content
					case lex_Identifier:
						tb, ok := f.LocalVar[tOp.OpElem[3].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
						}
						b = tb
					}
					if len(a) > 0 && len(b) > 0 {
						res = "true"
					}
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_or:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, b := "", ""
					switch tOp.OpElem[1].LT {
					case lex_Constant:
						a = tOp.OpElem[1].Content
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[1].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
						}
						a = ta
					}
					switch tOp.OpElem[3].LT {
					case lex_Constant:
						b = tOp.OpElem[3].Content
					case lex_Identifier:
						tb, ok := f.LocalVar[tOp.OpElem[3].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
						}
						b = tb
					}
					if len(a) > 0 || len(b) > 0 {
						res = "true"
					}
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_not:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a := ""
					switch tOp.OpElem[2].LT {
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[2].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[2].Content)
						}
						a = ta
					}
					if a == "" {
						res = "true"
					}
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_join:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, b := "", ""
					switch tOp.OpElem[1].LT {
					case lex_Constant:
						a = tOp.OpElem[1].Content
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[1].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
						}
						a = ta
					}
					switch tOp.OpElem[3].LT {
					case lex_Constant:
						b = tOp.OpElem[3].Content
					case lex_Identifier:
						tb, ok := f.LocalVar[tOp.OpElem[3].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
						}
						b = tb
					}
					res = a + b
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_contain:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, b := "", ""
					switch tOp.OpElem[1].LT {
					case lex_Constant:
						a = tOp.OpElem[1].Content
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[1].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
						}
						a = ta
					}
					switch tOp.OpElem[3].LT {
					case lex_Constant:
						b = tOp.OpElem[3].Content
					case lex_Identifier:
						tb, ok := f.LocalVar[tOp.OpElem[3].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
						}
						b = tb
					}
					if strings.Contains(a, b) {
						res = "true"
					}
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_hasprefix:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, b := "", ""
					switch tOp.OpElem[1].LT {
					case lex_Constant:
						a = tOp.OpElem[1].Content
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[1].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
						}
						a = ta
					}
					switch tOp.OpElem[3].LT {
					case lex_Constant:
						b = tOp.OpElem[3].Content
					case lex_Identifier:
						tb, ok := f.LocalVar[tOp.OpElem[3].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
						}
						b = tb
					}
					if strings.HasPrefix(a, b) {
						res = "true"
					}
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_hassuffix:
			if tOp.IsAssign {
				res := ""
				if tOp.IsPC {
					res = tOp.PCValue
				} else {
					a, b := "", ""
					switch tOp.OpElem[1].LT {
					case lex_Constant:
						a = tOp.OpElem[1].Content
					case lex_Identifier:
						ta, ok := f.LocalVar[tOp.OpElem[1].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
						}
						a = ta
					}
					switch tOp.OpElem[3].LT {
					case lex_Constant:
						b = tOp.OpElem[3].Content
					case lex_Identifier:
						tb, ok := f.LocalVar[tOp.OpElem[3].Content]
						if !ok {
							InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
						}
						b = tb
					}
					if strings.HasSuffix(a, b) {
						res = "true"
					}
				}
				f.LocalVar[tOp.OpElem[0].Content] = res
			}

		case op_invoke:
			prefix := 0
			if tOp.IsAssign {
				prefix = 1
			}
			paramlen := len(tOp.OpElem) - 2 - prefix
			params := make([]string, 0, paramlen)
			for _, p := range tOp.OpElem[2+prefix:] {
				switch p.LT {
				case lex_Constant:
					params = append(params, p.Content)
				case lex_Identifier:
					res, ok := f.LocalVar[p.Content]
					if !ok {
						InterpreteError(funcName, "can not find variable "+tOp.OpElem[3].Content)
					}
					params = append(params, res)
				}
			}
			value := FuncInvoker(tOp.OpElem[1+prefix].Content, &params, conn)
			if tOp.IsAssign {
				f.LocalVar[tOp.OpElem[0].Content] = value
			}

		case op_getmsg:
			recvmsg := GetMsg(conn)
			if tOp.IsAssign {
				f.LocalVar[tOp.OpElem[0].Content] = recvmsg
			}

		case op_postmsg:
			sendmsg := ""
			switch tOp.OpElem[1].LT {
			case lex_Constant:
				sendmsg = tOp.OpElem[1].Content
			case lex_Identifier:
				res, ok := f.LocalVar[tOp.OpElem[1].Content]
				if !ok {
					InterpreteError(funcName, "can not find variable "+tOp.OpElem[1].Content)
				}
				sendmsg = res
			}
			PostMsg(sendmsg, conn)

		default:
			InterpreteWarning(funcName, "unknown operation, skipping.")
		}
		f.PCp += 1
		if f.PCp == funcLen {
			return ""
		}
	}
}

//The struct of messages sent between server and clients.
type MsgStruct struct {
	Timestamp int64  `json:"timestamp"`
	Content   string `json:"content"`
}

//Fetch message from the client.
func GetMsg(conn *websocket.Conn) string {
	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(600))) //close connection after 600s
	recvmsg := new(MsgStruct)
	err := conn.ReadJSON(&recvmsg)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				InterpreteError("getmsg", "websocket receive message from "+conn.RemoteAddr().String()+" timeout")
			}
		}
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			InterpreteError("getmsg", fmt.Sprintf("websocket receive message from %v error: %v \n", conn.RemoteAddr(), err))
		}
		panic("[Server] Info : client disconnect")
	}
	InterpreteInfo("getmsg", "receive message from "+conn.RemoteAddr().String())
	return recvmsg.Content
}

//Send message to the client.
func PostMsg(content string, conn *websocket.Conn) {
	conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(600))) //close connection after 600s
	sendmsg := new(MsgStruct)
	sendmsg.Timestamp = time.Now().Unix()
	sendmsg.Content = content
	err := conn.WriteJSON(&sendmsg)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				InterpreteError("postmsg", "websocket send message to "+conn.RemoteAddr().String()+" timeout")
			}
		}
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			InterpreteError("postmsg", fmt.Sprintf("websocket send message to %v error: %v \n", conn.RemoteAddr(), err))
		}
		panic("[Server] Info : client disconnect")
	}
	InterpreteInfo("postmsg", "send message to "+conn.RemoteAddr().String())
}

//Show a error message of interpreter and exit.
func InterpreteError(fname, msg string) {
	log.Panicln("[Interpreter] Error in func", fname, ":", msg)
}

//Show a warning message of interpreter.
func InterpreteWarning(fname, msg string) {
	log.Println("[Interpreter] Warning in func", fname, ":", msg)
}

//Show a info message of interpreter.
func InterpreteInfo(fname, msg string) {
	log.Println("[Interpreter] Info in func", fname, ":", msg)
}
