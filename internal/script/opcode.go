package script

type OpCode int

const (
	OpConstant OpCode = iota
	OpInt
	OpFloat
	OpNil
	OpTrue
	OpFalse

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPow

	OpEq
	OpNeq
	OpLt
	OpGt
	OpLte
	OpGte

	OpAnd
	OpOr
	OpNot
	OpNegate

	OpIndex
	OpDot

	OpJump
	OpJumpIfFalse
	OpJumpIfTrue
	OpLoop

	OpGetLocal
	OpSetLocal
	OpGetGlobal
	OpSetGlobal
	OpDefineGlobal
	OpDefineConst

	OpCall
	OpReturn
	OpPop

	OpBuildList
	OpBuildMap

	OpBreakpoint

	OpNilReturn

	OpGetFree
	OpSetFree
	OpClosure
)

type Instruction struct {
	OpCode OpCode
	Arg    int
	Line   int
}

func (op OpCode) String() string {
	switch op {
	case OpConstant:
		return "CONSTANT"
	case OpInt:
		return "INT"
	case OpFloat:
		return "FLOAT"
	case OpNil:
		return "NIL"
	case OpTrue:
		return "TRUE"
	case OpFalse:
		return "FALSE"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpMod:
		return "MOD"
	case OpPow:
		return "POW"
	case OpEq:
		return "EQ"
	case OpNeq:
		return "NEQ"
	case OpLt:
		return "LT"
	case OpGt:
		return "GT"
	case OpLte:
		return "LTE"
	case OpGte:
		return "GTE"
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	case OpNot:
		return "NOT"
	case OpNegate:
		return "NEGATE"
	case OpIndex:
		return "INDEX"
	case OpDot:
		return "DOT"
	case OpJump:
		return "JUMP"
	case OpJumpIfFalse:
		return "JUMP_IF_FALSE"
	case OpJumpIfTrue:
		return "JUMP_IF_TRUE"
	case OpLoop:
		return "LOOP"
	case OpGetLocal:
		return "GET_LOCAL"
	case OpSetLocal:
		return "SET_LOCAL"
	case OpGetGlobal:
		return "GET_GLOBAL"
	case OpSetGlobal:
		return "SET_GLOBAL"
	case OpDefineGlobal:
		return "DEFINE_GLOBAL"
	case OpDefineConst:
		return "DEFINE_CONST"
	case OpCall:
		return "CALL"
	case OpReturn:
		return "RETURN"
	case OpPop:
		return "POP"
	case OpBuildList:
		return "BUILD_LIST"
	case OpBuildMap:
		return "BUILD_MAP"
	case OpBreakpoint:
		return "BREAKPOINT"
	case OpNilReturn:
		return "NIL_RETURN"
	case OpGetFree:
		return "GET_FREE"
	case OpSetFree:
		return "SET_FREE"
	case OpClosure:
		return "CLOSURE"
	default:
		return "UNKNOWN"
	}
}
