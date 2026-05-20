package script

import (
	"fmt"
	"strconv"
	"strings"
)

type loopContext struct {
	loopStart    int
	breakJumps   []int
	continueJumps []int
}

type Compiler struct {
	instructions []Instruction
	constants    []Object
	globals      map[string]int
	globalNames  []string
	localScopes  []map[string]int
	localDepth   int
	localCounts  []int
	freeVars     []string
	freeVarSet   map[string]int
	parentLocals []map[string]int
	loopStack    []loopContext
}

func NewCompiler() *Compiler {
	return &Compiler{
		globals:      make(map[string]int),
		globalNames:  []string{},
		localScopes:  []map[string]int{{}},
		localDepth:   0,
		localCounts:  []int{0},
		freeVars:     []string{},
		freeVarSet:   map[string]int{},
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

func (c *Compiler) allLocalScopes() []map[string]int {
	result := make([]map[string]int, len(c.localScopes))
	for i, scope := range c.localScopes {
		newScope := make(map[string]int, len(scope))
		for k, v := range scope {
			newScope[k] = v
		}
		result[i] = newScope
	}
	return result
}

func (c *Compiler) defineLocal(name string) int {
	scope := c.localScopes[len(c.localScopes)-1]
	idx := c.localCounts[len(c.localCounts)-1]
	scope[name] = idx
	c.localCounts[len(c.localCounts)-1]++
	return idx
}

func (c *Compiler) isParentLocal(name string) bool {
	for _, scope := range c.parentLocals {
		if _, ok := scope[name]; ok {
			return true
		}
	}
	return false
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

func (c *Compiler) compileLoopBody(node *Node) error {
	if node.Type != NodeBlockStmt {
		return c.compileNode(node)
	}
	c.pushScope()
	for _, child := range node.Children {
		if err := c.compileNode(child); err != nil {
			return err
		}
		if c.isExpressionNode(child.Type) {
			c.emit(OpPop, 0, child.Line)
		}
	}
	c.popScope()
	return nil
}

func (c *Compiler) isExpressionNode(t NodeType) bool {
	switch t {
	case NodeLetStmt, NodeConstStmt, NodeAssignStmt, NodeReturnStmt,
		NodeIfStmt, NodeWhileStmt, NodeForStmt, NodeFnDecl,
		NodeBreakStmt, NodeContinueStmt, NodeBreakpointStmt:
		return false
	default:
		return true
	}
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
		target := node.Children[0]
		isCompound := node.Token.Type == TokPlusAssign || node.Token.Type == TokMinusAssign ||
			node.Token.Type == TokStarAssign || node.Token.Type == TokSlashAssign

		if isCompound {
			switch target.Type {
			case NodeIdentExpr:
				isLocal, idx, _ := c.resolveVar(target.Value)
				if isLocal {
					c.emit(OpGetLocal, idx, target.Line)
				} else if freeIdx, ok := c.freeVarSet[target.Value]; ok {
					c.emit(OpGetFree, freeIdx, target.Line)
				} else if c.isParentLocal(target.Value) {
					freeIdx := len(c.freeVars)
					c.freeVars = append(c.freeVars, target.Value)
					c.freeVarSet[target.Value] = freeIdx
					c.emit(OpGetFree, freeIdx, target.Line)
				} else {
					if _, ok := c.globals[target.Value]; !ok {
						c.defineGlobal(target.Value)
					}
					c.emit(OpGetGlobal, c.globals[target.Value], target.Line)
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
				return fmt.Errorf("invalid compound assignment target at line %d", target.Line)
			}

			if err := c.compileNode(node.Children[1]); err != nil {
				return err
			}

			switch node.Token.Type {
			case TokPlusAssign:
				c.emit(OpAdd, 0, node.Line)
			case TokMinusAssign:
				c.emit(OpSub, 0, node.Line)
			case TokStarAssign:
				c.emit(OpMul, 0, node.Line)
			case TokSlashAssign:
				c.emit(OpDiv, 0, node.Line)
			}
		} else {
			if err := c.compileNode(node.Children[1]); err != nil {
				return err
			}
		}

		switch target.Type {
		case NodeIdentExpr:
			isLocal, idx, _ := c.resolveVar(target.Value)
			if isLocal {
				c.emit(OpSetLocal, idx, target.Line)
			} else if isCompound {
				if freeIdx, ok := c.freeVarSet[target.Value]; ok {
					c.emit(OpSetFree, freeIdx, target.Line)
				} else {
					c.emit(OpSetGlobal, c.globals[target.Value], target.Line)
				}
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
		if len(c.loopStack) == 0 {
			return fmt.Errorf("break outside loop at line %d", node.Line)
		}
		jumpIdx := c.emitJump(OpJump, node.Line)
		c.loopStack[len(c.loopStack)-1].breakJumps = append(c.loopStack[len(c.loopStack)-1].breakJumps, jumpIdx)

	case NodeContinueStmt:
		if len(c.loopStack) == 0 {
			return fmt.Errorf("continue outside loop at line %d", node.Line)
		}
		jumpIdx := c.emitJump(OpJump, node.Line)
		c.loopStack[len(c.loopStack)-1].continueJumps = append(c.loopStack[len(c.loopStack)-1].continueJumps, jumpIdx)

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
		} else if idx >= 0 {
			c.emit(OpGetGlobal, idx, node.Line)
		} else if freeIdx, ok := c.freeVarSet[node.Value]; ok {
			c.emit(OpGetFree, freeIdx, node.Line)
		} else if c.isParentLocal(node.Value) {
			freeIdx := len(c.freeVars)
			c.freeVars = append(c.freeVars, node.Value)
			c.freeVarSet[node.Value] = freeIdx
			c.emit(OpGetFree, freeIdx, node.Line)
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
	var jumpEnds []int

	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}
	jumpIfFalse := c.emitJump(OpJumpIfFalse, node.Line)

	if err := c.compileNode(node.Children[1]); err != nil {
		return err
	}
	jumpEnd := c.emitJump(OpJump, node.Line)
	jumpEnds = append(jumpEnds, jumpEnd)

	c.patchJump(jumpIfFalse)

	for childIdx := 2; childIdx < len(node.Children); childIdx++ {
		child := node.Children[childIdx]
		if child.Type == NodeIfStmt {
			if err := c.compileNode(child.Children[0]); err != nil {
				return err
			}
			jumpIfFalse = c.emitJump(OpJumpIfFalse, child.Line)
			if err := c.compileNode(child.Children[1]); err != nil {
				return err
			}
			jumpEnd = c.emitJump(OpJump, child.Line)
			jumpEnds = append(jumpEnds, jumpEnd)
			c.patchJump(jumpIfFalse)
		} else if child.Type == NodeBlockStmt {
			if err := c.compileNode(child); err != nil {
				return err
			}
		}
	}

	for _, je := range jumpEnds {
		c.patchJump(je)
	}

	return nil
}

func (c *Compiler) compileWhile(node *Node) error {
	loopStart := len(c.instructions)

	ctx := loopContext{loopStart: loopStart}
	c.loopStack = append(c.loopStack, ctx)

	if err := c.compileNode(node.Children[0]); err != nil {
		return err
	}

	jumpIfFalse := c.emitJump(OpJumpIfFalse, node.Line)

	if err := c.compileLoopBody(node.Children[1]); err != nil {
		return err
	}

	c.emit(OpLoop, loopStart, node.Line)
	c.patchJump(jumpIfFalse)

	loopEnd := len(c.instructions)
	curCtx := c.loopStack[len(c.loopStack)-1]
	c.loopStack = c.loopStack[:len(c.loopStack)-1]

	for _, j := range curCtx.breakJumps {
		c.instructions[j].Arg = loopEnd
	}
	for _, j := range curCtx.continueJumps {
		c.instructions[j].Arg = loopStart
	}

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
	iterLenIdx := c.defineGlobal("__iter_len")
	c.emit(OpGetGlobal, iterLenIdx, node.Line)
	c.emit(OpGetLocal, iterIdx, node.Line)
	c.emit(OpCall, 1, node.Line)
	c.emit(OpLt, 0, node.Line)

	jumpIfFalse := c.emitJump(OpJumpIfFalse, node.Line)

	ctx := loopContext{loopStart: loopStart}
	c.loopStack = append(c.loopStack, ctx)

	if err := c.compileLoopBody(node.Children[1]); err != nil {
		return err
	}

	incrementStart := len(c.instructions)

	c.emit(OpGetLocal, counterIdx, node.Line)
	oneIdx := c.addConstant(IntObject(1))
	c.emit(OpConstant, oneIdx, node.Line)
	c.emit(OpAdd, 0, node.Line)
	c.emit(OpSetLocal, counterIdx, node.Line)

	c.emit(OpLoop, loopStart, node.Line)
	c.patchJump(jumpIfFalse)

	loopEnd := len(c.instructions)
	curCtx := c.loopStack[len(c.loopStack)-1]
	c.loopStack = c.loopStack[:len(c.loopStack)-1]

	for _, j := range curCtx.breakJumps {
		c.instructions[j].Arg = loopEnd
	}
	for _, j := range curCtx.continueJumps {
		c.instructions[j].Arg = incrementStart
	}

	return nil
}

func (c *Compiler) compileFnDecl(node *Node) error {
	fn := c.compileFnBody(node)
	idx := c.addConstant(fn)
	if len(fn.FreeVars) > 0 {
		c.emit(OpConstant, idx, node.Line)
		c.emit(OpClosure, 0, node.Line)
	} else {
		c.emit(OpConstant, idx, node.Line)
	}
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
	if len(fn.FreeVars) > 0 {
		c.emit(OpConstant, idx, node.Line)
		c.emit(OpClosure, 0, node.Line)
	} else {
		c.emit(OpConstant, idx, node.Line)
	}
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
	subCompiler.globalNames = c.globalNames
	subCompiler.globals = c.globals
	subCompiler.parentLocals = c.allLocalScopes()
	subCompiler.pushScope()

	for _, p := range params {
		subCompiler.defineLocal(p)
	}

	if bodyNode.Type == NodeBlockStmt {
		for _, child := range bodyNode.Children {
			if err := subCompiler.compileNode(child); err != nil {
				subCompiler.emit(OpNilReturn, 0, 0)
				break
			}
		}
	} else {
		if err := subCompiler.compileNode(bodyNode); err != nil {
			subCompiler.emit(OpNilReturn, 0, 0)
		}
	}

	subCompiler.emit(OpNilReturn, 0, 0)

	freeVars := subCompiler.freeVars

	locals := make(map[string]int)
	for _, scope := range subCompiler.localScopes {
		for name, idx := range scope {
			locals[name] = idx
		}
	}

	totalLocals := len(params)
	for _, count := range subCompiler.localCounts[1:] {
		totalLocals += count
	}

	return &FnObject{
		Name:         node.Value,
		Params:       params,
		Locals:       locals,
		Instructions: subCompiler.Instructions(),
		Constants:    subCompiler.Constants(),
		NumLocals:    totalLocals,
		FreeVars:     freeVars,
	}
}
