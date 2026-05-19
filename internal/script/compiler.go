package script

import (
	"fmt"
	"strconv"
	"strings"
)

type Compiler struct {
	instructions []Instruction
	constants    []Object
	globals      map[string]int
	globalNames  []string
	localScopes  []map[string]int
	localDepth   int
	localCounts []int
}

func NewCompiler() *Compiler {
	return &Compiler{
		globals:     make(map[string]int),
		globalNames: []string{},
		localScopes: []map[string]int{{}},
		localDepth:  0,
		localCounts: []int{0},
	}
}

func (c *Compiler) Compile(node *Node) error {
	return c.compileNode(node)
}

func (c *Compiler) Instructions() []Instruction {
	return c.instructions
}

func (c *Compiler) Constants() []Object {
	return c.constants
}

func (c *Compiler) GlobalNames() []string {
	return c.globalNames
}

func (c *Compiler) emit(op OpCode, arg int, line int) {
	c.instructions = append(c.instructions, Instruction{OpCode: op, Arg: arg, Line: line})
}

func (c *Compiler) emitJump(op OpCode, line int) int {
	idx := len(c.instructions)
	c.emit(op, 0, line)
	return idx
}

func (c *Compiler) patchJump(idx int) {
	offset := len(c.instructions)
	c.instructions[idx].Arg = offset
}

func (c *Compiler) addConstant(obj Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) resolveVar(name string) (bool, int, bool) {
	for i := len(c.localScopes) - 1; i >= 0; i-- {
		if idx, ok := c.localScopes[i][name]; ok {
			return true, idx, i == 0 && len(c.localScopes) == 1
		}
	}
	if idx, ok := c.globals[name]; ok {
		return false, idx, false
	}
	return false, -1, false
}

func (c *Compiler) pushScope() {
	c.localScopes = append(c.localScopes, make(map[string]int))
	c.localCounts = append(c.localCounts, 0)
	c.localDepth++
}

func (c *Compiler) popScope() {
	c.localScopes = c.localScopes[:len(c.localScopes)-1]
	c.localCounts = c.localCounts[:len(c.localCounts)-1]
	c.localDepth--
}

func (c *Compiler) defineLocal(name string) int {
	scope := c.localScopes[len(c.localScopes)-1]
	idx := c.localCounts[len(c.localCounts)-1]
	scope[name] = idx
	c.localCounts[len(c.localCounts)-1]++
	return idx
}

func (c *Compiler) defineGlobal(name string) int {
	if idx, ok := c.globals[name]; ok {
		return idx
	}
	idx := len(c.globalNames)
	c.globalNames = append(c.globalNames, name)
	c.globals[name] = idx
	return idx
}

func (c *Compiler) compileNode(node *Node) error {
	if node == nil {
		return nil
	}

	switch node.Type {
	case NodeProgram:
		for _, child := range node.Children {
			if err := c.compileNode(child); err != nil {
				return err
			}
		}

	case NodeLetStmt:
		if c.localDepth > 0 {
			return c.compileLetLocal(node)
		}
		return c.compileLetGlobal(node)

	case NodeConstStmt:
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
		idx := c.defineGlobal(node.Value)
		c.emit(OpDefineConst, idx, node.Line)

	case NodeAssignStmt:
		if err := c.compileNode(node.Children[1]); err != nil {
			return err
		}
		target := node.Children[0]
		switch target.Type {
		case NodeIdentExpr:
			isLocal, idx, _ := c.resolveVar(target.Value)
			if isLocal {
				c.emit(OpSetLocal, idx, target.Line)
			} else {
				if _, ok := c.globals[target.Value]; !ok {
					idx = c.defineGlobal(target.Value)
				}
				c.emit(OpSetGlobal, c.globals[target.Value], target.Line)
			}
		case NodeIndexExpr:
			if err := c.compileNode(target.Children[0]); err != nil {
				return err
			}
			if err := c.compileNode(target.Children[1]); err != nil {
				return err
			}
			c.emit(OpIndex, 0, target.Line)
		case NodeDotExpr:
			if err := c.compileNode(target.Children[0]); err != nil {
				return err
			}
			idx := c.addConstant(StringObject(target.Value))
			c.emit(OpConstant, idx, target.Line)
			c.emit(OpIndex, 0, target.Line)
		default:
			return fmt.Errorf("invalid assignment target at line %d", target.Line)
		}

		if node.Token.Type == TokPlusAssign || node.Token.Type == TokMinusAssign ||
			node.Token.Type == TokStarAssign || node.Token.Type == TokSlashAssign {
		}

	case NodeReturnStmt:
		if len(node.Children) > 0 {
			if err := c.compileNode(node.Children[0]); err != nil {
				return err
			}
			c.emit(OpReturn, 0, node.Line)
		} else {
			c.emit(OpNilReturn, 0, node.Line)
		}

	case NodeIfStmt:
		return c.compileIf(node)

	case NodeWhileStmt:
		return c.compileWhile(node)

	case NodeForStmt:
		return c.compileFor(node)

	case NodeFnDecl:
		return c.compileFnDecl(node)

	case NodeBlockStmt:
		c.pushScope()
		for _, child := range node.Children {
			if err := c.compileNode(child); err != nil {
				return err
			}
		}
		c.popScope()

	case NodeBreakStmt:
		c.emit(OpJump, 0, node.Line)

	case NodeContinueStmt:
		c.emit(OpJump, 0, node.Line)

	case NodeBreakpointStmt:
		c.emit(OpBreakpoint, 0, node.Line)

	case NodeBinaryExpr:
		return c.compileBinary(node)

	case NodeUnaryExpr:
		return c.compileUnary(node)

	case NodeCallExpr:
		return c.compileCall(node)

	case NodeIndexExpr:
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
		if err := c.compileNode(node.Children[1]); err != nil {
			return err
		}
		c.emit(OpIndex, 0, node.Line)

	case NodeDotExpr:
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
		idx := c.addConstant(StringObject(node.Value))
		c.emit(OpConstant, idx, node.Line)
		c.emit(OpIndex, 0, node.Line)

	case NodeIntLit:
		val, _ := strconv.ParseInt(node.Value, 0, 64)
		idx := c.addConstant(IntObject(val))
		c.emit(OpConstant, idx, node.Line)

	case NodeFloatLit:
		val, _ := strconv.ParseFloat(node.Value, 64)
		idx := c.addConstant(FloatObject(val))
		c.emit(OpConstant, idx, node.Line)

	case NodeStringLit:
		idx := c.addConstant(StringObject(node.Value))
		c.emit(OpConstant, idx, node.Line)

	case NodeBoolLit:
		if node.Value == "true" {
			c.emit(OpTrue, 0, node.Line)
		} else {
			c.emit(OpFalse, 0, node.Line)
		}

	case NodeNilLit:
		c.emit(OpNil, 0, node.Line)

	case NodeIdentExpr:
		isLocal, idx, _ := c.resolveVar(node.Value)
		if isLocal {
			c.emit(OpGetLocal, idx, node.Line)
		} else {
			idx = c.defineGlobal(node.Value)
			c.emit(OpGetGlobal, idx, node.Line)
		}

	case NodeListLit:
		for _, child := range node.Children {
			if err := c.compileNode(child); err != nil {
				return err
			}
		}
		c.emit(OpBuildList, len(node.Children), node.Line)

	case NodeMapLit:
		for _, pair := range node.Children {
			idx := c.addConstant(StringObject(pair.Value))
			c.emit(OpConstant, idx, pair.Line)
			if err := c.compileNode(pair.Children[0]); err != nil {
				return err
			}
		}
		c.emit(OpBuildMap, len(node.Children), node.Line)

	case NodeFnLiteral:
		return c.compileFnLiteral(node)

	default:
		return fmt.Errorf("unknown node type: %d at line %d", node.Type, node.Line)
	}

	return nil
}

func (c *Compiler) compileLetLocal(node *Node) error {
	if len(node.Children) > 0 {
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
	} else {
		c.emit(OpNil, 0, node.Line)
	}
	idx := c.defineLocal(node.Value)
	c.emit(OpSetLocal, idx, node.Line)
	return nil
}

func (c *Compiler) compileLetGlobal(node *Node) error {
	if len(node.Children) > 0 {
		if err := c.compileNode(node.Children[0]); err != nil {
			return err
		}
	} else {
		c.emit(OpNil, 0, node.Line)
	}
	idx := c.defineGlobal(node.Value)
	c.emit(OpDefineGlobal, idx, node.Line)
	return nil
}

func (c *Compiler) compileBinary(node *Node) error {
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}
	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}
	switch node.Token.Type {
	case TokPlus:
		c.emit(OpAdd, 0, node.Line)
	case TokMinus:
		c.emit(OpSub, 0, node.Line)
	case TokStar:
		c.emit(OpMul, 0, node.Line)
	case TokSlash:
		c.emit(OpDiv, 0, node.Line)
	case TokPercent:
		c.emit(OpMod, 0, node.Line)
	case TokPower:
		c.emit(OpPow, 0, node.Line)
	case TokEq:
		c.emit(OpEq, 0, node.Line)
	case TokNeq:
		c.emit(OpNeq, 0, node.Line)
	case TokLt:
		c.emit(OpLt, 0, node.Line)
	case TokGt:
		c.emit(OpGt, 0, node.Line)
	case TokLte:
		c.emit(OpLte, 0, node.Line)
	case TokGte:
		c.emit(OpGte, 0, node.Line)
	case TokAnd:
		c.emit(OpAnd, 0, node.Line)
	case TokOr:
		c.emit(OpOr, 0, node.Line)
	}
	return nil
}

func (c *Compiler) compileUnary(node *Node) error {
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}
	switch node.Token.Type {
	case TokMinus:
		c.emit(OpNegate, 0, node.Line)
	case TokNot:
		c.emit(OpNot, 0, node.Line)
	}
	return nil
}

func (c *Compiler) compileCall(node *Node) error {
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}
	argCount := len(node.Children) - 1
	for i := 1; i < len(node.Children); i++ {
		if err := c.compileNode(node.Children[i]); err != nil {
			return err
		}
	}
	c.emit(OpCall, argCount, node.Line)
	return nil
}

func (c *Compiler) compileIf(node *Node) error {
	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}
	jumpIfFalse := c.emitJump(OpJumpIfFalse, node.Line)

	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}

	childIdx := 2
	if childIdx < len(node.Children) {
		jumpEnd := c.emitJump(OpJump, node.Line)
		c.patchJump(jumpIfFalse)

		for childIdx < len(node.Children) {
			child := node.Children[childIdx]
			childIdx++
			if child.Type == NodeBlockStmt {
				if err := c.compileNode(child); err != nil {
					return err
				}
			} else if child.Type == NodeIfStmt {
				if err := c.compileIf(child); err != nil {
					return err
				}
			}
		}
		c.patchJump(jumpEnd)
	} else {
		c.patchJump(jumpIfFalse)
	}

	return nil
}

func (c *Compiler) compileWhile(node *Node) error {
	loopStart := len(c.instructions)

	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}

	jumpIfFalse := c.emitJump(OpJumpIfFalse, node.Line)

	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}

	c.emit(OpLoop, loopStart, node.Line)
	c.patchJump(jumpIfFalse)

	return nil
}

func (c *Compiler) compileFor(node *Node) error {
	c.emit(OpNil, 0, node.Line)
	idx := c.defineLocal("__iter_" + node.Value)
	c.emit(OpSetLocal, idx, node.Line)

	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}

	iterIdx := c.defineLocal("__iter_list")
	c.emit(OpSetLocal, iterIdx, node.Line)

	c.emit(OpInt, 0, node.Line)
	counterIdx := c.defineLocal("__iter_idx")
	c.emit(OpSetLocal, counterIdx, node.Line)

	loopStart := len(c.instructions)

	c.emit(OpGetLocal, iterIdx, node.Line)
	c.emit(OpGetLocal, counterIdx, node.Line)
	c.emit(OpIndex, 0, node.Line)
	valIdx := c.defineLocal(node.Value)
	c.emit(OpSetLocal, valIdx, node.Line)

	c.emit(OpGetLocal, counterIdx, node.Line)
	listLenIdx := c.defineGlobal("list_len")
	c.emit(OpGetGlobal, listLenIdx, node.Line)
	c.emit(OpGetLocal, iterIdx, node.Line)
	c.emit(OpCall, 1, node.Line)
	c.emit(OpLt, 0, node.Line)

	jumpIfFalse := c.emitJump(OpJumpIfFalse, node.Line)

	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}

	c.emit(OpGetLocal, counterIdx, node.Line)
	oneIdx := c.addConstant(IntObject(1))
	c.emit(OpConstant, oneIdx, node.Line)
	c.emit(OpAdd, 0, node.Line)
	c.emit(OpSetLocal, counterIdx, node.Line)

	c.emit(OpLoop, loopStart, node.Line)
	c.patchJump(jumpIfFalse)

	return nil
}

func (c *Compiler) compileFnDecl(node *Node) error {
	fn := c.compileFnBody(node)
	idx := c.addConstant(fn)
	c.emit(OpConstant, idx, node.Line)
	name := node.Value
	if i := strings.Index(name, "("); i > 0 {
		name = name[:i]
	}
	globalIdx := c.defineGlobal(name)
	c.emit(OpDefineGlobal, globalIdx, node.Line)
	return nil
}

func (c *Compiler) compileFnLiteral(node *Node) error {
	fn := c.compileFnBody(node)
	idx := c.addConstant(fn)
	c.emit(OpConstant, idx, node.Line)
	return nil
}

func (c *Compiler) compileFnBody(node *Node) *FnObject {
	paramNode := node.Children[0]
	bodyNode := node.Children[len(node.Children)-1]

	params := []string{}
	if paramNode != nil {
		for _, p := range paramNode.Children {
			params = append(params, p.Value)
		}
	}

	subCompiler := NewCompiler()
	subCompiler.pushScope()

	for _, p := range params {
		subCompiler.defineLocal(p)
	}

	if err := subCompiler.compileNode(bodyNode); err != nil {
		subCompiler.emit(OpNilReturn, 0, 0)
	}

	subCompiler.emit(OpNilReturn, 0, 0)

	return &FnObject{
		Name:         node.Value,
		Params:       params,
		Instructions: subCompiler.Instructions(),
		Constants:    subCompiler.Constants(),
		NumLocals:    subCompiler.localCounts[0],
	}
}
