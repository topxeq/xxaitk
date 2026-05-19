package script

import (
	"fmt"
)

type ObjectType string

const (
	ObjNil     ObjectType = "nil"
	ObjBool    ObjectType = "bool"
	ObjInt     ObjectType = "int"
	ObjFloat   ObjectType = "float"
	ObjString  ObjectType = "string"
	ObjList    ObjectType = "list"
	ObjMap     ObjectType = "map"
	ObjFn      ObjectType = "fn"
	ObjBuiltin ObjectType = "builtin"
	ObjError  ObjectType = "error"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type NilObject struct{}

func (n NilObject) Type() ObjectType { return ObjNil }
func (n NilObject) Inspect() string  { return "nil" }

type BoolObject bool

func (b BoolObject) Type() ObjectType { return ObjBool }
func (b BoolObject) Inspect() string {
	if bool(b) {
		return "true"
	}
	return "false"
}

type IntObject int64

func (i IntObject) Type() ObjectType { return ObjInt }
func (i IntObject) Inspect() string  { return fmt.Sprintf("%d", int64(i)) }

type FloatObject float64

func (f FloatObject) Type() ObjectType { return ObjFloat }
func (f FloatObject) Inspect() string  { return fmt.Sprintf("%g", float64(f)) }

type StringObject string

func (s StringObject) Type() ObjectType { return ObjString }
func (s StringObject) Inspect() string  { return string(s) }

type ListObject struct {
	Elements []Object
}

func (l ListObject) Type() ObjectType { return ObjList }
func (l ListObject) Inspect() string {
	s := "["
	for i, e := range l.Elements {
		if i > 0 {
			s += ", "
		}
		s += inspectValue(e)
	}
	s += "]"
	return s
}

type MapObject struct {
	Pairs map[string]Object
	Keys  []string
}

func (m MapObject) Type() ObjectType { return ObjMap }
func (m MapObject) Inspect() string {
	s := "{"
	for i, k := range m.Keys {
		if i > 0 {
			s += ", "
		}
		s += k + ": " + inspectValue(m.Pairs[k])
	}
	s += "}"
	return s
}

type FnObject struct {
	Name         string
	Params       []string
	Instructions []Instruction
	Constants    []Object
	NumLocals    int
	FreeVars     []string
	Closure      []Object
}

func (f FnObject) Type() ObjectType { return ObjFn }
func (f FnObject) Inspect() string  { return "fn " + f.Name }

type BuiltinFn struct {
	Name string
	Fn   func(...Object) Object
}

func (b BuiltinFn) Type() ObjectType { return ObjBuiltin }
func (b BuiltinFn) Inspect() string  { return "builtin:" + b.Name }

type ErrorObject struct {
	Message string
}

func (e ErrorObject) Type() ObjectType { return ObjError }
func (e ErrorObject) Inspect() string  { return "error: " + e.Message }

func inspectValue(obj Object) string {
	if obj == nil {
		return "nil"
	}
	switch obj.Type() {
	case ObjString:
		return "\"" + obj.Inspect() + "\""
	default:
		return obj.Inspect()
	}
}

func IsTruthy(obj Object) bool {
	switch obj.Type() {
	case ObjNil:
		return false
	case ObjBool:
		return bool(obj.(BoolObject))
	case ObjInt:
		return obj.(IntObject) != 0
	case ObjFloat:
		return obj.(FloatObject) != 0.0
	case ObjString:
		return obj.(StringObject) != ""
	case ObjList:
		return len(obj.(ListObject).Elements) > 0
	case ObjMap:
		return len(obj.(MapObject).Pairs) > 0
	default:
		return true
	}
}
