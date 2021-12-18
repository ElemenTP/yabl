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
func funcInvoker(funcName string, params *[]string, conn *websocket.Conn) string {
	n, ok := IL[funcName]
	if !ok {
		interpreteError(funcName, "can not find a function with name this name.")
	}
	f := FuncField{0, stack.NewStack(), stack.NewStack(), make(map[string]string)}
	for i, k := range n.params {
		if i >= len(*params) {
			interpreteError(funcName, "can not fetch the param "+k+".")
		} else {
			f.localVar[k] = (*params)[i]
		}
	}
	funcLen := len(n.ops)
	for {
		tOp := &n.ops[f.PCp]
		switch tOp.opType {
		case op_null:
			if tOp.assignment {
				switch tOp.opElem[1].lexType {
				case lex_Constant:
					f.localVar[tOp.opElem[0].content] = tOp.opElem[1].content
				case lex_Identifier:
					res, ok := f.localVar[tOp.opElem[1].content]
					if !ok {
						interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
					} else {
						f.localVar[tOp.opElem[0].content] = res
					}
				}
			}

		case op_if:
			res := false
			if tOp.haspc {
				res = len(tOp.pcValue) > 0
			} else {
				a, ok := f.localVar[tOp.opElem[1].content]
				if !ok {
					interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
				} else {
					res = len(a) > 0
				}
			}
			f.branchStack.Push(op_if)
			if !res {
				curpos := f.branchStack.Len()
				f.PCp += 1
			find1:
				for {
					switch n.ops[f.PCp].opType {
					case op_if:
						f.branchStack.Push(op_if)
					case op_else:
						if f.branchStack.Len() == curpos {
							break find1
						}
					case op_elif:
						if f.branchStack.Len() == curpos {
							ress := false
							tOpp := &n.ops[f.PCp]
							if tOpp.haspc {
								ress = len(tOpp.pcValue) > 0
							} else {
								a, ok := f.localVar[tOpp.opElem[1].content]
								if !ok {
									interpreteError(funcName, "can not find variable "+tOpp.opElem[1].content)
								} else {
									ress = len(a) > 0
								}
							}
							if ress {
								break find1
							}
						}
					case op_fi:
						f.branchStack.Pop()
						if f.branchStack.Len() == curpos-1 {
							break find1
						}
					}
					f.PCp += 1
				}
			}

		case op_else:
			curpos := f.branchStack.Len()
			f.PCp += 1
		find2:
			for {
				switch n.ops[f.PCp].opType {
				case op_if:
					f.branchStack.Push(op_if)
				case op_fi:
					f.branchStack.Pop()
					if f.branchStack.Len() == curpos-1 {
						break find2
					}
				}
				f.PCp += 1
			}

		case op_elif:
			curpos := f.branchStack.Len()
			f.PCp += 1
		find3:
			for {
				switch n.ops[f.PCp].opType {
				case op_if:
					f.branchStack.Push(op_if)
				case op_fi:
					f.branchStack.Pop()
					if f.branchStack.Len() == curpos-1 {
						break find3
					}
				}
				f.PCp += 1
			}

		case op_fi:

		case op_loop:
			f.cycleStack.Push(f.PCp)

		case op_pool:
			switch value := f.cycleStack.Peek().(type) {
			case int:
				f.PCp = value
			}

		case op_continue:
			switch value := f.cycleStack.Peek().(type) {
			case int:
				f.PCp = value
			}

		case op_break:
			curpos := f.cycleStack.Len()
			f.PCp += 1
		find4:
			for {
				switch n.ops[f.PCp].opType {
				case op_loop:
					f.cycleStack.Push(op_loop)
				case op_pool:
					f.cycleStack.Pop()
					if f.cycleStack.Len() == curpos-1 {
						break find4
					}
				}
				f.PCp += 1
			}

		case op_return:
			if len(tOp.opElem) == 1 {
				return ""
			} else {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, ok := f.localVar[tOp.opElem[1].content]
					if !ok {
						interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
					} else {
						res = a
					}
				}
				return res
			}

		case op_equal:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, b := "", ""
					switch tOp.opElem[1].lexType {
					case lex_Constant:
						a = tOp.opElem[1].content
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[1].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
						}
						a = ta
					}
					switch tOp.opElem[3].lexType {
					case lex_Constant:
						b = tOp.opElem[3].content
					case lex_Identifier:
						tb, ok := f.localVar[tOp.opElem[3].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
						}
						b = tb
					}
					if a == b {
						res = "true"
					}
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_and:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, b := "", ""
					switch tOp.opElem[1].lexType {
					case lex_Constant:
						a = tOp.opElem[1].content
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[1].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
						}
						a = ta
					}
					switch tOp.opElem[3].lexType {
					case lex_Constant:
						b = tOp.opElem[3].content
					case lex_Identifier:
						tb, ok := f.localVar[tOp.opElem[3].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
						}
						b = tb
					}
					if len(a) > 0 && len(b) > 0 {
						res = "true"
					}
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_or:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, b := "", ""
					switch tOp.opElem[1].lexType {
					case lex_Constant:
						a = tOp.opElem[1].content
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[1].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
						}
						a = ta
					}
					switch tOp.opElem[3].lexType {
					case lex_Constant:
						b = tOp.opElem[3].content
					case lex_Identifier:
						tb, ok := f.localVar[tOp.opElem[3].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
						}
						b = tb
					}
					if len(a) > 0 || len(b) > 0 {
						res = "true"
					}
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_not:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a := ""
					switch tOp.opElem[2].lexType {
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[2].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[2].content)
						}
						a = ta
					}
					if a == "" {
						res = "true"
					}
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_join:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, b := "", ""
					switch tOp.opElem[1].lexType {
					case lex_Constant:
						a = tOp.opElem[1].content
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[1].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
						}
						a = ta
					}
					switch tOp.opElem[3].lexType {
					case lex_Constant:
						b = tOp.opElem[3].content
					case lex_Identifier:
						tb, ok := f.localVar[tOp.opElem[3].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
						}
						b = tb
					}
					res = a + b
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_contain:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, b := "", ""
					switch tOp.opElem[1].lexType {
					case lex_Constant:
						a = tOp.opElem[1].content
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[1].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
						}
						a = ta
					}
					switch tOp.opElem[3].lexType {
					case lex_Constant:
						b = tOp.opElem[3].content
					case lex_Identifier:
						tb, ok := f.localVar[tOp.opElem[3].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
						}
						b = tb
					}
					if strings.Contains(a, b) {
						res = "true"
					}
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_hasprefix:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, b := "", ""
					switch tOp.opElem[1].lexType {
					case lex_Constant:
						a = tOp.opElem[1].content
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[1].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
						}
						a = ta
					}
					switch tOp.opElem[3].lexType {
					case lex_Constant:
						b = tOp.opElem[3].content
					case lex_Identifier:
						tb, ok := f.localVar[tOp.opElem[3].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
						}
						b = tb
					}
					if strings.HasPrefix(a, b) {
						res = "true"
					}
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_hassuffix:
			if tOp.assignment {
				res := ""
				if tOp.haspc {
					res = tOp.pcValue
				} else {
					a, b := "", ""
					switch tOp.opElem[1].lexType {
					case lex_Constant:
						a = tOp.opElem[1].content
					case lex_Identifier:
						ta, ok := f.localVar[tOp.opElem[1].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
						}
						a = ta
					}
					switch tOp.opElem[3].lexType {
					case lex_Constant:
						b = tOp.opElem[3].content
					case lex_Identifier:
						tb, ok := f.localVar[tOp.opElem[3].content]
						if !ok {
							interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
						}
						b = tb
					}
					if strings.HasSuffix(a, b) {
						res = "true"
					}
				}
				f.localVar[tOp.opElem[0].content] = res
			}

		case op_invoke:
			prefix := 0
			if tOp.assignment {
				prefix = 1
			}
			paramlen := len(tOp.opElem) - 2 - prefix
			params := make([]string, 0, paramlen)
			for _, p := range tOp.opElem[2+prefix:] {
				switch p.lexType {
				case lex_Constant:
					params = append(params, p.content)
				case lex_Identifier:
					res, ok := f.localVar[p.content]
					if !ok {
						interpreteError(funcName, "can not find variable "+tOp.opElem[3].content)
					}
					params = append(params, res)
				}
			}
			value := funcInvoker(tOp.opElem[1+prefix].content, &params, conn)
			if tOp.assignment {
				f.localVar[tOp.opElem[0].content] = value
			}

		case op_getmsg:
			recvmsg := getmsg(conn)
			if tOp.assignment {
				f.localVar[tOp.opElem[0].content] = recvmsg
			}

		case op_postmsg:
			sendmsg := ""
			switch tOp.opElem[1].lexType {
			case lex_Constant:
				sendmsg = tOp.opElem[1].content
			case lex_Identifier:
				res, ok := f.localVar[tOp.opElem[1].content]
				if !ok {
					interpreteError(funcName, "can not find variable "+tOp.opElem[1].content)
				}
				sendmsg = res
			}
			postmsg(sendmsg, conn)

		default:
			interpreteWarning(funcName, "unknown operation, skipping.")
		}
		f.PCp += 1
		if f.PCp == funcLen {
			return ""
		}
	}
}

type MsgStruct struct {
	Timestamp int64  `json:"timestamp"`
	Content   string `json:"content"`
}

func getmsg(conn *websocket.Conn) string {
	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(600))) //close connection after 600s
	recvmsg := new(MsgStruct)
	err := conn.ReadJSON(&recvmsg)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				interpreteError("getmsg", "websocket receive message from "+conn.RemoteAddr().String()+" timeout")
			}
		}
		interpreteError("getmsg", fmt.Sprintf("websocket receive message from %v error: %v \n", conn.RemoteAddr(), err))
	}
	interpreteInfo("getmsg", "receive message from "+conn.RemoteAddr().String())
	return recvmsg.Content
}

func postmsg(content string, conn *websocket.Conn) {
	conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(600))) //close connection after 600s
	sendmsg := new(MsgStruct)
	sendmsg.Timestamp = time.Now().Unix()
	sendmsg.Content = content
	err := conn.WriteJSON(&sendmsg)
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				interpreteError("postmsg", "websocket send message to "+conn.RemoteAddr().String()+" timeout")
			}
		}
		interpreteError("postmsg", fmt.Sprintf("websocket send message to %v error: %v \n", conn.RemoteAddr(), err))
	}
	interpreteInfo("postmsg", "send message to "+conn.RemoteAddr().String())
}

//Show a error message of interpreter and exit
func interpreteError(fname, msg string) {
	log.Panicln("[Interpreter] Error in func", fname, ":", msg)
}

//Show a warning message of interpreter
func interpreteWarning(fname, msg string) {
	log.Println("[Interpreter] Warning in func", fname, ":", msg)
}

//Show a info message of interpreter
func interpreteInfo(fname, msg string) {
	log.Println("[Interpreter] Info in func", fname, ":", msg)
}
