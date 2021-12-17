package lib

var (
	keywordMap map[string]OpType
)

type LexType int

const (
	lex_Identifier = iota
	lex_Keyword
	lex_Constant
)

type LexElem struct {
	lexType LexType
	content string
}

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

type Operation struct {
	opType     OpType
	haspc      bool
	pcValue    string
	assignment bool
	opLocation int
	opElem     []LexElem
}

type Function struct {
	params []string
	ops    []Operation
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

func getOpType(content string) OpType {
	res, ok := keywordMap[content]
	if ok {
		return OpType(res)
	}
	return op_null
}
