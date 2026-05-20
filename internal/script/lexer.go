package script

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	input   []rune
	pos     int
	line    int
	col     int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
		line:  1,
		col:   1,
	}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokEOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() rune {
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) peekAt(offset int) rune {
	idx := l.pos + offset
	if idx >= len(l.input) {
		return 0
	}
	return l.input[idx]
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			break
		}
		l.pos++
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.pos++
			l.col++
		} else if ch == '/' && l.peekAt(1) == '/' {
			l.skipComment()
		} else {
			break
		}
	}
}

var keywords = map[string]TokenType{
	"let":         TokLet,
	"const":       TokConst,
	"fn":          TokFn,
	"return":      TokReturn,
	"if":          TokIf,
	"elif":        TokElif,
	"else":        TokElse,
	"while":       TokWhile,
	"for":         TokFor,
	"in":          TokIn,
	"true":        TokTrue,
	"false":       TokFalse,
	"nil":         TokNil,
	"break":       TokBreak,
	"continue":    TokContinue,
	"breakpoint":  TokBreakpoint,
}

func (l *Lexer) nextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokEOF, Line: l.line, Col: l.col}
	}

	ch := l.input[l.pos]

	if ch == '\n' {
		l.advance()
		return Token{Type: TokNewline, Literal: "\n", Line: l.line - 1, Col: l.col - 1}
	}

	if unicode.IsLetter(ch) || ch == '_' {
		return l.readIdent()
	}

	if unicode.IsDigit(ch) {
		return l.readNumber()
	}

	if ch == '"' {
		return l.readString()
	}

	line, col := l.line, l.col

	switch ch {
	case '+':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokPlusAssign, Literal: "+=", Line: line, Col: col}
		}
		return Token{Type: TokPlus, Literal: "+", Line: line, Col: col}
	case '-':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokMinusAssign, Literal: "-=", Line: line, Col: col}
		}
		if l.peek() == '>' {
			l.advance()
			return Token{Type: TokArrow, Literal: "->", Line: line, Col: col}
		}
		return Token{Type: TokMinus, Literal: "-", Line: line, Col: col}
	case '*':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokStarAssign, Literal: "*=", Line: line, Col: col}
		}
		return Token{Type: TokStar, Literal: "*", Line: line, Col: col}
	case '/':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokSlashAssign, Literal: "/=", Line: line, Col: col}
		}
		return Token{Type: TokSlash, Literal: "/", Line: line, Col: col}
	case '%':
		l.advance()
		return Token{Type: TokPercent, Literal: "%", Line: line, Col: col}
	case '^':
		l.advance()
		return Token{Type: TokPower, Literal: "^", Line: line, Col: col}
	case '=':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokEq, Literal: "==", Line: line, Col: col}
		}
		return Token{Type: TokAssign, Literal: "=", Line: line, Col: col}
	case '!':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokNeq, Literal: "!=", Line: line, Col: col}
		}
		return Token{Type: TokNot, Literal: "!", Line: line, Col: col}
	case '<':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokLte, Literal: "<=", Line: line, Col: col}
		}
		return Token{Type: TokLt, Literal: "<", Line: line, Col: col}
	case '>':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokGte, Literal: ">=", Line: line, Col: col}
		}
		return Token{Type: TokGt, Literal: ">", Line: line, Col: col}
	case '&':
		l.advance()
		if l.peek() == '&' {
			l.advance()
			return Token{Type: TokAnd, Literal: "&&", Line: line, Col: col}
		}
		return Token{Type: TokBitAnd, Literal: "&", Line: line, Col: col}
	case '|':
		l.advance()
		if l.peek() == '|' {
			l.advance()
			return Token{Type: TokOr, Literal: "||", Line: line, Col: col}
		}
		return Token{Type: TokBitOr, Literal: "|", Line: line, Col: col}
	case '(':
		l.advance()
		return Token{Type: TokLParen, Literal: "(", Line: line, Col: col}
	case ')':
		l.advance()
		return Token{Type: TokRParen, Literal: ")", Line: line, Col: col}
	case '{':
		l.advance()
		return Token{Type: TokLBrace, Literal: "{", Line: line, Col: col}
	case '}':
		l.advance()
		return Token{Type: TokRBrace, Literal: "}", Line: line, Col: col}
	case '[':
		l.advance()
		return Token{Type: TokLBracket, Literal: "[", Line: line, Col: col}
	case ']':
		l.advance()
		return Token{Type: TokRBracket, Literal: "]", Line: line, Col: col}
	case ',':
		l.advance()
		return Token{Type: TokComma, Literal: ",", Line: line, Col: col}
	case ':':
		l.advance()
		return Token{Type: TokColon, Literal: ":", Line: line, Col: col}
	case '.':
		l.advance()
		return Token{Type: TokDot, Literal: ".", Line: line, Col: col}
	case ';':
		l.advance()
		return Token{Type: TokSemicolon, Literal: ";", Line: line, Col: col}
	}

	l.advance()
	return Token{Type: TokEOF, Line: line, Col: col}
}

func (l *Lexer) readIdent() Token {
	line, col := l.line, l.col
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
		l.col++
	}
	literal := string(l.input[start:l.pos])
	tt, isKeyword := keywords[literal]
	if !isKeyword {
		tt = TokIdent
	}
	return Token{Type: tt, Literal: literal, Line: line, Col: col}
}

func (l *Lexer) readNumber() Token {
	line, col := l.line, l.col
	start := l.pos
	isFloat := false

	for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
		l.pos++
		l.col++
	}

	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		next := l.peekAt(1)
		if unicode.IsDigit(next) {
			isFloat = true
			l.pos++
			l.col++
			for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
				l.pos++
				l.col++
			}
		}
	}

	if l.pos < len(l.input) && (l.input[l.pos] == 'e' || l.input[l.pos] == 'E') {
		isFloat = true
		l.pos++
		l.col++
		if l.pos < len(l.input) && (l.input[l.pos] == '+' || l.input[l.pos] == '-') {
			l.pos++
			l.col++
		}
		for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
			l.pos++
			l.col++
		}
	}

	if l.pos < len(l.input) && (l.input[l.pos] == 'x' || l.input[l.pos] == 'X') {
		l.pos++
		l.col++
		for l.pos < len(l.input) && isHexDigit(l.input[l.pos]) {
			l.pos++
			l.col++
		}
	}

	literal := string(l.input[start:l.pos])
	if isFloat {
		return Token{Type: TokFloat, Literal: literal, Line: line, Col: col}
	}
	return Token{Type: TokInt, Literal: literal, Line: line, Col: col}
}

func (l *Lexer) readString() Token {
	line, col := l.line, l.col
	l.advance()

	var runes []rune
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		ch := l.input[l.pos]
		if ch == '\\' {
			l.advance()
			if l.pos >= len(l.input) {
				break
			}
			escaped := l.input[l.pos]
			switch escaped {
			case 'n':
				runes = append(runes, '\n')
			case 't':
				runes = append(runes, '\t')
			case 'r':
				runes = append(runes, '\r')
			case '\\':
				runes = append(runes, '\\')
			case '"':
				runes = append(runes, '"')
			case '0':
				runes = append(runes, 0)
			default:
				runes = append(runes, '\\', escaped)
			}
			l.advance()
		} else {
			runes = append(runes, ch)
			l.advance()
		}
	}

	if l.pos < len(l.input) {
		l.advance()
	}

	return Token{Type: TokString, Literal: string(runes), Line: line, Col: col}
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func FormatError(msg string, line, col int) string {
	return fmt.Sprintf("line %d, col %d: %s", line, col, msg)
}
