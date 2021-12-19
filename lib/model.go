package lib

var (
	keywordMap map[string]OpType //a string to op code mapping, to simplify keyword identifing
)

//Type of lexical elements.
type LexType int

const (
	lex_Identifier = iota
	lex_Keyword
	lex_Constant
)

//A struct to contain lexical elements.
type LexElem struct {
	LT      LexType //type of the lexical elements
	Content string  //content of the lexical elements
}

//Type of operations.
type OpType int

const (
	op_null = iota
	op_if
	op_else
	op_elif
	op_fi
	op_loop
	op_pool
	op_continue
	op_break
	op_return
	op_equal
	op_and
	op_or
	op_not
	op_join
	op_contain
	op_hasprefix
	op_hassuffix
	op_invoke
	op_getmsg
	op_postmsg
)

//A struct to contain operations.
type Operation struct {
	OT       OpType    //type of the operation
	IsPC     bool      //if the result of the operation is precompiled
	PCValue  string    //value of precompiled value
	IsAssign bool      //if the operation has assignment
	OpLoc    int       //location of the operation element, only used in compiling.
	OpElem   []LexElem //slice of lexical elements composed of the operation.
}

//A struct to contain functions.
type Function struct {
	Params   []string    //param slice to give to the function
	FuncElem []Operation //slice of operations composed of the function.
}

func init() {
	keywordMap = map[string]OpType{
		"if":        op_if,
		"else":      op_else,
		"elif":      op_elif,
		"fi":        op_fi,
		"loop":      op_loop,
		"pool":      op_pool,
		"continue":  op_continue,
		"break":     op_break,
		"return":    op_return,
		"equal":     op_equal,
		"and":       op_and,
		"or":        op_or,
		"not":       op_not,
		"join":      op_join,
		"contain":   op_contain,
		"hasprefix": op_hasprefix,
		"hassuffix": op_hassuffix,
		"invoke":    op_invoke,
		"getmsg":    op_getmsg,
		"postmsg":   op_postmsg,
	}
}

//Use the string-opcode map to identify keywords.
func GetOpType(content string) OpType {
	res, ok := keywordMap[content]
	if ok {
		return OpType(res)
	}
	return op_null
}
