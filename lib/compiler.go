package lib

import (
	"log"
	"strings"
	"yabl/stack"
)

var (
	Script map[string]interface{} //unmarshalled yabl script
)

func init() {
	Script = make(map[string]interface{})
}

//Check the validity of the input script, and generates IL for interpreter.
func Compile() {
	mainExists := false //a flag to show if exists a valid main function
	for k, v := range Script {
		spiltkey := strings.Fields(k)
		if spiltkey[0] == "func" {
			if len(spiltkey) < 2 {
				continue
			}

			funcName := spiltkey[1]
			//Check if exists one and only one main function.
			if funcName == "main" {
				if mainExists {
					CompileError(funcName, "duplicate main function.")
				} else {
					mainExists = true
				}
				if len(spiltkey) > 2 {
					CompileError(funcName, "main function must not have params.")
				}
			}

			//Check if function names use keywords.
			if GetOpType(funcName) != op_null {
				CompileError(funcName, "can not use keywords as function names.")
			}

			//Check if functions are valid.
			switch ifaceslice := v.(type) {
			case []interface{}:
				strslice := make([]string, 0, len(ifaceslice))
				for _, elem := range ifaceslice {
					switch elemTyped := elem.(type) {
					case string:
						strslice = append(strslice, elemTyped)
					default:
						CompileError(funcName, "wrong function structure.")
					}
				}

				//Compile string script to IL.
				tempFuncIL := Function{spiltkey[2:], make([]Operation, 0, len(strslice))}
				branchStack := stack.NewStack()
				cycleStack := stack.NewStack()
				for _, s := range strslice {
					tempOpIL := Operation{op_null, false, "", false, 0, make([]LexElem, 0)}
					spiltstr := strings.Fields(s)
					for i := 0; i < len(spiltstr); i += 1 {
						tempLexIL := LexElem{}
						//The word is a const string.
						if spiltstr[i][0] == '"' {
							tempLexIL.LT = lex_Constant
							tempString := ""
							runeslise := []rune(spiltstr[i])
							status := false
							ptr := 1
						findstr:
							for {
								switch runeslise[ptr] {
								case '\\':
									if status {
										tempString += "\\"
										status = false
									} else {
										status = true
									}

								case 'n':
									if status {
										tempString += "\n"
										status = false
									} else {
										tempString += "n"
									}

								case 't':
									if status {
										tempString += "\t"
										status = false
									} else {
										tempString += "t"
									}

								case '"':
									if status {
										tempString += "\""
										status = false
									} else {
										tempLexIL.Content = tempString
										if ptr != len(runeslise)-1 {
											CompileWarning(funcName, "ignore character after \"")
										}
										break findstr
									}

								default:
									if status {
										CompileError(funcName, "unknown escape character "+"\\"+string(runeslise[ptr]))
									} else {
										tempString += string(runeslise[ptr])
									}
								}
								ptr += 1
								if ptr == len(runeslise) {
									i += 1
									if i == len(spiltstr) {
										CompileError(funcName, "incomplete string constant.")
									}
									runeslise = []rune(spiltstr[i])
									tempString += " "
									ptr = 0
								}
							}
							tempOpIL.OpElem = append(tempOpIL.OpElem, tempLexIL)
							continue
						}

						//The word is a assignment op.
						if spiltstr[i] == "=" {
							if i == 1 {
								tempOpIL.IsAssign = true
								continue
							} else {
								CompileError(funcName, "misplaced or duplicate \"=\" operation.")
							}
						}

						//The word is identifier or keyword.
						opType := GetOpType(spiltstr[i])
						tempLexIL.Content = spiltstr[i]
						if opType == op_null {
							tempLexIL.LT = lex_Identifier
						} else {
							tempLexIL.LT = lex_Keyword
							if tempOpIL.OT == op_null {
								tempOpIL.OT = opType
								if tempOpIL.IsAssign {
									tempOpIL.OpLoc = i - 1
								} else {
									tempOpIL.OpLoc = i
								}
							}
						}
						tempOpIL.OpElem = append(tempOpIL.OpElem, tempLexIL)
					}

					//Op validity check
					opType := tempOpIL.OT          //type of the operation
					opLoc := tempOpIL.OpLoc        //location of the operation element
					hasAssign := tempOpIL.IsAssign //if the operation has assignment
					opLen := len(tempOpIL.OpElem)  //length of the slice of components of the operation
					switch opType {                //Check validity by opcode.
					case op_null:
						/*
							op_null
							____	____
							assign	param
						*/
						if hasAssign {
							if opLen < 2 {
								CompileError(funcName, "nothing to assign.")
							} else if opLen > 2 {
								CompileWarning(funcName, "more than one params are given, ignore all but first.")
							}
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							}
						} else {
							CompileWarning(funcName, "useless expression, ignoring.")
						}

					case op_if:
						/*
							op_if
							if	____
							op	condition
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a if operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "if operation is in wrong position.")
						}
						if opLen < 2 {
							CompileError(funcName, "if operation has no condition specified.")
						} else if opLen > 2 {
							CompileWarning(funcName, "more than one condition is given to if operation, ignore all but first.")
						}
						switch tempOpIL.OpElem[1].LT {
						case lex_Constant:
							tempOpIL.IsPC = true
							if len(tempOpIL.OpElem[1].Content) > 0 {
								tempOpIL.PCValue = "true"
							}
							CompileWarning(funcName, "useless if block, condition is a constant.")
						case lex_Keyword:
							CompileError(funcName, "condition given to if operation is a keyword "+tempOpIL.OpElem[1].Content+".")
						}
						branchStack.Push(op_if)

					case op_else:
						/*
							op_else
							else
							op
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a else operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "else operation is in wrong position.")
						}
						if opLen > 1 {
							CompileError(funcName, "unexpected elem behind else operation.")
						}
						if branchStack.Len() == 0 {
							CompileError(funcName, "else out of if-fi block.")
						}
						branchStack.Update(op_else)

					case op_elif:
						/*
							op_elif
							elif	____
							op		condition
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a elif operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "elif operation is in wrong position.")
						}
						if opLen < 2 {
							CompileError(funcName, "elif operation has no condition specified.")
						} else if opLen > 2 {
							CompileWarning(funcName, "more than one condition is given to elif operation, ignore all but first.")
						}
						if branchStack.Len() == 0 {
							CompileError(funcName, "elif out of if-fi block.")
						}
						c := branchStack.Peek()
						switch value := c.(type) {
						case int:
							if value == op_else {
								CompileError(funcName, "elif after else.")
							}
						}
						switch tempOpIL.OpElem[1].LT {
						case lex_Constant:
							tempOpIL.IsPC = true
							if len(tempOpIL.OpElem[1].Content) > 0 {
								tempOpIL.PCValue = "true"
							}
							CompileWarning(funcName, "useless elif block, condition is a constant.")
						case lex_Keyword:
							CompileError(funcName, "condition given to elif operation is a keyword "+tempOpIL.OpElem[1].Content+".")
						}

					case op_fi:
						/*
							op_fi
							fi
							op
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a fi operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "fi operation is in wrong position.")
						}
						if opLen > 1 {
							CompileError(funcName, "unexpected elem behind fi operation.")
						}
						if branchStack.Len() == 0 {
							CompileError(funcName, "fi out of if-fi block.")
						}
						branchStack.Pop()

					case op_loop:
						/*
							op_loop
							loop
							op
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a loop operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "loop operation is in wrong position.")
						}
						if opLen > 1 {
							CompileError(funcName, "unexpected elem behind loop operation.")
						}
						cycleStack.Push(op_loop)

					case op_pool:
						/*
							op_pool
							pool
							op
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a pool operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "pool operation is in wrong position.")
						}
						if opLen > 1 {
							CompileError(funcName, "unexpected elem behind pool operation.")
						}
						if cycleStack.Len() == 0 {
							CompileError(funcName, "pool out of loop block.")
						}
						c := cycleStack.Pop()
						switch value := c.(type) {
						case int:
							if value == op_loop {
								CompileInfo(funcName, "infinity loop detected.")
							}
						}

					case op_continue:
						/*
							op_continue
							continue
							op
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a continue operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "continue operation is in wrong position.")
						}
						if opLen > 1 {
							CompileError(funcName, "unexpected elem behind continue operation.")
						}
						if cycleStack.Len() == 0 {
							CompileError(funcName, "continue out of loop block.")
						}

					case op_break:
						/*
							op_break
							break
							op
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a break operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "break operation is in wrong position.")
						}
						if opLen > 1 {
							CompileError(funcName, "unexpected elem behind break operation.")
						}
						if cycleStack.Len() == 0 {
							CompileError(funcName, "break out of loop block.")
						}
						cycleStack.Update(op_break)

					case op_return:
						/*
							op_retrun
							return	____
							op		param
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a return operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "return operation is in wrong position.")
						}
						if opLen > 2 {
							CompileWarning(funcName, "more than one variable is given to return operation, ignore all but first.")
						}
						switch tempOpIL.OpElem[1].LT {
						case lex_Constant:
							tempOpIL.IsPC = true
							tempOpIL.PCValue = tempOpIL.OpElem[1].Content
						case lex_Keyword:
							CompileError(funcName, "condition given to return operation is a keyword "+tempOpIL.OpElem[1].Content+".")
						}

					case op_equal:
						/*
							op_equal
							____	____	equal	____
							assign	param1	op		param2
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 1+prefix {
							CompileError(funcName, "equal operation is in wrong position.")
						}
						if opLen > 3+prefix {
							CompileWarning(funcName, "more than two variable is given to equal operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[0+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to equal operation is a keyword "+tempOpIL.OpElem[0+prefix].Content+".")
						}
						switch tempOpIL.OpElem[2+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to equal operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 2 {
								tempOpIL.IsPC = true
								if tempOpIL.OpElem[0+prefix].Content == tempOpIL.OpElem[2+prefix].Content {
									tempOpIL.PCValue = "true"
								}
								CompileWarning(funcName, "useless calculation, two const strings are given to equal operation.")
							}
						} else {
							CompileWarning(funcName, "result of equal operation is not assigned.")
						}

					case op_and:
						/*
							op_and
							____	____	and		____
							assign	param1	op		param2
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 1+prefix {
							CompileError(funcName, "and operation is in wrong position.")
						}
						if opLen > 3+prefix {
							CompileWarning(funcName, "more than two variable is given to and operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[0+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to and operation is a keyword "+tempOpIL.OpElem[0+prefix].Content+".")
						}
						switch tempOpIL.OpElem[2+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to and operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 2 {
								tempOpIL.IsPC = true
								if len(tempOpIL.OpElem[0+prefix].Content) > 0 && len(tempOpIL.OpElem[2+prefix].Content) > 0 {
									tempOpIL.PCValue = "true"
								}
								CompileWarning(funcName, "useless calculation, two const strings are given to and operation.")
							}
						} else {
							CompileWarning(funcName, "result of and operation is not assigned.")
						}

					case op_or:
						/*
							op_or
							____	____	or	____
							assign	param1	op	param2
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 1+prefix {
							CompileError(funcName, "or operation is in wrong position.")
						}
						if opLen > 3+prefix {
							CompileWarning(funcName, "more than two variable is given to or operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[0+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to or operation is a keyword "+tempOpIL.OpElem[0+prefix].Content+".")
						}
						switch tempOpIL.OpElem[2+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to or operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 2 {
								tempOpIL.IsPC = true
								if len(tempOpIL.OpElem[0+prefix].Content) > 0 || len(tempOpIL.OpElem[2+prefix].Content) > 0 {
									tempOpIL.PCValue = "true"
								}
								CompileWarning(funcName, "useless calculation, two const strings are given to or operation.")
							}
						} else {
							CompileWarning(funcName, "result of or operation is not assigned.")
						}

					case op_not:
						/*
							op_not
							____	not		____
							assign	op		param1
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 0+prefix {
							CompileError(funcName, "and operation is in wrong position.")
						}
						if opLen > 2+prefix {
							CompileWarning(funcName, "more than two variable is given to and operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[1+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to and operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 1 {
								tempOpIL.IsPC = true
								if tempOpIL.OpElem[1+prefix].Content == "" {
									tempOpIL.PCValue = "true"
								}
								CompileWarning(funcName, "useless calculation, two const strings are given to and operation.")
							}
						} else {
							CompileWarning(funcName, "result of and operation is not assigned.")
						}

					case op_join:
						/*
							op_join
							____	____	join	____
							assign	param1	op		param2
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 1+prefix {
							CompileError(funcName, "join operation is in wrong position.")
						}
						if opLen > 3+prefix {
							CompileWarning(funcName, "more than two variable is given to join operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[0+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to join operation is a keyword "+tempOpIL.OpElem[0+prefix].Content+".")
						}
						switch tempOpIL.OpElem[2+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to join operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 2 {
								tempOpIL.IsPC = true
								tempOpIL.PCValue = tempOpIL.OpElem[0+prefix].Content + tempOpIL.OpElem[2+prefix].Content
								CompileWarning(funcName, "useless calculation, two const strings are given to join operation.")
							}
						} else {
							CompileWarning(funcName, "result of join operation is not assigned.")
						}

					case op_contain:
						/*
							op_contain
							____	____	contain	____
							assign	param1	op		param2
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 1+prefix {
							CompileError(funcName, "contain operation is in wrong position.")
						}
						if opLen > 3+prefix {
							CompileWarning(funcName, "more than two variable is given to contain operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[0+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to contain operation is a keyword "+tempOpIL.OpElem[0+prefix].Content+".")
						}
						switch tempOpIL.OpElem[2+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to contain operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 2 {
								tempOpIL.IsPC = true
								if strings.Contains(tempOpIL.OpElem[0+prefix].Content, tempOpIL.OpElem[2+prefix].Content) {
									tempOpIL.PCValue = "true"
								}
								CompileWarning(funcName, "useless calculation, two const strings are given to contain operation.")
							}
						} else {
							CompileWarning(funcName, "result of contain operation is not assigned.")
						}

					case op_hasprefix:
						/*
							op_hasprefix
							____	____	hasprefix	____
							assign	param1	op			param2
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 1+prefix {
							CompileError(funcName, "hasprefix operation is in wrong position.")
						}
						if opLen > 3+prefix {
							CompileWarning(funcName, "more than two variable is given to hasprefix operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[0+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to hasprefix operation is a keyword "+tempOpIL.OpElem[0+prefix].Content+".")
						}
						switch tempOpIL.OpElem[2+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to hasprefix operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 2 {
								tempOpIL.IsPC = true
								if strings.HasPrefix(tempOpIL.OpElem[0+prefix].Content, tempOpIL.OpElem[2+prefix].Content) {
									tempOpIL.PCValue = "true"
								}
								CompileWarning(funcName, "useless calculation, two const strings are given to hasprefix operation.")
							}
						} else {
							CompileWarning(funcName, "result of hasprefix operation is not assigned.")
						}

					case op_hassuffix:
						/*
							op_hassuffix
							____	____	hassuffix	____
							assign	param1	op			param2
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 1+prefix {
							CompileError(funcName, "hassuffix operation is in wrong position.")
						}
						if opLen > 3+prefix {
							CompileWarning(funcName, "more than two variable is given to hassuffix operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.OpElem[0+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to hassuffix operation is a keyword "+tempOpIL.OpElem[0+prefix].Content+".")
						}
						switch tempOpIL.OpElem[2+prefix].LT {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							CompileError(funcName, "variable given to hassuffix operation is a keyword "+tempOpIL.OpElem[2+prefix].Content+".")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
							if constCount == 2 {
								tempOpIL.IsPC = true
								if strings.HasSuffix(tempOpIL.OpElem[0+prefix].Content, tempOpIL.OpElem[2+prefix].Content) {
									tempOpIL.PCValue = "true"
								}
								CompileWarning(funcName, "useless calculation, two const strings are given to hassuffix operation.")
							}
						} else {
							CompileWarning(funcName, "result of hassuffix operation is not assigned.")
						}

					case op_invoke:
						/*
							op_invoke
							____	invoke	____	____	...
							assign	op		func	param1	params
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 0+prefix {
							CompileError(funcName, "invoke operation is in wrong position.")
						}
						if opLen < 2+prefix {
							CompileError(funcName, "no function is provided to invoke.")
						}
						switch tempOpIL.OpElem[1+prefix].LT {
						case lex_Constant:
							CompileError(funcName, "function name given to invoke operation is a const string.")
						case lex_Keyword:
							CompileError(funcName, "function name given to invoke operation is a keyword "+tempOpIL.OpElem[1+prefix].Content+".")
						}
						for i := 2 + prefix; i < opLen; i++ {
							switch tempOpIL.OpElem[i].LT {
							case lex_Keyword:
								CompileError(funcName, "variable given to invoke operation is a keyword "+tempOpIL.OpElem[i].Content+".")
							}
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
						} else {
							CompileWarning(funcName, "result of invoke operation is not assigned.")
						}

					case op_getmsg:
						/*
							op_getmsg
							____	equal
							assign	op
						*/
						prefix := 0
						if hasAssign {
							prefix = 1
						}
						if opLoc != 0+prefix {
							CompileError(funcName, "getmsg operation is in wrong position.")
						}
						if opLen > 1+prefix {
							CompileError(funcName, "unexpected elem behind getmsg operation.")
						}
						if hasAssign {
							switch tempOpIL.OpElem[0].LT {
							case lex_Constant:
								CompileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								CompileError(funcName, "result to assign to is a keyword "+tempOpIL.OpElem[0].Content+".")
							}
						} else {
							CompileWarning(funcName, "result of getmsg operation is not assigned.")
						}

					case op_postmsg:
						/*
							op_postmsg
							postmsg	____
							op		param1
						*/
						if hasAssign {
							CompileError(funcName, "can not assign a postmsg operation.")
						}
						if opLoc != 0 {
							CompileError(funcName, "postmsg operation is in wrong position.")
						}
						if opLen < 2 {
							CompileError(funcName, "postmsg operation has no variable specified.")
						} else if opLen > 2 {
							CompileWarning(funcName, "more than one variable is given to postmsg operation, ignore all but first.")
						}
						switch tempOpIL.OpElem[1].LT {
						case lex_Keyword:
							CompileError(funcName, "variable given to postmsg operation is a keyword "+tempOpIL.OpElem[1].Content+".")
						}

					}
					tempFuncIL.FuncElem = append(tempFuncIL.FuncElem, tempOpIL)
				}
				if branchStack.Len() != 0 {
					CompileError(funcName, "if-fi block is not closed")
				}
				if cycleStack.Len() != 0 {
					CompileError(funcName, "loop block is not closed")
				}
				CompileInfo(funcName, "identified function.")
				if _, ok := IL[funcName]; ok {
					CompileWarning(funcName, "duplicated function "+funcName+", keep the first function")
				} else {
					IL[funcName] = tempFuncIL
				}
			default:
				CompileError(funcName, "wrong function structure.")
			}
		}
	}
	if !mainExists {
		log.Fatalln("[Compiler] Error : no func main found in script.")
	}
}

//Show a error message of compiler and exit.
func CompileError(fname, msg string) {
	log.Fatalln("[Compiler] Error in func", fname, ":", msg)
}

//Show a warning message of compiler.
func CompileWarning(fname, msg string) {
	log.Println("[Compiler] Warning in func", fname, ":", msg)
}

//Show a info message of compiler.
func CompileInfo(fname, msg string) {
	log.Println("[Compiler] Info in func", fname, ":", msg)
}
