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

func Compile() {
	mainExists := false //a flag to show if exists a valid main function
	for k, v := range Script {
		spiltkey := strings.Fields(k)
		if spiltkey[0] == "func" {
			if len(spiltkey) < 2 {
				continue
			}

			funcName := spiltkey[1]
			//check if exists one and only one main function
			if funcName == "main" {
				if mainExists {
					compileError(funcName, "duplicate main function.")
				} else {
					mainExists = true
				}
				if len(spiltkey) > 2 {
					compileError(funcName, "main function must not have params.")
				}
			}

			//check if function names use keywords
			if getOpType(funcName) != op_null {
				compileError(funcName, "can not use keywords as function names.")
			}

			//check if functions are valid
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
				branchStack := stack.NewStack()
				cycleStack := stack.NewStack()
				for _, s := range strslice {
					tempOpIL := Operation{op_null, false, "", false, 0, make([]LexElem, 0)}
					spiltstr := strings.Fields(s)
					for i := 0; i < len(spiltstr); i += 1 {
						tempLexIL := LexElem{}
						//the word is a const string.
						if spiltstr[i][0] == '"' {
							tempLexIL.lexType = lex_Constant
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
										tempLexIL.content = tempString
										if ptr != len(runeslise)-1 {
											compileWarning(funcName, "ignore character after \"")
										}
										break findstr
									}

								default:
									if status {
										compileError(funcName, "unknown escape character "+"\\"+string(runeslise[ptr]))
									} else {
										tempString += string(runeslise[ptr])
									}
								}
								ptr += 1
								if ptr == len(runeslise) {
									i += 1
									if i == len(spiltstr) {
										compileError(funcName, "incomplete string constant.")
									}
									runeslise = []rune(spiltstr[i])
									tempString += " "
									ptr = 0
								}
							}
							tempOpIL.opElem = append(tempOpIL.opElem, tempLexIL)
							continue
						}

						//word is a assignment op.
						if spiltstr[i] == "=" {
							if i == 1 {
								tempOpIL.assignment = true
								continue
							} else {
								compileError(funcName, "misplaced or duplicate \"=\" operation.")
							}
						}

						//word is identifier or keyword.
						opType := getOpType(spiltstr[i])
						tempLexIL.content = spiltstr[i]
						if opType == op_null {
							tempLexIL.lexType = lex_Identifier
						} else {
							tempLexIL.lexType = lex_Keyword
							if tempOpIL.opType == op_null {
								tempOpIL.opType = opType
								if tempOpIL.assignment {
									tempOpIL.opLocation = i - 1
								} else {
									tempOpIL.opLocation = i
								}
							}
						}
						tempOpIL.opElem = append(tempOpIL.opElem, tempLexIL)
					}

					//op validity check
					opType := tempOpIL.opType        //type of the operation
					opLoc := tempOpIL.opLocation     //location of the operation element
					hasAssign := tempOpIL.assignment //if the operation has assignment
					opLen := len(tempOpIL.opElem)    //length of the slice of components of the operation
					switch opType {                  //check validity by opcode
					case op_null:
						/*
							op_null
							____	____
							assign	param
						*/
						if hasAssign {
							if opLen < 2 {
								compileError(funcName, "nothing to assign.")
							} else if opLen > 2 {
								compileWarning(funcName, "more than one params are given, ignore all but first.")
							}
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							}
						} else {
							compileWarning(funcName, "useless expression, ignoring.")
						}

					case op_if:
						/*
							op_if
							if	____
							op	condition
						*/
						if hasAssign {
							compileError(funcName, "can not assign a if operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "if operation is in wrong position.")
						}
						if opLen < 2 {
							compileError(funcName, "if operation has no condition specified.")
						} else if opLen > 2 {
							compileWarning(funcName, "more than one condition is given to if operation, ignore all but first.")
						}
						switch tempOpIL.opElem[1].lexType {
						case lex_Constant:
							tempOpIL.haspc = true
							if len(tempOpIL.opElem[1].content) > 0 {
								tempOpIL.pcValue = "true"
							}
							compileWarning(funcName, "useless if block, condition is a constant.")
						case lex_Keyword:
							compileError(funcName, "condition given to if operation is a keyword "+tempOpIL.opElem[1].content+".")
						}
						branchStack.Push(op_if)

					case op_else:
						/*
							op_else
							else
							op
						*/
						if hasAssign {
							compileError(funcName, "can not assign a else operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "else operation is in wrong position.")
						}
						if opLen > 1 {
							compileError(funcName, "unexpected elem behind else operation.")
						}
						if branchStack.Len() == 0 {
							compileError(funcName, "else out of if-fi block.")
						}
						branchStack.Update(op_else)

					case op_elif:
						/*
							op_elif
							elif	____
							op		condition
						*/
						if hasAssign {
							compileError(funcName, "can not assign a elif operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "elif operation is in wrong position.")
						}
						if opLen < 2 {
							compileError(funcName, "elif operation has no condition specified.")
						} else if opLen > 2 {
							compileWarning(funcName, "more than one condition is given to elif operation, ignore all but first.")
						}
						if branchStack.Len() == 0 {
							compileError(funcName, "elif out of if-fi block.")
						}
						c := branchStack.Peek()
						switch value := c.(type) {
						case int:
							if value == op_else {
								compileError(funcName, "elif after else.")
							}
						}
						switch tempOpIL.opElem[1].lexType {
						case lex_Constant:
							tempOpIL.haspc = true
							if len(tempOpIL.opElem[1].content) > 0 {
								tempOpIL.pcValue = "true"
							}
							compileWarning(funcName, "useless elif block, condition is a constant.")
						case lex_Keyword:
							compileError(funcName, "condition given to elif operation is a keyword "+tempOpIL.opElem[1].content+".")
						}

					case op_fi:
						/*
							op_fi
							fi
							op
						*/
						if hasAssign {
							compileError(funcName, "can not assign a fi operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "fi operation is in wrong position.")
						}
						if opLen > 1 {
							compileError(funcName, "unexpected elem behind fi operation.")
						}
						if branchStack.Len() == 0 {
							compileError(funcName, "fi out of if-fi block.")
						}
						branchStack.Pop()

					case op_loop:
						/*
							op_loop
							loop
							op
						*/
						if hasAssign {
							compileError(funcName, "can not assign a loop operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "loop operation is in wrong position.")
						}
						if opLen > 1 {
							compileError(funcName, "unexpected elem behind loop operation.")
						}
						cycleStack.Push(op_loop)

					case op_pool:
						/*
							op_pool
							pool
							op
						*/
						if hasAssign {
							compileError(funcName, "can not assign a pool operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "pool operation is in wrong position.")
						}
						if opLen > 1 {
							compileError(funcName, "unexpected elem behind pool operation.")
						}
						if cycleStack.Len() == 0 {
							compileError(funcName, "pool out of loop block.")
						}
						c := cycleStack.Pop()
						switch value := c.(type) {
						case int:
							if value == op_loop {
								compileInfo(funcName, "infinity loop detected.")
							}
						}

					case op_continue:
						/*
							op_continue
							continue
							op
						*/
						if hasAssign {
							compileError(funcName, "can not assign a continue operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "continue operation is in wrong position.")
						}
						if opLen > 1 {
							compileError(funcName, "unexpected elem behind continue operation.")
						}
						if cycleStack.Len() == 0 {
							compileError(funcName, "continue out of loop block.")
						}

					case op_break:
						/*
							op_break
							break
							op
						*/
						if hasAssign {
							compileError(funcName, "can not assign a break operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "break operation is in wrong position.")
						}
						if opLen > 1 {
							compileError(funcName, "unexpected elem behind break operation.")
						}
						if cycleStack.Len() == 0 {
							compileError(funcName, "break out of loop block.")
						}
						cycleStack.Update(op_break)

					case op_return:
						/*
							op_retrun
							return	____
							op		param
						*/
						if hasAssign {
							compileError(funcName, "can not assign a return operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "return operation is in wrong position.")
						}
						if opLen > 2 {
							compileWarning(funcName, "more than one variable is given to return operation, ignore all but first.")
						}
						switch tempOpIL.opElem[1].lexType {
						case lex_Constant:
							tempOpIL.haspc = true
							tempOpIL.pcValue = tempOpIL.opElem[1].content
						case lex_Keyword:
							compileError(funcName, "condition given to return operation is a keyword "+tempOpIL.opElem[1].content+".")
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
							compileError(funcName, "equal operation is in wrong position.")
						}
						if opLen > 3+prefix {
							compileWarning(funcName, "more than two variable is given to equal operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[0+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to equal operation is a keyword "+tempOpIL.opElem[0+prefix].content+".")
						}
						switch tempOpIL.opElem[2+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to equal operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 2 {
								tempOpIL.haspc = true
								if tempOpIL.opElem[0+prefix].content == tempOpIL.opElem[2+prefix].content {
									tempOpIL.pcValue = "true"
								}
								compileWarning(funcName, "useless calculation, two const strings are given to equal operation.")
							}
						} else {
							compileWarning(funcName, "result of equal operation is not assigned.")
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
							compileError(funcName, "and operation is in wrong position.")
						}
						if opLen > 3+prefix {
							compileWarning(funcName, "more than two variable is given to and operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[0+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to and operation is a keyword "+tempOpIL.opElem[0+prefix].content+".")
						}
						switch tempOpIL.opElem[2+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to and operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 2 {
								tempOpIL.haspc = true
								if len(tempOpIL.opElem[0+prefix].content) > 0 && len(tempOpIL.opElem[2+prefix].content) > 0 {
									tempOpIL.pcValue = "true"
								}
								compileWarning(funcName, "useless calculation, two const strings are given to and operation.")
							}
						} else {
							compileWarning(funcName, "result of and operation is not assigned.")
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
							compileError(funcName, "or operation is in wrong position.")
						}
						if opLen > 3+prefix {
							compileWarning(funcName, "more than two variable is given to or operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[0+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to or operation is a keyword "+tempOpIL.opElem[0+prefix].content+".")
						}
						switch tempOpIL.opElem[2+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to or operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 2 {
								tempOpIL.haspc = true
								if len(tempOpIL.opElem[0+prefix].content) > 0 || len(tempOpIL.opElem[2+prefix].content) > 0 {
									tempOpIL.pcValue = "true"
								}
								compileWarning(funcName, "useless calculation, two const strings are given to or operation.")
							}
						} else {
							compileWarning(funcName, "result of or operation is not assigned.")
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
							compileError(funcName, "and operation is in wrong position.")
						}
						if opLen > 2+prefix {
							compileWarning(funcName, "more than two variable is given to and operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[1+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to and operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 1 {
								tempOpIL.haspc = true
								if tempOpIL.opElem[1+prefix].content == "" {
									tempOpIL.pcValue = "true"
								}
								compileWarning(funcName, "useless calculation, two const strings are given to and operation.")
							}
						} else {
							compileWarning(funcName, "result of and operation is not assigned.")
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
							compileError(funcName, "join operation is in wrong position.")
						}
						if opLen > 3+prefix {
							compileWarning(funcName, "more than two variable is given to join operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[0+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to join operation is a keyword "+tempOpIL.opElem[0+prefix].content+".")
						}
						switch tempOpIL.opElem[2+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to join operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 2 {
								tempOpIL.haspc = true
								tempOpIL.pcValue = tempOpIL.opElem[0+prefix].content + tempOpIL.opElem[2+prefix].content
								compileWarning(funcName, "useless calculation, two const strings are given to join operation.")
							}
						} else {
							compileWarning(funcName, "result of join operation is not assigned.")
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
							compileError(funcName, "contain operation is in wrong position.")
						}
						if opLen > 3+prefix {
							compileWarning(funcName, "more than two variable is given to contain operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[0+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to contain operation is a keyword "+tempOpIL.opElem[0+prefix].content+".")
						}
						switch tempOpIL.opElem[2+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to contain operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 2 {
								tempOpIL.haspc = true
								if strings.Contains(tempOpIL.opElem[0+prefix].content, tempOpIL.opElem[2+prefix].content) {
									tempOpIL.pcValue = "true"
								}
								compileWarning(funcName, "useless calculation, two const strings are given to contain operation.")
							}
						} else {
							compileWarning(funcName, "result of contain operation is not assigned.")
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
							compileError(funcName, "hasprefix operation is in wrong position.")
						}
						if opLen > 3+prefix {
							compileWarning(funcName, "more than two variable is given to hasprefix operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[0+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to hasprefix operation is a keyword "+tempOpIL.opElem[0+prefix].content+".")
						}
						switch tempOpIL.opElem[2+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to hasprefix operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 2 {
								tempOpIL.haspc = true
								if strings.HasPrefix(tempOpIL.opElem[0+prefix].content, tempOpIL.opElem[2+prefix].content) {
									tempOpIL.pcValue = "true"
								}
								compileWarning(funcName, "useless calculation, two const strings are given to hasprefix operation.")
							}
						} else {
							compileWarning(funcName, "result of hasprefix operation is not assigned.")
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
							compileError(funcName, "hassuffix operation is in wrong position.")
						}
						if opLen > 3+prefix {
							compileWarning(funcName, "more than two variable is given to hassuffix operation, ignore all but first two.")
						}
						constCount := 0
						switch tempOpIL.opElem[0+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to hassuffix operation is a keyword "+tempOpIL.opElem[0+prefix].content+".")
						}
						switch tempOpIL.opElem[2+prefix].lexType {
						case lex_Constant:
							constCount += 1
						case lex_Keyword:
							compileError(funcName, "variable given to hassuffix operation is a keyword "+tempOpIL.opElem[2+prefix].content+".")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
							if constCount == 2 {
								tempOpIL.haspc = true
								if strings.HasSuffix(tempOpIL.opElem[0+prefix].content, tempOpIL.opElem[2+prefix].content) {
									tempOpIL.pcValue = "true"
								}
								compileWarning(funcName, "useless calculation, two const strings are given to hassuffix operation.")
							}
						} else {
							compileWarning(funcName, "result of hassuffix operation is not assigned.")
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
							compileError(funcName, "invoke operation is in wrong position.")
						}
						if opLen < 2+prefix {
							compileError(funcName, "no function is provided to invoke.")
						}
						switch tempOpIL.opElem[1+prefix].lexType {
						case lex_Constant:
							compileError(funcName, "function name given to invoke operation is a const string.")
						case lex_Keyword:
							compileError(funcName, "function name given to invoke operation is a keyword "+tempOpIL.opElem[1+prefix].content+".")
						}
						for i := 2 + prefix; i < opLen; i++ {
							switch tempOpIL.opElem[i].lexType {
							case lex_Keyword:
								compileError(funcName, "variable given to invoke operation is a keyword "+tempOpIL.opElem[i].content+".")
							}
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
						} else {
							compileWarning(funcName, "result of invoke operation is not assigned.")
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
							compileError(funcName, "getmsg operation is in wrong position.")
						}
						if opLen > 1+prefix {
							compileError(funcName, "unexpected elem behind getmsg operation.")
						}
						if hasAssign {
							switch tempOpIL.opElem[0].lexType {
							case lex_Constant:
								compileError(funcName, "result to assign to is a const string.")
							case lex_Keyword:
								compileError(funcName, "result to assign to is a keyword "+tempOpIL.opElem[0].content+".")
							}
						} else {
							compileWarning(funcName, "result of getmsg operation is not assigned.")
						}

					case op_postmsg:
						/*
							op_postmsg
							____	postmsg	____
							assign	op		param1
						*/
						if hasAssign {
							compileError(funcName, "can not assign a postmsg operation.")
						}
						if opLoc != 0 {
							compileError(funcName, "postmsg operation is in wrong position.")
						}
						if opLen < 2 {
							compileError(funcName, "postmsg operation has no variable specified.")
						} else if opLen > 2 {
							compileWarning(funcName, "more than one variable is given to postmsg operation, ignore all but first.")
						}
						switch tempOpIL.opElem[1].lexType {
						case lex_Keyword:
							compileError(funcName, "variable given to postmsg operation is a keyword "+tempOpIL.opElem[1].content+".")
						}

					}
					tempFuncIL.ops = append(tempFuncIL.ops, tempOpIL)
				}
				if branchStack.Len() != 0 {
					compileError(funcName, "if-fi block is not closed")
				}
				if cycleStack.Len() != 0 {
					compileError(funcName, "loop block is not closed")
				}
				compileInfo(funcName, "identified function.")
				if _, ok := IL[funcName]; ok {
					compileWarning(funcName, "duplicated function "+funcName+", keep the first function")
				} else {
					IL[funcName] = tempFuncIL
				}
			default:
				compileError(funcName, "wrong function structure.")
			}
		}
	}
	if !mainExists {
		log.Fatalln("[Compiler] Error : no func main found in script.")
	}
}

//Show a error message of compiler and exit
func compileError(fname, msg string) {
	log.Fatalln("[Compiler] Error in func", fname, ":", msg)
}

//Show a warning message of compiler
func compileWarning(fname, msg string) {
	log.Println("[Compiler] Warning in func", fname, ":", msg)
}

//Show a info message of compiler
func compileInfo(fname, msg string) {
	log.Println("[Compiler] Info in func", fname, ":", msg)
}
