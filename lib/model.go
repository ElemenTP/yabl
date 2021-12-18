package lib

var (
	keywordMap map[string]OpType //a string to op code mapping, to simplify keyword identifing
)

//type of lexical elements
type LexType int

const (
	lex_Identifier = iota
	lex_Keyword
	lex_Constant
)

//a struct to contain lexical elements
type LexElem struct {
	lexType LexType //type of the lexical elements
	content string  //content of the lexical elements
}

//type of operations.
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

//a struct to contain operations
type Operation struct {
	opType     OpType    //type of the operation
	haspc      bool      //if the result of the operation is precompiled
	pcValue    string    //value of precompiled value
	assignment bool      //if the operation has assignment
	opLocation int       //location of the operation element, only used in compiling.
	opElem     []LexElem //slice of lexical elements composed of the operation.
}

//a struct to contain functions
type Function struct {
	params []string    //param slice to give to the function
	ops    []Operation //slice of operations composed of the function.
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

//use the string-opcode map to identify keywords
func getOpType(content string) OpType {
	res, ok := keywordMap[content]
	if ok {
		return OpType(res)
	}
	return op_null
}
