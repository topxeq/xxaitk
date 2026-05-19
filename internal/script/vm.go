package script

import (
	"fmt"
	"math"
)

type CallFrame struct {
	Fn          *FnObject
	IP          int
	BasePointer int
}

type VM struct {
	stack       []Object
	globals     map[string]Object
	globalNames []string
	frames      []CallFrame
	outputs     []string
	builtins    map[string]BuiltinFn
	unsafe      bool
	maxOps      int
	opCount     int
	printFn     func(string)
}

func NewVM(builtins map[string]BuiltinFn, unsafe bool) *VM {
	vm := &VM{
		stack:       []Object{},
		globals:     make(map[string]Object),
		globalNames: []string{},
		frames:      []CallFrame{},
		outputs:     []string{},
		builtins:    builtins,
		unsafe:      unsafe,
		maxOps:      1000000,
	}
	vm.printFn = func(s string) {
		vm.outputs = append(vm.outputs, s)
	}
	return vm
}

func NewVMWithGlobals(builtins map[string]BuiltinFn, unsafe bool, globalNames []string) *VM {
	vm := NewVM(builtins, unsafe)
	vm.globalNames = globalNames
	return vm
}

func (vm *VM) Run(instructions []Instruction, constants []Object) (Object, error) {
	frame := CallFrame{
		Fn: &FnObject{
			Instructions: instructions,
			Constants:    constants,
			NumLocals:    0,
		},
		IP:          0,
		BasePointer: 0,
	}
	vm.frames = append(vm.frames, frame)

	_, err := vm.execute()
	if err != nil {
		return nil, err
	}

	if len(vm.stack) > 0 {
		return vm.stack[len(vm.stack)-1], nil
	}
	return NilObject{}, nil
}

func (vm *VM) Outputs() []string {
	return vm.outputs
}

func (vm *VM) OpCount() int {
	return vm.opCount
}

func (vm *VM) AddOutput(s string) {
	vm.outputs = append(vm.outputs, s)
}

func (vm *VM) currentFn() *FnObject {
	if len(vm.frames) == 0 {
		return nil
	}
	return vm.frames[len(vm.frames)-1].Fn
}

func (vm *VM) resolveVarInCurrentFrame(name string) (bool, int, bool) {
	frame := vm.currentFn()
	if frame == nil {
		return false, -1, false
	}
	for i, p := range frame.Params {
		if p == name {
			return true, i, false
		}
	}
	return false, -1, false
}

func (vm *VM) execute() (Object, error) {
	for {
		if vm.opCount > vm.maxOps {
			return nil, fmt.Errorf("execution limit exceeded (%d operations)", vm.maxOps)
		}
		vm.opCount++

		frame := &vm.frames[len(vm.frames)-1]
		if frame.IP >= len(frame.Fn.Instructions) {
			break
		}

		inst := frame.Fn.Instructions[frame.IP]
		frame.IP++

		switch inst.OpCode {
		case OpConstant:
			obj := frame.Fn.Constants[inst.Arg]
			vm.push(obj)

		case OpInt:
			vm.push(IntObject(inst.Arg))

		case OpNil:
			vm.push(NilObject{})

		case OpTrue:
			vm.push(BoolObject(true))

		case OpFalse:
			vm.push(BoolObject(false))

		case OpAdd:
			b, a := vm.pop2()
			vm.push(vm.applyArithmetic(a, b, "+"))

		case OpSub:
			b, a := vm.pop2()
			vm.push(vm.applyArithmetic(a, b, "-"))

		case OpMul:
			b, a := vm.pop2()
			vm.push(vm.applyArithmetic(a, b, "*"))

		case OpDiv:
			b, a := vm.pop2()
			vm.push(vm.applyArithmetic(a, b, "/"))

		case OpMod:
			b, a := vm.pop2()
			ai, aok := toInt(a)
			bi, bok := toInt(b)
			if aok && bok {
				if bi == 0 {
					vm.push(NilObject{})
				} else {
					vm.push(IntObject(ai % bi))
				}
			} else {
				vm.push(NilObject{})
			}

		case OpPow:
			b, a := vm.pop2()
			af, aok := toFloat(a)
			bf, bok := toFloat(b)
			if aok && bok {
				vm.push(FloatObject(math.Pow(af, bf)))
			} else {
				vm.push(NilObject{})
			}

		case OpEq:
			b, a := vm.pop2()
			vm.push(BoolObject(objectEquals(a, b)))

		case OpNeq:
			b, a := vm.pop2()
			vm.push(BoolObject(!objectEquals(a, b)))

		case OpLt:
			b, a := vm.pop2()
			vm.push(BoolObject(compareObjects(a, b) < 0))

		case OpGt:
			b, a := vm.pop2()
			vm.push(BoolObject(compareObjects(a, b) > 0))

		case OpLte:
			b, a := vm.pop2()
			vm.push(BoolObject(compareObjects(a, b) <= 0))

		case OpGte:
			b, a := vm.pop2()
			vm.push(BoolObject(compareObjects(a, b) >= 0))

		case OpAnd:
			b, a := vm.pop2()
			if IsTruthy(a) {
				vm.push(b)
			} else {
				vm.push(a)
			}

		case OpOr:
			b, a := vm.pop2()
			if IsTruthy(a) {
				vm.push(a)
			} else {
				vm.push(b)
			}

		case OpNot:
			a := vm.pop()
			vm.push(BoolObject(!IsTruthy(a)))

		case OpNegate:
			a := vm.pop()
			switch v := a.(type) {
			case IntObject:
				vm.push(IntObject(-v))
			case FloatObject:
				vm.push(FloatObject(-v))
			default:
				vm.push(NilObject{})
			}

		case OpIndex:
			idx := vm.pop()
			obj := vm.pop()
			vm.push(vm.indexObject(obj, idx))

		case OpDot:
			idx := vm.pop()
			obj := vm.pop()
			vm.push(vm.indexObject(obj, idx))

		case OpJump:
			frame.IP = inst.Arg

		case OpJumpIfFalse:
			cond := vm.pop()
			if !IsTruthy(cond) {
				frame.IP = inst.Arg
			}

		case OpJumpIfTrue:
			cond := vm.pop()
			if IsTruthy(cond) {
				frame.IP = inst.Arg
			}

		case OpLoop:
			frame.IP = inst.Arg

		case OpGetLocal:
			idx := vm.frameBase() + inst.Arg
			if idx < len(vm.stack) {
				vm.push(vm.stack[idx])
			} else {
				vm.push(NilObject{})
			}

		case OpSetLocal:
			idx := vm.frameBase() + inst.Arg
			val := vm.pop()
			for len(vm.stack) <= idx {
				vm.stack = append(vm.stack, NilObject{})
			}
			vm.stack[idx] = val

		case OpGetGlobal:
			name := ""
			if inst.Arg < len(vm.globalNames) {
				name = vm.globalNames[inst.Arg]
			}
			if val, ok := vm.globals[name]; ok {
				vm.push(val)
			} else if b, ok := vm.builtins[name]; ok {
				vm.push(b)
			} else {
				vm.push(NilObject{})
			}

		case OpSetGlobal:
			name := ""
			if inst.Arg < len(vm.globalNames) {
				name = vm.globalNames[inst.Arg]
			}
			val := vm.pop()
			vm.globals[name] = val

		case OpDefineGlobal:
			name := ""
			if inst.Arg < len(vm.globalNames) {
				name = vm.globalNames[inst.Arg]
			}
			val := vm.pop()
			vm.globals[name] = val

		case OpDefineConst:
			name := ""
			if inst.Arg < len(vm.globalNames) {
				name = vm.globalNames[inst.Arg]
			}
			val := vm.pop()
			vm.globals[name] = val

		case OpGetFree:
			fn := vm.currentFn()
			if fn != nil && inst.Arg < len(fn.Closure) {
				vm.push(fn.Closure[inst.Arg])
			} else {
				vm.push(NilObject{})
			}

		case OpClosure:
			fnObj := vm.pop()
			if fn, ok := fnObj.(*FnObject); ok {
				closure := make([]Object, len(fn.FreeVars))
				for i, varName := range fn.FreeVars {
					isLocal, idx, _ := vm.resolveVarInCurrentFrame(varName)
					if isLocal {
						stackIdx := vm.frameBase() + idx
						if stackIdx < len(vm.stack) {
							closure[i] = vm.stack[stackIdx]
						} else {
							closure[i] = NilObject{}
						}
					} else if val, ok := vm.globals[varName]; ok {
						closure[i] = val
					} else {
						closure[i] = NilObject{}
					}
				}
				closedFn := &FnObject{
					Name:         fn.Name,
					Params:       fn.Params,
					Instructions: fn.Instructions,
					Constants:    fn.Constants,
					NumLocals:    fn.NumLocals,
					FreeVars:     fn.FreeVars,
					Closure:      closure,
				}
				vm.push(closedFn)
			} else {
				vm.push(fnObj)
			}

		case OpCall:
			argCount := inst.Arg
			args := make([]Object, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			fn := vm.pop()

			switch f := fn.(type) {
			case *FnObject:
				newFrame := CallFrame{
					Fn:          f,
					IP:          0,
					BasePointer: len(vm.stack),
				}
				for _, arg := range args {
					vm.push(arg)
				}
				for len(vm.stack) < newFrame.BasePointer+f.NumLocals {
					vm.push(NilObject{})
				}
				vm.frames = append(vm.frames, newFrame)
				frame = &vm.frames[len(vm.frames)-1]

			case BuiltinFn:
				result := f.Fn(args...)
				vm.push(result)

			default:
				return nil, fmt.Errorf("not callable: %s", fn.Inspect())
			}

		case OpReturn:
			result := vm.pop()
			if len(vm.frames) <= 1 {
				vm.push(result)
				return result, nil
			}
			calledFrame := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]
			vm.stack = vm.stack[:calledFrame.BasePointer]
			vm.push(result)

		case OpNilReturn:
			if len(vm.frames) <= 1 {
				vm.push(NilObject{})
				return NilObject{}, nil
			}
			calledFrame := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]
			vm.stack = vm.stack[:calledFrame.BasePointer]
			vm.push(NilObject{})

		case OpPop:
			if len(vm.stack) > 0 {
				vm.pop()
			}

		case OpBuildList:
			elements := make([]Object, inst.Arg)
			for i := inst.Arg - 1; i >= 0; i-- {
				elements[i] = vm.pop()
			}
			vm.push(ListObject{Elements: elements})

		case OpBuildMap:
			pairs := make(map[string]Object)
			keys := make([]string, 0, inst.Arg)
			for i := 0; i < inst.Arg; i++ {
				val := vm.pop()
				keyObj := vm.pop()
				key := keyObj.Inspect()
				pairs[key] = val
				keys = append(keys, key)
			}
			for i, j := 0, len(keys)-1; i < j; i, j = i+1, j-1 {
				keys[i], keys[j] = keys[j], keys[i]
			}
			vm.push(MapObject{Pairs: pairs, Keys: keys})

		case OpBreakpoint:
		}
	}

	if len(vm.stack) > 0 {
		return vm.stack[len(vm.stack)-1], nil
	}
	return NilObject{}, nil
}

func (vm *VM) push(obj Object) {
	vm.stack = append(vm.stack, obj)
}

func (vm *VM) pop() Object {
	if len(vm.stack) == 0 {
		return NilObject{}
	}
	obj := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return obj
}

func (vm *VM) pop2() (Object, Object) {
	b := vm.pop()
	a := vm.pop()
	return b, a
}

func (vm *VM) frameBase() int {
	if len(vm.frames) == 0 {
		return 0
	}
	return vm.frames[len(vm.frames)-1].BasePointer
}

func (vm *VM) applyArithmetic(a, b Object, op string) Object {
	if as, aok := a.(StringObject); aok {
		if bs, bok := b.(StringObject); bok && op == "+" {
			return StringObject(string(as) + string(bs))
		}
	}

	ai, aIsInt := toInt(a)
	bi, bIsInt := toInt(b)

	if aIsInt && bIsInt {
		switch op {
		case "+":
			return IntObject(ai + bi)
		case "-":
			return IntObject(ai - bi)
		case "*":
			return IntObject(ai * bi)
		case "/":
			if bi == 0 {
				return NilObject{}
			}
			return IntObject(ai / bi)
		}
	}

	af, aIsFloat := toFloat(a)
	bf, bIsFloat := toFloat(b)

	if aIsFloat && bIsFloat {
		switch op {
		case "+":
			return FloatObject(af + bf)
		case "-":
			return FloatObject(af - bf)
		case "*":
			return FloatObject(af * bf)
		case "/":
			if bf == 0 {
				return NilObject{}
			}
			return FloatObject(af / bf)
		}
	}

	return NilObject{}
}

func (vm *VM) indexObject(obj, idx Object) Object {
	switch o := obj.(type) {
	case ListObject:
		if i, ok := toInt(idx); ok {
			elems := o.Elements
			if i >= 0 && int(i) < len(elems) {
				return elems[i]
			}
		}
		return NilObject{}
	case MapObject:
		key := idx.Inspect()
		if val, ok := o.Pairs[key]; ok {
			return val
		}
		return NilObject{}
	case StringObject:
		if i, ok := toInt(idx); ok {
			s := string(o)
			if i >= 0 && int(i) < len(s) {
				return StringObject(string(s[i]))
			}
		}
		return NilObject{}
	default:
		return NilObject{}
	}
}

func toInt(obj Object) (int64, bool) {
	switch v := obj.(type) {
	case IntObject:
		return int64(v), true
	case FloatObject:
		return int64(v), true
	default:
		return 0, false
	}
}

func toFloat(obj Object) (float64, bool) {
	switch v := obj.(type) {
	case IntObject:
		return float64(v), true
	case FloatObject:
		return float64(v), true
	default:
		return 0, false
	}
}

func objectEquals(a, b Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case NilObject:
		return true
	case BoolObject:
		return av == b.(BoolObject)
	case IntObject:
		return av == b.(IntObject)
	case FloatObject:
		return float64(av) == float64(b.(FloatObject))
	case StringObject:
		return string(av) == string(b.(StringObject))
	default:
		return false
	}
}

func compareObjects(a, b Object) int {
	at, bt := a.Type(), b.Type()
	if at != bt {
		if at == ObjInt && bt == ObjFloat {
			af := float64(a.(IntObject))
			bf := float64(b.(FloatObject))
			return compareFloats(af, bf)
		}
		if at == ObjFloat && bt == ObjInt {
			af := float64(a.(FloatObject))
			bf := float64(b.(IntObject))
			return compareFloats(af, bf)
		}
		return -1
	}

	switch av := a.(type) {
	case IntObject:
		bv := b.(IntObject)
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
		return 0
	case FloatObject:
		bv := b.(FloatObject)
		return compareFloats(float64(av), float64(bv))
	case StringObject:
		bv := b.(StringObject)
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func compareFloats(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
