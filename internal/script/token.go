package script

type TokenType int

const (
	TokEOF TokenType = iota
	TokNewline

	TokIdent
	TokInt
	TokFloat
	TokString

	TokLet
	TokConst
	TokFn
	TokReturn
	TokIf
	TokElif
	TokElse
	TokWhile
	TokFor
	TokIn
	TokTrue
	TokFalse
	TokNil
	TokBreak
	TokContinue
	TokBreakpoint

	TokAssign
	TokPlusAssign
	TokMinusAssign
	TokStarAssign
	TokSlashAssign

	TokPlus
	TokMinus
	TokStar
	TokSlash
	TokPercent
	TokPower

	TokEq
	TokNeq
	TokLt
	TokGt
	TokLte
	TokGte

	TokAnd
	TokOr
	TokNot

	TokLParen
	TokRParen
	TokLBrace
	TokRBrace
	TokLBracket
	TokRBracket

	TokComma
	TokColon
	TokDot
	TokSemicolon

	TokArrow
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

func (t Token) String() string {
	if t.Literal != "" {
		return t.Literal
	}
	return tokenTypeName(t.Type)
}

func tokenTypeName(tt TokenType) string {
	switch tt {
	case TokEOF:
		return "EOF"
	case TokNewline:
		return "newline"
	case TokIdent:
		return "ident"
	case TokInt:
		return "int"
	case TokFloat:
		return "float"
	case TokString:
		return "string"
	case TokLet:
		return "let"
	case TokConst:
		return "const"
	case TokFn:
		return "fn"
	case TokReturn:
		return "return"
	case TokIf:
		return "if"
	case TokElif:
		return "elif"
	case TokElse:
		return "else"
	case TokWhile:
		return "while"
	case TokFor:
		return "for"
	case TokIn:
		return "in"
	case TokTrue:
		return "true"
	case TokFalse:
		return "false"
	case TokNil:
		return "nil"
	case TokBreak:
		return "break"
	case TokContinue:
		return "continue"
	case TokBreakpoint:
		return "breakpoint"
	default:
		return "unknown"
	}
}
