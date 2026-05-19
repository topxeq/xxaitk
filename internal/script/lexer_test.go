package script

import "testing"

func TestLexerTokens(t *testing.T) {
	input := `let x = 10 + 20`
	lexer := NewLexer(input)
	tokens := lexer.Tokenize()
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TokLet, "let"},
		{TokIdent, "x"},
		{TokAssign, "="},
		{TokInt, "10"},
		{TokPlus, "+"},
		{TokInt, "20"},
		{TokEOF, ""},
	}
	for i, exp := range expected {
		if i >= len(tokens) {
			t.Fatalf("not enough tokens, expected %d, got %d", len(expected), len(tokens))
		}
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d].Type = %v, want %v", i, tokens[i].Type, exp.typ)
		}
		if exp.lit != "" && tokens[i].Literal != exp.lit {
			t.Errorf("token[%d].Literal = %q, want %q", i, tokens[i].Literal, exp.lit)
		}
	}
}

func TestLexerString(t *testing.T) {
	input := `"hello world"`
	tokens := NewLexer(input).Tokenize()
	if tokens[0].Type != TokString {
		t.Fatalf("expected TokString, got %v", tokens[0].Type)
	}
	if tokens[0].Literal != "hello world" {
		t.Errorf("Literal = %q, want %q", tokens[0].Literal, "hello world")
	}
}

func TestLexerStringEscapes(t *testing.T) {
	input := `"hello\nworld\ttab"`
	tokens := NewLexer(input).Tokenize()
	if tokens[0].Literal != "hello\nworld\ttab" {
		t.Errorf("Literal = %q, want %q", tokens[0].Literal, "hello\nworld\ttab")
	}
}

func TestLexerKeywords(t *testing.T) {
	keywords := map[string]TokenType{
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
	for kw, expected := range keywords {
		tokens := NewLexer(kw).Tokenize()
		if tokens[0].Type != expected {
			t.Errorf("keyword %q: got %v, want %v", kw, tokens[0].Type, expected)
		}
	}
}

func TestLexerFloat(t *testing.T) {
	input := "3.14 1e10 0.5"
	tokens := NewLexer(input).Tokenize()
	for i, exp := range []TokenType{TokFloat, TokFloat, TokFloat} {
		if tokens[i].Type != exp {
			t.Errorf("token[%d] = %v, want %v", i, tokens[i].Type, exp)
		}
	}
}

func TestLexerOperators(t *testing.T) {
	input := `== != <= >= && || += -= *= /= ->`
	tokens := NewLexer(input).Tokenize()
	expected := []TokenType{TokEq, TokNeq, TokLte, TokGte, TokAnd, TokOr, TokPlusAssign, TokMinusAssign, TokStarAssign, TokSlashAssign, TokArrow}
	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("token[%d] = %v, want %v (lit=%q)", i, tokens[i].Type, exp, tokens[i].Literal)
		}
	}
}

func TestLexerComment(t *testing.T) {
	input := "x // this is a comment\ny"
	tokens := NewLexer(input).Tokenize()
	var lits []string
	for _, tok := range tokens {
		if tok.Type == TokIdent {
			lits = append(lits, tok.Literal)
		}
	}
	if len(lits) != 2 || lits[0] != "x" || lits[1] != "y" {
		t.Errorf("expected [x y], got %v", lits)
	}
}
