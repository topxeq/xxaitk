package script

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var allBuiltins = map[string]BuiltinFn{}
var PrintCallback func(string)

func registerBuiltin(name string, fn func(...Object) Object) {
	allBuiltins[name] = BuiltinFn{Name: name, Fn: fn}
}

func init() {
	registerStrBuiltins()
	registerMathBuiltins()
	registerListBuiltins()
	registerMapBuiltins()
	registerJSONBuiltins()
	registerIOBuiltins()
	registerNetBuiltins()
	registerTypeBuiltins()
	registerConvBuiltins()
	registerLogBuiltins()
	registerTimeBuiltins()
	registerErrorHandlingBuiltins()
	registerOSBuiltins()
}

var safeBuiltins = map[string]bool{}

func init() {
	safe := []string{
		"print", "str_len", "str_concat", "str_split", "str_join", "str_sub",
		"str_trim", "str_upper", "str_lower", "str_replace", "str_has_prefix",
		"str_has_suffix", "str_contains", "str_index", "str_from_int",
		"str_from_float", "str_to_int", "str_to_float", "str_repeat",
		"str_reverse", "str_pad_left", "str_pad_right", "str_interp",
		"math_abs", "math_max", "math_min", "math_floor", "math_ceil",
		"math_round", "math_sqrt", "math_pow", "math_mod", "math_rand",
		"math_rand_int", "math_log", "math_exp", "math_sin", "math_cos",
		"list_len", "list_push", "list_pop", "list_shift", "list_get",
		"list_set", "list_contains", "list_index", "list_join", "list_map",
		"list_filter", "list_sort", "list_reverse", "list_slice", "list_flat",
		"list_reduce", "list_find",
		"map_get", "map_set", "map_has", "map_keys", "map_values",
		"map_del", "map_len", "map_merge",
		"json_encode", "json_decode", "json_get", "json_set", "json_has",
		"io_read_file", "io_exists", "io_is_dir", "io_is_file",
		"io_list_dir", "io_size", "io_mkdir",
		"net_http_get", "net_dns_lookup",
		"type_of", "type_is_nil", "type_is_bool", "type_is_int",
		"type_is_float", "type_is_string", "type_is_list", "type_is_map",
		"type_is_fn",
		"conv_to_int", "conv_to_float", "conv_to_string", "conv_to_bool",
		"conv_to_list",
		"conv_hex_encode", "conv_hex_decode", "conv_b64_encode", "conv_b64_decode",
		"log_info", "log_warn", "log_error", "log_debug",
		"time_now", "time_now_unix", "time_format", "time_parse", "time_sleep", "time_duration",
		"try", "catch", "is_error",
	}
	for _, name := range safe {
		safeBuiltins[name] = true
	}
}

func GetBuiltins(unsafe bool) map[string]BuiltinFn {
	result := map[string]BuiltinFn{}
	for name, fn := range allBuiltins {
		if unsafe || safeBuiltins[name] {
			result[name] = fn
		}
	}
	return result
}

func registerStrBuiltins() {
	registerBuiltin("print", func(args ...Object) Object {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Inspect()
		}
		output := strings.Join(parts, " ")
		if PrintCallback != nil {
			PrintCallback(output)
		}
		return NilObject{}
	})
	registerBuiltin("str_len", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		return IntObject(len(args[0].Inspect()))
	})
	registerBuiltin("str_concat", func(args ...Object) Object {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Inspect()
		}
		return StringObject(strings.Join(parts, ""))
	})
	registerBuiltin("str_split", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		s := args[0].Inspect()
		sep := ""
		if len(args) >= 2 {
			sep = args[1].Inspect()
		}
		if sep == "" {
			sep = " "
		}
		parts := strings.Split(s, sep)
		elems := make([]Object, len(parts))
		for i, p := range parts {
			elems[i] = StringObject(p)
		}
		return ListObject{Elements: elems}
	})
	registerBuiltin("str_join", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		sep := ""
		if len(args) >= 2 {
			sep = args[1].Inspect()
		}
		if list, ok := args[0].(ListObject); ok {
			parts := make([]string, len(list.Elements))
			for i, e := range list.Elements {
				parts[i] = e.Inspect()
			}
			return StringObject(strings.Join(parts, sep))
		}
		return StringObject("")
	})
	registerBuiltin("str_sub", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		s := args[0].Inspect()
		start := 0
		if len(args) >= 2 {
			if i, ok := toInt(args[1]); ok && int(i) < len(s) {
				start = int(i)
			}
		}
		end := len(s)
		if len(args) >= 3 {
			if i, ok := toInt(args[2]); ok && int(i) <= len(s) {
				end = int(i)
			}
		}
		if start > end {
			start = end
		}
		return StringObject(s[start:end])
	})
	registerBuiltin("str_trim", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		return StringObject(strings.TrimSpace(args[0].Inspect()))
	})
	registerBuiltin("str_upper", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		return StringObject(strings.ToUpper(args[0].Inspect()))
	})
	registerBuiltin("str_lower", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		return StringObject(strings.ToLower(args[0].Inspect()))
	})
	registerBuiltin("str_replace", func(args ...Object) Object {
		if len(args) < 3 {
			return StringObject("")
		}
		s := args[0].Inspect()
		old := args[1].Inspect()
		new_ := args[2].Inspect()
		n := -1
		if len(args) >= 4 {
			if i, ok := toInt(args[3]); ok {
				n = int(i)
			}
		}
		return StringObject(strings.Replace(s, old, new_, n))
	})
	registerBuiltin("str_has_prefix", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		return BoolObject(strings.HasPrefix(args[0].Inspect(), args[1].Inspect()))
	})
	registerBuiltin("str_has_suffix", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		return BoolObject(strings.HasSuffix(args[0].Inspect(), args[1].Inspect()))
	})
	registerBuiltin("str_contains", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		return BoolObject(strings.Contains(args[0].Inspect(), args[1].Inspect()))
	})
	registerBuiltin("str_index", func(args ...Object) Object {
		if len(args) < 2 {
			return IntObject(-1)
		}
		return IntObject(strings.Index(args[0].Inspect(), args[1].Inspect()))
	})
	registerBuiltin("str_from_int", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		if i, ok := toInt(args[0]); ok {
			return StringObject(fmt.Sprintf("%d", i))
		}
		return StringObject("")
	})
	registerBuiltin("str_from_float", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		if f, ok := toFloat(args[0]); ok {
			return StringObject(fmt.Sprintf("%g", f))
		}
		return StringObject("")
	})
	registerBuiltin("str_to_int", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		var i int64
		if _, err := fmt.Sscanf(args[0].Inspect(), "%d", &i); err == nil {
			return IntObject(i)
		}
		return IntObject(0)
	})
	registerBuiltin("str_to_float", func(args ...Object) Object {
		if len(args) < 1 {
			return FloatObject(0)
		}
		var f float64
		if _, err := fmt.Sscanf(args[0].Inspect(), "%f", &f); err == nil {
			return FloatObject(f)
		}
		return FloatObject(0)
	})
	registerBuiltin("str_repeat", func(args ...Object) Object {
		if len(args) < 2 {
			return StringObject("")
		}
		s := args[0].Inspect()
		n, _ := toInt(args[1])
		return StringObject(strings.Repeat(s, int(n)))
	})
	registerBuiltin("str_reverse", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		s := args[0].Inspect()
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return StringObject(string(runes))
	})
	registerBuiltin("str_pad_left", func(args ...Object) Object {
		if len(args) < 2 {
			return StringObject("")
		}
		s := args[0].Inspect()
		n, _ := toInt(args[1])
		pad := " "
		if len(args) >= 3 {
			pad = args[2].Inspect()
		}
		for len(s) < int(n) {
			s = pad + s
		}
		return StringObject(s)
	})
	registerBuiltin("str_pad_right", func(args ...Object) Object {
		if len(args) < 2 {
			return StringObject("")
		}
		s := args[0].Inspect()
		n, _ := toInt(args[1])
		pad := " "
		if len(args) >= 3 {
			pad = args[2].Inspect()
		}
		for len(s) < int(n) {
			s = s + pad
		}
		return StringObject(s)
	})
	registerBuiltin("str_interp", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		tmpl := args[0].Inspect()
		if len(args) >= 2 {
			if m, ok := args[1].(MapObject); ok {
				for k, v := range m.Pairs {
					placeholder := "${" + k + "}"
					tmpl = strings.ReplaceAll(tmpl, placeholder, v.Inspect())
				}
			}
		}
		return StringObject(tmpl)
	})
}

func registerMathBuiltins() {
	registerBuiltin("math_abs", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		if i, ok := toInt(args[0]); ok {
			if i < 0 {
				return IntObject(-i)
			}
			return IntObject(i)
		}
		if f, ok := toFloat(args[0]); ok {
			return FloatObject(math.Abs(f))
		}
		return IntObject(0)
	})
	registerBuiltin("math_max", func(args ...Object) Object {
		if len(args) < 2 {
			return IntObject(0)
		}
		a, _ := toFloat(args[0])
		b, _ := toFloat(args[1])
		return FloatObject(math.Max(a, b))
	})
	registerBuiltin("math_min", func(args ...Object) Object {
		if len(args) < 2 {
			return IntObject(0)
		}
		a, _ := toFloat(args[0])
		b, _ := toFloat(args[1])
		return FloatObject(math.Min(a, b))
	})
	registerBuiltin("math_floor", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		f, _ := toFloat(args[0])
		return IntObject(int64(math.Floor(f)))
	})
	registerBuiltin("math_ceil", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		f, _ := toFloat(args[0])
		return IntObject(int64(math.Ceil(f)))
	})
	registerBuiltin("math_round", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		f, _ := toFloat(args[0])
		return IntObject(int64(math.Round(f)))
	})
	registerBuiltin("math_sqrt", func(args ...Object) Object {
		if len(args) < 1 {
			return FloatObject(0)
		}
		f, _ := toFloat(args[0])
		return FloatObject(math.Sqrt(f))
	})
	registerBuiltin("math_pow", func(args ...Object) Object {
		if len(args) < 2 {
			return FloatObject(0)
		}
		a, _ := toFloat(args[0])
		b, _ := toFloat(args[1])
		return FloatObject(math.Pow(a, b))
	})
	registerBuiltin("math_mod", func(args ...Object) Object {
		if len(args) < 2 {
			return IntObject(0)
		}
		a, _ := toInt(args[0])
		b, _ := toInt(args[1])
		if b == 0 {
			return IntObject(0)
		}
		return IntObject(a % b)
	})
	registerBuiltin("math_rand", func(args ...Object) Object {
		return FloatObject(rand.Float64())
	})
	registerBuiltin("math_rand_int", func(args ...Object) Object {
		n := int64(100)
		if len(args) >= 1 {
			if i, ok := toInt(args[0]); ok {
				n = i
			}
		}
		return IntObject(rand.Int63n(n))
	})
	registerBuiltin("math_log", func(args ...Object) Object {
		if len(args) < 1 {
			return FloatObject(0)
		}
		f, _ := toFloat(args[0])
		return FloatObject(math.Log(f))
	})
	registerBuiltin("math_exp", func(args ...Object) Object {
		if len(args) < 1 {
			return FloatObject(1)
		}
		f, _ := toFloat(args[0])
		return FloatObject(math.Exp(f))
	})
	registerBuiltin("math_sin", func(args ...Object) Object {
		if len(args) < 1 {
			return FloatObject(0)
		}
		f, _ := toFloat(args[0])
		return FloatObject(math.Sin(f))
	})
	registerBuiltin("math_cos", func(args ...Object) Object {
		if len(args) < 1 {
			return FloatObject(1)
		}
		f, _ := toFloat(args[0])
		return FloatObject(math.Cos(f))
	})
}

func registerListBuiltins() {
	registerBuiltin("list_len", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		if list, ok := args[0].(ListObject); ok {
			return IntObject(len(list.Elements))
		}
		return IntObject(0)
	})
	registerBuiltin("list_push", func(args ...Object) Object {
		if len(args) < 2 {
			return args[0]
		}
		if list, ok := args[0].(ListObject); ok {
			return ListObject{Elements: append(list.Elements, args[1])}
		}
		return args[0]
	})
	registerBuiltin("list_pop", func(args ...Object) Object {
		if len(args) < 1 {
			return NilObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			if len(list.Elements) == 0 {
				return NilObject{}
			}
			return list.Elements[len(list.Elements)-1]
		}
		return NilObject{}
	})
	registerBuiltin("list_shift", func(args ...Object) Object {
		if len(args) < 1 {
			return NilObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			if len(list.Elements) == 0 {
				return NilObject{}
			}
			return list.Elements[0]
		}
		return NilObject{}
	})
	registerBuiltin("list_get", func(args ...Object) Object {
		if len(args) < 2 {
			return NilObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			if i, ok := toInt(args[1]); ok && int(i) < len(list.Elements) {
				return list.Elements[i]
			}
		}
		return NilObject{}
	})
	registerBuiltin("list_set", func(args ...Object) Object {
		if len(args) < 3 {
			return args[0]
		}
		if list, ok := args[0].(ListObject); ok {
			if i, ok := toInt(args[1]); ok && int(i) < len(list.Elements) {
				elems := make([]Object, len(list.Elements))
				copy(elems, list.Elements)
				elems[i] = args[2]
				return ListObject{Elements: elems}
			}
		}
		return args[0]
	})
	registerBuiltin("list_contains", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		if list, ok := args[0].(ListObject); ok {
			for _, e := range list.Elements {
				if objectEquals(e, args[1]) {
					return BoolObject(true)
				}
			}
		}
		return BoolObject(false)
	})
	registerBuiltin("list_index", func(args ...Object) Object {
		if len(args) < 2 {
			return IntObject(-1)
		}
		if list, ok := args[0].(ListObject); ok {
			for i, e := range list.Elements {
				if objectEquals(e, args[1]) {
					return IntObject(int64(i))
				}
			}
		}
		return IntObject(-1)
	})
	registerBuiltin("list_join", func(args ...Object) Object {
		if len(args) < 2 {
			return StringObject("")
		}
		if list, ok := args[0].(ListObject); ok {
			sep := args[1].Inspect()
			parts := make([]string, len(list.Elements))
			for i, e := range list.Elements {
				parts[i] = e.Inspect()
			}
			return StringObject(strings.Join(parts, sep))
		}
		return StringObject("")
	})
	registerBuiltin("list_map", func(args ...Object) Object {
		if len(args) < 2 {
			return ListObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			if fn, ok := args[1].(BuiltinFn); ok {
				result := make([]Object, len(list.Elements))
				for i, e := range list.Elements {
					result[i] = fn.Fn(e)
				}
				return ListObject{Elements: result}
			}
		}
		return ListObject{}
	})
	registerBuiltin("list_filter", func(args ...Object) Object {
		if len(args) < 2 {
			return ListObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			if fn, ok := args[1].(BuiltinFn); ok {
				var result []Object
				for _, e := range list.Elements {
					if IsTruthy(fn.Fn(e)) {
						result = append(result, e)
					}
				}
				return ListObject{Elements: result}
			}
		}
		return ListObject{}
	})
	registerBuiltin("list_sort", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			elems := make([]Object, len(list.Elements))
			copy(elems, list.Elements)
			for i := 0; i < len(elems); i++ {
				for j := i + 1; j < len(elems); j++ {
					if compareObjects(elems[i], elems[j]) > 0 {
						elems[i], elems[j] = elems[j], elems[i]
					}
				}
			}
			return ListObject{Elements: elems}
		}
		return ListObject{}
	})
	registerBuiltin("list_reverse", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			elems := make([]Object, len(list.Elements))
			copy(elems, list.Elements)
			for i, j := 0, len(elems)-1; i < j; i, j = i+1, j-1 {
				elems[i], elems[j] = elems[j], elems[i]
			}
			return ListObject{Elements: elems}
		}
		return ListObject{}
	})
	registerBuiltin("list_slice", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			start := 0
			end := len(list.Elements)
			if len(args) >= 2 {
				if i, ok := toInt(args[1]); ok {
					start = int(i)
				}
			}
			if len(args) >= 3 {
				if i, ok := toInt(args[2]); ok {
					end = int(i)
				}
			}
			if start < 0 {
				start = 0
			}
			if end > len(list.Elements) {
				end = len(list.Elements)
			}
			if start > end {
				start = end
			}
			return ListObject{Elements: list.Elements[start:end]}
		}
		return ListObject{}
	})
	registerBuiltin("list_flat", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			var result []Object
			for _, e := range list.Elements {
				if inner, ok := e.(ListObject); ok {
					result = append(result, inner.Elements...)
				} else {
					result = append(result, e)
				}
			}
			return ListObject{Elements: result}
		}
		return ListObject{}
	})
	registerBuiltin("list_reduce", func(args ...Object) Object {
		if len(args) < 3 {
			return NilObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			if fn, ok := args[1].(BuiltinFn); ok {
				acc := args[2]
				for _, e := range list.Elements {
					acc = fn.Fn(acc, e)
				}
				return acc
			}
		}
		return NilObject{}
	})
	registerBuiltin("list_find", func(args ...Object) Object {
		if len(args) < 2 {
			return NilObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			if fn, ok := args[1].(BuiltinFn); ok {
				for _, e := range list.Elements {
					if IsTruthy(fn.Fn(e)) {
						return e
					}
				}
			}
		}
		return NilObject{}
	})
}

func registerMapBuiltins() {
	registerBuiltin("map_get", func(args ...Object) Object {
		if len(args) < 2 {
			return NilObject{}
		}
		if m, ok := args[0].(MapObject); ok {
			key := args[1].Inspect()
			if val, ok := m.Pairs[key]; ok {
				return val
			}
		}
		if len(args) >= 3 {
			return args[2]
		}
		return NilObject{}
	})
	registerBuiltin("map_set", func(args ...Object) Object {
		if len(args) < 3 {
			return args[0]
		}
		if m, ok := args[0].(MapObject); ok {
			pairs := make(map[string]Object)
			for k, v := range m.Pairs {
				pairs[k] = v
			}
			keys := make([]string, len(m.Keys))
			copy(keys, m.Keys)
			key := args[1].Inspect()
			if _, exists := pairs[key]; !exists {
				keys = append(keys, key)
			}
			pairs[key] = args[2]
			return MapObject{Pairs: pairs, Keys: keys}
		}
		return args[0]
	})
	registerBuiltin("map_has", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		if m, ok := args[0].(MapObject); ok {
			_, ok := m.Pairs[args[1].Inspect()]
			return BoolObject(ok)
		}
		return BoolObject(false)
	})
	registerBuiltin("map_keys", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		if m, ok := args[0].(MapObject); ok {
			elems := make([]Object, len(m.Keys))
			for i, k := range m.Keys {
				elems[i] = StringObject(k)
			}
			return ListObject{Elements: elems}
		}
		return ListObject{}
	})
	registerBuiltin("map_values", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		if m, ok := args[0].(MapObject); ok {
			elems := make([]Object, len(m.Keys))
			for i, k := range m.Keys {
				elems[i] = m.Pairs[k]
			}
			return ListObject{Elements: elems}
		}
		return ListObject{}
	})
	registerBuiltin("map_del", func(args ...Object) Object {
		if len(args) < 2 {
			return args[0]
		}
		if m, ok := args[0].(MapObject); ok {
			pairs := make(map[string]Object)
			for k, v := range m.Pairs {
				pairs[k] = v
			}
			key := args[1].Inspect()
			delete(pairs, key)
			var keys []string
			for _, k := range m.Keys {
				if k != key {
					keys = append(keys, k)
				}
			}
			return MapObject{Pairs: pairs, Keys: keys}
		}
		return args[0]
	})
	registerBuiltin("map_len", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		if m, ok := args[0].(MapObject); ok {
			return IntObject(len(m.Pairs))
		}
		return IntObject(0)
	})
	registerBuiltin("map_merge", func(args ...Object) Object {
		if len(args) < 2 {
			return args[0]
		}
		if m, ok := args[0].(MapObject); ok {
			if other, ok := args[1].(MapObject); ok {
				pairs := make(map[string]Object)
				for k, v := range m.Pairs {
					pairs[k] = v
				}
				keysMap := map[string]bool{}
				for _, k := range m.Keys {
					keysMap[k] = true
				}
				for k, v := range other.Pairs {
					pairs[k] = v
					keysMap[k] = true
				}
				var keys []string
				for _, k := range m.Keys {
					keys = append(keys, k)
				}
				for _, k := range other.Keys {
					if !keysMap[k] || !containsStr(m.Keys, k) {
						if !containsStr(keys, k) {
							keys = append(keys, k)
						}
					}
				}
				return MapObject{Pairs: pairs, Keys: keys}
			}
		}
		return args[0]
	})
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func registerJSONBuiltins() {
	registerBuiltin("json_encode", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("null")
		}
		v := objectToJSON(args[0])
		b, err := json.Marshal(v)
		if err != nil {
			return StringObject("")
		}
		return StringObject(string(b))
	})
	registerBuiltin("json_decode", func(args ...Object) Object {
		if len(args) < 1 {
			return NilObject{}
		}
		var v interface{}
		if err := json.Unmarshal([]byte(args[0].Inspect()), &v); err != nil {
			return NilObject{}
		}
		return jsonToObject(v)
	})
	registerBuiltin("json_get", func(args ...Object) Object {
		if len(args) < 2 {
			return NilObject{}
		}
		if m, ok := args[0].(MapObject); ok {
			key := args[1].Inspect()
			if val, ok := m.Pairs[key]; ok {
				return val
			}
		}
		return NilObject{}
	})
	registerBuiltin("json_set", func(args ...Object) Object {
		if len(args) < 3 {
			return args[0]
		}
		if m, ok := args[0].(MapObject); ok {
			pairs := make(map[string]Object)
			for k, v := range m.Pairs {
				pairs[k] = v
			}
			keys := make([]string, len(m.Keys))
			copy(keys, m.Keys)
			key := args[1].Inspect()
			if _, exists := pairs[key]; !exists {
				keys = append(keys, key)
			}
			pairs[key] = args[2]
			return MapObject{Pairs: pairs, Keys: keys}
		}
		return args[0]
	})
	registerBuiltin("json_has", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		if m, ok := args[0].(MapObject); ok {
			_, ok := m.Pairs[args[1].Inspect()]
			return BoolObject(ok)
		}
		return BoolObject(false)
	})
}

func objectToJSON(obj Object) interface{} {
	switch v := obj.(type) {
	case NilObject:
		return nil
	case BoolObject:
		return bool(v)
	case IntObject:
		return int64(v)
	case FloatObject:
		return float64(v)
	case StringObject:
		return string(v)
	case ListObject:
		arr := make([]interface{}, len(v.Elements))
		for i, e := range v.Elements {
			arr[i] = objectToJSON(e)
		}
		return arr
	case MapObject:
		m := make(map[string]interface{})
		for k, val := range v.Pairs {
			m[k] = objectToJSON(val)
		}
		return m
	default:
		return nil
	}
}

func jsonToObject(v interface{}) Object {
	switch val := v.(type) {
	case nil:
		return NilObject{}
	case bool:
		return BoolObject(val)
	case float64:
		if val == float64(int64(val)) && val < 1e15 {
			return IntObject(int64(val))
		}
		return FloatObject(val)
	case string:
		return StringObject(val)
	case []interface{}:
		elems := make([]Object, len(val))
		for i, e := range val {
			elems[i] = jsonToObject(e)
		}
		return ListObject{Elements: elems}
	case map[string]interface{}:
		pairs := make(map[string]Object)
		keys := make([]string, 0, len(val))
		for k, v := range val {
			pairs[k] = jsonToObject(v)
			keys = append(keys, k)
		}
		return MapObject{Pairs: pairs, Keys: keys}
	default:
		return NilObject{}
	}
}

func registerIOBuiltins() {
	registerBuiltin("io_read_file", func(args ...Object) Object {
		if len(args) < 1 {
			return ErrorObject{Message: "io_read_file requires a path argument"}
		}
		data, err := os.ReadFile(args[0].Inspect())
		if err != nil {
			return ErrorObject{Message: err.Error()}
		}
		return StringObject(string(data))
	})
	registerBuiltin("io_write_file", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		err := os.WriteFile(args[0].Inspect(), []byte(args[1].Inspect()), 0644)
		return BoolObject(err == nil)
	})
	registerBuiltin("io_append_file", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		f, err := os.OpenFile(args[0].Inspect(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return BoolObject(false)
		}
		defer f.Close()
		_, err = f.WriteString(args[1].Inspect())
		return BoolObject(err == nil)
	})
	registerBuiltin("io_exists", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		_, err := os.Stat(args[0].Inspect())
		return BoolObject(err == nil)
	})
	registerBuiltin("io_is_dir", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		info, err := os.Stat(args[0].Inspect())
		return BoolObject(err == nil && info.IsDir())
	})
	registerBuiltin("io_is_file", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		info, err := os.Stat(args[0].Inspect())
		return BoolObject(err == nil && !info.IsDir())
	})
	registerBuiltin("io_list_dir", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		entries, err := os.ReadDir(args[0].Inspect())
		if err != nil {
			return ListObject{}
		}
		elems := make([]Object, len(entries))
		for i, e := range entries {
			elems[i] = StringObject(e.Name())
		}
		return ListObject{Elements: elems}
	})
	registerBuiltin("io_size", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		info, err := os.Stat(args[0].Inspect())
		if err != nil {
			return IntObject(0)
		}
		return IntObject(info.Size())
	})
	registerBuiltin("io_mkdir", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		perm := os.FileMode(0755)
		if len(args) >= 2 {
			if p, ok := toInt(args[1]); ok {
				perm = os.FileMode(p)
			}
		}
		err := os.MkdirAll(args[0].Inspect(), perm)
		return BoolObject(err == nil)
	})
	registerBuiltin("io_copy", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		src, err := os.ReadFile(args[0].Inspect())
		if err != nil {
			return BoolObject(false)
		}
		err = os.WriteFile(args[1].Inspect(), src, 0644)
		return BoolObject(err == nil)
	})
	registerBuiltin("io_move", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		err := os.Rename(args[0].Inspect(), args[1].Inspect())
		return BoolObject(err == nil)
	})
	registerBuiltin("io_remove", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		err := os.Remove(args[0].Inspect())
		return BoolObject(err == nil)
	})
	registerBuiltin("io_temp_dir", func(args ...Object) Object {
		return StringObject(os.TempDir())
	})
	registerBuiltin("io_abs_path", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		path, err := filepath.Abs(args[0].Inspect())
		if err != nil {
			return StringObject(args[0].Inspect())
		}
		return StringObject(path)
	})
}

func registerNetBuiltins() {
	registerBuiltin("net_http_get", func(args ...Object) Object {
		if len(args) < 1 {
			return NilObject{}
		}
		urlStr := args[0].Inspect()
		insecure := false
		timeout := 30
		if len(args) >= 2 {
			if m, ok := args[1].(MapObject); ok {
				if v, ok := m.Pairs["insecure"]; ok {
					insecure = IsTruthy(v)
				}
				if v, ok := m.Pairs["timeout"]; ok {
					if i, ok := toInt(v); ok {
						timeout = int(i)
					}
				}
			}
		}
		transport := &http.Transport{}
		if insecure {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		client := &http.Client{
			Timeout:   time.Duration(timeout) * time.Second,
			Transport: transport,
		}
		resp, err := client.Get(urlStr)
		if err != nil {
			return NilObject{}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NilObject{}
		}
		return StringObject(string(body))
	})
	registerBuiltin("net_dns_lookup", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		addrs, err := net.LookupHost(args[0].Inspect())
		if err != nil {
			return ListObject{}
		}
		elems := make([]Object, len(addrs))
		for i, a := range addrs {
			elems[i] = StringObject(a)
		}
		return ListObject{Elements: elems}
	})
	registerBuiltin("net_tcp_connect", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		addr := args[0].Inspect()
		timeout := 10 * time.Second
		if len(args) >= 2 {
			if t, ok := toInt(args[1]); ok {
				timeout = time.Duration(t) * time.Second
			}
		}
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return BoolObject(false)
		}
		conn.Close()
		return BoolObject(true)
	})
}

func registerTypeBuiltins() {
	registerBuiltin("type_of", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("nil")
		}
		return StringObject(string(args[0].Type()))
	})
	registerBuiltin("type_is_nil", func(args ...Object) Object {
		return BoolObject(len(args) > 0 && args[0].Type() == ObjNil)
	})
	registerBuiltin("type_is_bool", func(args ...Object) Object {
		return BoolObject(len(args) > 0 && args[0].Type() == ObjBool)
	})
	registerBuiltin("type_is_int", func(args ...Object) Object {
		return BoolObject(len(args) > 0 && args[0].Type() == ObjInt)
	})
	registerBuiltin("type_is_float", func(args ...Object) Object {
		return BoolObject(len(args) > 0 && args[0].Type() == ObjFloat)
	})
	registerBuiltin("type_is_string", func(args ...Object) Object {
		return BoolObject(len(args) > 0 && args[0].Type() == ObjString)
	})
	registerBuiltin("type_is_list", func(args ...Object) Object {
		return BoolObject(len(args) > 0 && args[0].Type() == ObjList)
	})
	registerBuiltin("type_is_map", func(args ...Object) Object {
		return BoolObject(len(args) > 0 && args[0].Type() == ObjMap)
	})
	registerBuiltin("type_is_fn", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		t := args[0].Type()
		return BoolObject(t == ObjFn || t == ObjBuiltin)
	})
}

func registerConvBuiltins() {
	registerBuiltin("conv_to_int", func(args ...Object) Object {
		if len(args) < 1 {
			return IntObject(0)
		}
		if i, ok := toInt(args[0]); ok {
			return IntObject(i)
		}
		return IntObject(0)
	})
	registerBuiltin("conv_to_float", func(args ...Object) Object {
		if len(args) < 1 {
			return FloatObject(0)
		}
		if f, ok := toFloat(args[0]); ok {
			return FloatObject(f)
		}
		return FloatObject(0)
	})
	registerBuiltin("conv_to_string", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		return StringObject(args[0].Inspect())
	})
	registerBuiltin("conv_to_bool", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		return BoolObject(IsTruthy(args[0]))
	})
	registerBuiltin("conv_to_list", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{}
		}
		if list, ok := args[0].(ListObject); ok {
			return list
		}
		return ListObject{Elements: []Object{args[0]}}
	})
	registerBuiltin("conv_hex_encode", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		return StringObject(hexEncode(args[0].Inspect()))
	})
	registerBuiltin("conv_hex_decode", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		decoded, err := hexDecode(args[0].Inspect())
		if err != nil {
			return StringObject("")
		}
		return StringObject(decoded)
	})
	registerBuiltin("conv_b64_encode", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		return StringObject(base64.StdEncoding.EncodeToString([]byte(args[0].Inspect())))
	})
	registerBuiltin("conv_b64_decode", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		decoded, err := base64.StdEncoding.DecodeString(args[0].Inspect())
		if err != nil {
			return StringObject("")
		}
		return StringObject(string(decoded))
	})
}

func registerLogBuiltins() {
	registerBuiltin("log_info", func(args ...Object) Object {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Inspect()
		}
		fmt.Printf("[INFO] %s\n", strings.Join(parts, " "))
		return NilObject{}
	})
	registerBuiltin("log_warn", func(args ...Object) Object {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Inspect()
		}
		fmt.Printf("[WARN] %s\n", strings.Join(parts, " "))
		return NilObject{}
	})
	registerBuiltin("log_error", func(args ...Object) Object {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Inspect()
		}
		fmt.Printf("[ERROR] %s\n", strings.Join(parts, " "))
		return NilObject{}
	})
	registerBuiltin("log_debug", func(args ...Object) Object {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Inspect()
		}
		fmt.Printf("[DEBUG] %s\n", strings.Join(parts, " "))
		return NilObject{}
	})
}

func registerTimeBuiltins() {
	registerBuiltin("time_now", func(args ...Object) Object {
		return StringObject(time.Now().Format(time.RFC3339))
	})
	registerBuiltin("time_now_unix", func(args ...Object) Object {
		return IntObject(time.Now().Unix())
	})
	registerBuiltin("time_format", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		layout := time.RFC3339
		if len(args) >= 2 {
			layout = args[1].Inspect()
		}
		t, err := time.Parse(time.RFC3339, args[0].Inspect())
		if err != nil {
			return StringObject("")
		}
		return StringObject(t.Format(layout))
	})
	registerBuiltin("time_parse", func(args ...Object) Object {
		if len(args) < 2 {
			return NilObject{}
		}
		layout := args[0].Inspect()
		value := args[1].Inspect()
		t, err := time.Parse(layout, value)
		if err != nil {
			return NilObject{}
		}
		return StringObject(t.Format(time.RFC3339))
	})
	registerBuiltin("time_sleep", func(args ...Object) Object {
		if len(args) < 1 {
			return NilObject{}
		}
		ms, ok := toInt(args[0])
		if !ok {
			return NilObject{}
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return NilObject{}
	})
	registerBuiltin("time_duration", func(args ...Object) Object {
		if len(args) < 2 {
			return IntObject(0)
		}
		t1, err1 := time.Parse(time.RFC3339, args[0].Inspect())
		t2, err2 := time.Parse(time.RFC3339, args[1].Inspect())
		if err1 != nil || err2 != nil {
			return IntObject(0)
		}
		return IntObject(int64(t2.Sub(t1).Milliseconds()))
	})
}

func registerErrorHandlingBuiltins() {
	registerBuiltin("try", func(args ...Object) Object {
		if len(args) < 1 {
			return ListObject{Elements: []Object{BoolObject(false), ErrorObject{Message: "try requires a function argument"}}}
		}
		fn, ok := args[0].(BuiltinFn)
		if !ok {
			if _, ok2 := args[0].(*FnObject); ok2 {
				return ListObject{Elements: []Object{BoolObject(false), ErrorObject{Message: "try with script fn not yet supported, use try with builtin fn"}}}
			}
			return ListObject{Elements: []Object{BoolObject(false), ErrorObject{Message: "argument is not callable"}}}
		}
		callArgs := []Object{}
		if len(args) > 1 {
			callArgs = args[1:]
		}
		defer func() {
			recover()
		}()
		result := fn.Fn(callArgs...)
		if err, ok := result.(ErrorObject); ok {
			return ListObject{Elements: []Object{BoolObject(false), err}}
		}
		return ListObject{Elements: []Object{BoolObject(true), result}}
	})
	registerBuiltin("catch", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		if err, ok := args[0].(ErrorObject); ok {
			return StringObject(err.Message)
		}
		return StringObject("")
	})
	registerBuiltin("is_error", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		return BoolObject(args[0].Type() == ObjError)
	})
}

func registerOSBuiltins() {
	registerBuiltin("os_exec", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		cmdStr := args[0].Inspect()
		result, err := execShell(cmdStr)
		if err != nil {
			return StringObject(err.Error())
		}
		return StringObject(result)
	})
	registerBuiltin("os_env", func(args ...Object) Object {
		pairs := make(map[string]Object)
		keys := []string{}
		for _, e := range os.Environ() {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				pairs[parts[0]] = StringObject(parts[1])
				keys = append(keys, parts[0])
			}
		}
		return MapObject{Pairs: pairs, Keys: keys}
	})
	registerBuiltin("os_getenv", func(args ...Object) Object {
		if len(args) < 1 {
			return StringObject("")
		}
		return StringObject(os.Getenv(args[0].Inspect()))
	})
	registerBuiltin("os_setenv", func(args ...Object) Object {
		if len(args) < 2 {
			return BoolObject(false)
		}
		err := os.Setenv(args[0].Inspect(), args[1].Inspect())
		return BoolObject(err == nil)
	})
	registerBuiltin("os_cwd", func(args ...Object) Object {
		dir, err := os.Getwd()
		if err != nil {
			return StringObject("")
		}
		return StringObject(dir)
	})
	registerBuiltin("os_chdir", func(args ...Object) Object {
		if len(args) < 1 {
			return BoolObject(false)
		}
		err := os.Chdir(args[0].Inspect())
		return BoolObject(err == nil)
	})
	registerBuiltin("os_args", func(args ...Object) Object {
		elems := make([]Object, len(os.Args))
		for i, a := range os.Args {
			elems[i] = StringObject(a)
		}
		return ListObject{Elements: elems}
	})
	registerBuiltin("os_exit", func(args ...Object) Object {
		code := 0
		if len(args) >= 1 {
			if i, ok := toInt(args[0]); ok {
				code = int(i)
			}
		}
		os.Exit(code)
		return NilObject{}
	})
	registerBuiltin("os_hostname", func(args ...Object) Object {
		name, _ := os.Hostname()
		return StringObject(name)
	})
	registerBuiltin("os_platform", func(args ...Object) Object {
		return StringObject(runtime.GOOS)
	})
	registerBuiltin("os_arch", func(args ...Object) Object {
		return StringObject(runtime.GOARCH)
	})
}

func hexEncode(s string) string {
	return hex.EncodeToString([]byte(s))
}

func hexDecode(s string) (string, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func execShell(cmdStr string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}
