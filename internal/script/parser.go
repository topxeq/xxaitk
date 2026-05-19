package script

import (
	"fmt"
)

type Parser struct {
	tokens  []Token
	pos     int
}

func NewParser(tokens []Token) *Parser {
	tokens = filterNewlines(tokens)
	return &Parser{tokens: tokens}
}

func filterNewlines(tokens []Token) []Token {
	var result []Token
	for _, t := range tokens {
		if t.Type != TokNewline {
			result = append(result, t)
		}
	}
	return result
}

func (p *Parser) Parse() (*Node, error) {
	program := &Node{Type: NodeProgram}
	for !p.atEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.addChild(stmt)
		}
	}
	return program, nil
}

func (p *Parser) cur() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek() Token {
	if p.pos+1 >= len(p.tokens) {
		return Token{Type: TokEOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() Token {
	t := p.cur()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return t
}

func (p *Parser) expect(tt TokenType) (Token, error) {
	t := p.cur()
	if t.Type != tt {
		return t, fmt.Errorf("expected %s, got %s (%q) at line %d",
			tokenTypeName(tt), tokenTypeName(t.Type), t.Literal, t.Line)
	}
	p.advance()
	return t, nil
}

func (p *Parser) atEnd() bool {
	return p.cur().Type == TokEOF
}

func (p *Parser) parseStatement() (*Node, error) {
	t := p.cur()
	switch t.Type {
	case TokLet:
		return p.parseLet()
	case TokConst:
		return p.parseConst()
	case TokReturn:
		return p.parseReturn()
	case TokIf:
		return p.parseIf()
	case TokWhile:
		return p.parseWhile()
	case TokFor:
		return p.parseFor()
	case TokFn:
		if p.peek().Type == TokIdent {
			return p.parseFnDecl()
		}
		return p.parseExprStatement()
	case TokBreak:
		p.advance()
		return &Node{Type: NodeBreakStmt, Token: t, Line: t.Line}, nil
	case TokContinue:
		p.advance()
		return &Node{Type: NodeContinueStmt, Token: t, Line: t.Line}, nil
	case TokBreakpoint:
		p.advance()
		return &Node{Type: NodeBreakpointStmt, Token: t, Line: t.Line}, nil
	default:
		return p.parseExprStatement()
	}
}

func (p *Parser) parseLet() (*Node, error) {
	t := p.advance()
	name, err := p.expect(TokIdent)
	if err != nil {
		return nil, err
	}
	node := &Node{Type: NodeLetStmt, Token: t, Value: name.Literal, Line: t.Line}
	if p.cur().Type == TokAssign {
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		node.addChild(expr)
	}
	return node, nil
}

func (p *Parser) parseConst() (*Node, error) {
	t := p.advance()
	name, err := p.expect(TokIdent)
	if err != nil {
		return nil, err
	}
	node := &Node{Type: NodeConstStmt, Token: t, Value: name.Literal, Line: t.Line}
	_, err = p.expect(TokAssign)
	if err != nil {
		return nil, err
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.addChild(expr)
	return node, nil
}

func (p *Parser) parseReturn() (*Node, error) {
	t := p.advance()
	node := &Node{Type: NodeReturnStmt, Token: t, Line: t.Line}
	if !p.isExprStart() {
		return node, nil
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.addChild(expr)
	return node, nil
}

func (p *Parser) isExprStart() bool {
	t := p.cur()
	switch t.Type {
	case TokInt, TokFloat, TokString, TokTrue, TokFalse, TokNil,
		TokIdent, TokLParen, TokLBracket, TokLBrace, TokFn,
		TokMinus, TokNot:
		return true
	}
	return false
}

func (p *Parser) parseIf() (*Node, error) {
	t := p.advance()
	node := &Node{Type: NodeIfStmt, Token: t, Line: t.Line}

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.addChild(cond)

	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	node.addChild(block)

	for p.cur().Type == TokElif {
		elifNode := &Node{Type: NodeIfStmt, Token: p.cur(), Line: p.cur().Line}
		p.advance()
		elifCond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elifNode.addChild(elifCond)
		elifBlock, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		elifNode.addChild(elifBlock)
		node.addChild(elifNode)
	}

	if p.cur().Type == TokElse {
		p.advance()
		elseBlock, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		node.addChild(elseBlock)
	}

	return node, nil
}

func (p *Parser) parseWhile() (*Node, error) {
	t := p.advance()
	node := &Node{Type: NodeWhileStmt, Token: t, Line: t.Line}

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.addChild(cond)

	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	node.addChild(block)

	return node, nil
}

func (p *Parser) parseFor() (*Node, error) {
	t := p.advance()
	name, err := p.expect(TokIdent)
	if err != nil {
		return nil, err
	}
	_, err = p.expect(TokIn)
	if err != nil {
		return nil, err
	}

	node := &Node{Type: NodeForStmt, Token: t, Value: name.Literal, Line: t.Line}

	iter, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.addChild(iter)

	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	node.addChild(block)

	return node, nil
}

func (p *Parser) parseFnDecl() (*Node, error) {
	t := p.advance()
	name, err := p.expect(TokIdent)
	if err != nil {
		return nil, err
	}
	node := &Node{Type: NodeFnDecl, Token: t, Value: name.Literal, Line: t.Line}

	_, err = p.expect(TokLParen)
	if err != nil {
		return nil, err
	}

	params := []string{}
	for p.cur().Type != TokRParen {
		param, err := p.expect(TokIdent)
		if err != nil {
			return nil, err
		}
		params = append(params, param.Literal)
		if p.cur().Type == TokComma {
			p.advance()
		} else {
			break
		}
	}
	_, err = p.expect(TokRParen)
	if err != nil {
		return nil, err
	}

	node.Value = name.Literal + "(" + joinParams(params) + ")"

	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	node.addChild(block)

	paramNode := &Node{Value: joinParams(params)}
	for _, param := range params {
		paramNode.addChild(&Node{Value: param})
	}
	node.Children = append([]*Node{paramNode}, node.Children...)

	return node, nil
}

func joinParams(params []string) string {
	result := ""
	for i, p := range params {
		if i > 0 {
			result += ","
		}
		result += p
	}
	return result
}

func (p *Parser) parseBlock() (*Node, error) {
	_, err := p.expect(TokLBrace)
	if err != nil {
		return nil, err
	}
	block := &Node{Type: NodeBlockStmt}
	for p.cur().Type != TokRBrace && !p.atEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			block.addChild(stmt)
		}
	}
	_, err = p.expect(TokRBrace)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (p *Parser) parseExprStatement() (*Node, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.cur().Type == TokAssign {
		p.advance()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		assign := &Node{Type: NodeAssignStmt, Line: expr.Line}
		assign.addChild(expr)
		assign.addChild(right)
		return assign, nil
	}

	compoundOps := map[TokenType]bool{
		TokPlusAssign:  true,
		TokMinusAssign: true,
		TokStarAssign:  true,
		TokSlashAssign: true,
	}
	if compoundOps[p.cur().Type] {
		opTok := p.advance()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		assign := &Node{Type: NodeAssignStmt, Token: opTok, Line: expr.Line}
		assign.addChild(expr)
		assign.addChild(right)
		return assign, nil
	}

	return expr, nil
}

func (p *Parser) parseExpression() (*Node, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (*Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TokOr {
		t := p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeBinaryExpr, Token: t, Line: t.Line}
		node.addChild(left)
		node.addChild(right)
		left = node
	}
	return left, nil
}

func (p *Parser) parseAnd() (*Node, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TokAnd {
		t := p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeBinaryExpr, Token: t, Line: t.Line}
		node.addChild(left)
		node.addChild(right)
		left = node
	}
	return left, nil
}

func (p *Parser) parseEquality() (*Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TokEq || p.cur().Type == TokNeq {
		t := p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeBinaryExpr, Token: t, Line: t.Line}
		node.addChild(left)
		node.addChild(right)
		left = node
	}
	return left, nil
}

func (p *Parser) parseComparison() (*Node, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TokLt || p.cur().Type == TokGt || p.cur().Type == TokLte || p.cur().Type == TokGte {
		t := p.advance()
		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeBinaryExpr, Token: t, Line: t.Line}
		node.addChild(left)
		node.addChild(right)
		left = node
	}
	return left, nil
}

func (p *Parser) parseAddition() (*Node, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TokPlus || p.cur().Type == TokMinus {
		t := p.advance()
		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeBinaryExpr, Token: t, Line: t.Line}
		node.addChild(left)
		node.addChild(right)
		left = node
	}
	return left, nil
}

func (p *Parser) parseMultiplication() (*Node, error) {
	left, err := p.parsePower()
	if err != nil {
		return nil, err
	}
	for p.cur().Type == TokStar || p.cur().Type == TokSlash || p.cur().Type == TokPercent {
		t := p.advance()
		right, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeBinaryExpr, Token: t, Line: t.Line}
		node.addChild(left)
		node.addChild(right)
		left = node
	}
	return left, nil
}

func (p *Parser) parsePower() (*Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	if p.cur().Type == TokPower {
		t := p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeBinaryExpr, Token: t, Line: t.Line}
		node.addChild(left)
		node.addChild(right)
		left = node
	}
	return left, nil
}

func (p *Parser) parseUnary() (*Node, error) {
	if p.cur().Type == TokMinus || p.cur().Type == TokNot {
		t := p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		node := &Node{Type: NodeUnaryExpr, Token: t, Line: t.Line}
		node.addChild(operand)
		return node, nil
	}
	return p.parseCall()
}

func (p *Parser) parseCall() (*Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		switch p.cur().Type {
		case TokLParen:
			t := p.advance()
			callNode := &Node{Type: NodeCallExpr, Token: t, Line: t.Line}
			callNode.addChild(expr)
			for p.cur().Type != TokRParen && !p.atEnd() {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				callNode.addChild(arg)
				if p.cur().Type == TokComma {
					p.advance()
				} else {
					break
				}
			}
			_, err = p.expect(TokRParen)
			if err != nil {
				return nil, err
			}
			expr = callNode
		case TokLBracket:
			t := p.advance()
			idx, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(TokRBracket)
			if err != nil {
				return nil, err
			}
			indexNode := &Node{Type: NodeIndexExpr, Token: t, Line: t.Line}
			indexNode.addChild(expr)
			indexNode.addChild(idx)
			expr = indexNode
		case TokDot:
			p.advance()
			field, err := p.expect(TokIdent)
			if err != nil {
				return nil, err
			}
			dotNode := &Node{Type: NodeDotExpr, Token: field, Value: field.Literal, Line: field.Line}
			dotNode.addChild(expr)
			expr = dotNode
		default:
			return expr, nil
		}
	}
}

func (p *Parser) parsePrimary() (*Node, error) {
	t := p.cur()
	switch t.Type {
	case TokInt:
		p.advance()
		var intVal int64
		fmt.Sscanf(t.Literal, "%d", &intVal)
		return &Node{Type: NodeIntLit, Token: t, Value: t.Literal, IntVal: intVal, Line: t.Line}, nil
	case TokFloat:
		p.advance()
		var floatVal float64
		fmt.Sscanf(t.Literal, "%f", &floatVal)
		return &Node{Type: NodeFloatLit, Token: t, Value: t.Literal, FloatVal: floatVal, Line: t.Line}, nil
	case TokString:
		p.advance()
		return &Node{Type: NodeStringLit, Token: t, Value: t.Literal, Line: t.Line}, nil
	case TokTrue:
		p.advance()
		return &Node{Type: NodeBoolLit, Token: t, Value: "true", Line: t.Line}, nil
	case TokFalse:
		p.advance()
		return &Node{Type: NodeBoolLit, Token: t, Value: "false", Line: t.Line}, nil
	case TokNil:
		p.advance()
		return &Node{Type: NodeNilLit, Token: t, Line: t.Line}, nil
	case TokIdent:
		p.advance()
		return &Node{Type: NodeIdentExpr, Token: t, Value: t.Literal, Line: t.Line}, nil
	case TokLParen:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(TokRParen)
		if err != nil {
			return nil, err
		}
		return expr, nil
	case TokLBracket:
		return p.parseListLit()
	case TokLBrace:
		return p.parseMapLit()
	case TokFn:
		return p.parseFnLiteral()
	default:
		return nil, fmt.Errorf("unexpected token %s (%q) at line %d",
			tokenTypeName(t.Type), t.Literal, t.Line)
	}
}

func (p *Parser) parseListLit() (*Node, error) {
	t := p.advance()
	node := &Node{Type: NodeListLit, Token: t, Line: t.Line}
	for p.cur().Type != TokRBracket && !p.atEnd() {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		node.addChild(elem)
		if p.cur().Type == TokComma {
			p.advance()
		} else {
			break
		}
	}
	_, err := p.expect(TokRBracket)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseMapLit() (*Node, error) {
	t := p.advance()
	node := &Node{Type: NodeMapLit, Token: t, Line: t.Line}
	for p.cur().Type != TokRBrace && !p.atEnd() {
		var keyStr string
		if p.cur().Type == TokIdent {
			keyStr = p.advance().Literal
		} else {
			key, err := p.expect(TokString)
			if err != nil {
				return nil, err
			}
			keyStr = key.Literal
		}
		_, err := p.expect(TokColon)
		if err != nil {
			return nil, err
		}
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		pair := &Node{Value: keyStr}
		pair.addChild(val)
		node.addChild(pair)
		if p.cur().Type == TokComma {
			p.advance()
		} else {
			break
		}
	}
	_, err := p.expect(TokRBrace)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseFnLiteral() (*Node, error) {
	t := p.advance()
	node := &Node{Type: NodeFnLiteral, Token: t, Line: t.Line}

	_, err := p.expect(TokLParen)
	if err != nil {
		return nil, err
	}
	params := []string{}
	for p.cur().Type != TokRParen {
		param, err := p.expect(TokIdent)
		if err != nil {
			return nil, err
		}
		params = append(params, param.Literal)
		if p.cur().Type == TokComma {
			p.advance()
		} else {
			break
		}
	}
	_, err = p.expect(TokRParen)
	if err != nil {
		return nil, err
	}

	paramNode := &Node{Value: joinParams(params)}
	for _, param := range params {
		paramNode.addChild(&Node{Value: param})
	}
	node.addChild(paramNode)

	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	node.addChild(block)

	return node, nil
}
