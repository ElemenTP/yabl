package lib

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
	op_null = iota
	op_if
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
	op_contain
	op_continue
	op_break
)

type Operation struct {
	opType OpType
	opElem []LexElem
}

type Function struct {
	params []string
	ops    []Operation
}
