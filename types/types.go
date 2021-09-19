package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Position struct {
	Module *string
	Row    int
	Col    int
}

type Token struct {
	Value  string
	Cursor Position
}

// Errors/Exceptions
type MalError struct {
	Obj MalType
}

func (e MalError) Error() string {
	return fmt.Sprintf("%#v", e.Obj)
}

// General types
type MalType interface {
}

type EnvType interface {
	Find(key Symbol) EnvType
	Set(key Symbol, value MalType) MalType
	Get(key Symbol) (MalType, error)
	Remove(key Symbol) error
	Map() *sync.Map
	Trace() bool
	SetTrace(bool)
}

// Scalars
func Nil_Q(obj MalType) bool {
	return obj == nil
}

func True_Q(obj MalType) bool {
	b, ok := obj.(bool)
	return ok && b
}

func False_Q(obj MalType) bool {
	b, ok := obj.(bool)
	return ok && !b
}

func Number_Q(obj MalType) bool {
	_, ok := obj.(int)
	return ok
}

// Symbols
type Symbol struct {
	Val string
}

func Symbol_Q(obj MalType) bool {
	_, ok := obj.(Symbol)
	return ok
}

// Keywords
func NewKeyword(s string) (MalType, error) {
	return "\u029e" + s, nil
}

func Keyword_Q(obj MalType) bool {
	s, ok := obj.(string)
	return ok && strings.HasPrefix(s, "\u029e")
}

// Strings
func String_Q(obj MalType) bool {
	_, ok := obj.(string)
	return ok
}

// Functions
type Func struct {
	Fn     func([]MalType, *context.Context) (MalType, error)
	Meta   MalType
	Cursor *Position
}

func Func_Q(obj MalType) bool {
	_, ok := obj.(Func)
	return ok
}

type MalFunc struct {
	Eval    func(MalType, EnvType, *context.Context) (MalType, error)
	Exp     MalType
	Env     EnvType
	Params  MalType
	IsMacro bool
	GenEnv  func(EnvType, MalType, MalType) (EnvType, error)
	Meta    MalType
	Cursor  *Position
}

func MalFunc_Q(obj MalType) bool {
	_, ok := obj.(MalFunc)
	return ok
}

func (f MalFunc) SetMacro() MalType {
	f.IsMacro = true
	return f
}

func (f MalFunc) GetMacro() bool {
	return f.IsMacro
}

// Take either a MalFunc or regular function and apply it to the
// arguments
func Apply(f_mt MalType, a []MalType, ctx *context.Context) (MalType, error) {
	switch f := f_mt.(type) {
	case MalFunc:
		env, e := f.GenEnv(f.Env, f.Params, List{a, nil, f.Cursor})
		if e != nil {
			return nil, e
		}
		return f.Eval(f.Exp, env, ctx)
	case Func:
		return f.Fn(a, ctx)
	case func([]MalType) (MalType, error):
		return f(a)
	default:
		return nil, errors.New("Invalid function to Apply")
	}
}

// Lists
type List struct {
	Val    []MalType
	Meta   MalType
	Cursor *Position
}

func NewList(a ...MalType) MalType {
	return List{Val: a}
}

func List_Q(obj MalType) bool {
	_, ok := obj.(List)
	return ok
}

// Vectors
type Vector struct {
	Val    []MalType
	Meta   MalType
	Cursor *Position
}

func Vector_Q(obj MalType) bool {
	_, ok := obj.(Vector)
	return ok
}

func GetSlice(seq MalType) ([]MalType, error) {
	switch obj := seq.(type) {
	case List:
		return obj.Val, nil
	case Vector:
		return obj.Val, nil
	default:
		return nil, errors.New("GetSlice called on non-sequence")
	}
}

// Hash Maps
type HashMap struct {
	Val    map[string]MalType
	Meta   MalType
	Cursor *Position
}

func NewHashMap(seq MalType) (MalType, error) {
	lst, e := GetSlice(seq)
	if e != nil {
		return nil, e
	}
	if len(lst)%2 == 1 {
		return nil, errors.New("Odd number of arguments to NewHashMap")
	}
	m := map[string]MalType{}
	for i := 0; i < len(lst); i += 2 {
		str, ok := lst[i].(string)
		if !ok {
			return nil, errors.New("expected hash-map key string")
		}
		m[str] = lst[i+1]
	}
	return HashMap{Val: m}, nil
}

func HashMap_Q(obj MalType) bool {
	_, ok := obj.(HashMap)
	return ok
}

// Atoms
type Atom struct {
	Mutex  sync.RWMutex
	Val    MalType
	Meta   MalType
	Cursor *Position
}

func (a *Atom) Set(val MalType) MalType {
	a.Val = val
	return a
}

func Atom_Q(obj MalType) bool {
	_, ok := obj.(*Atom)
	return ok
}

// General functions

func _obj_type(obj MalType) string {
	if obj == nil {
		return "nil"
	}
	return reflect.TypeOf(obj).Name()
}

func Sequential_Q(seq MalType) bool {
	if seq == nil {
		return false
	}
	return (reflect.TypeOf(seq).Name() == "List") ||
		(reflect.TypeOf(seq).Name() == "Vector")
}

func Equal_Q(a MalType, b MalType) bool {
	ota := reflect.TypeOf(a)
	otb := reflect.TypeOf(b)
	if !((ota == otb) || (Sequential_Q(a) && Sequential_Q(b))) {
		return false
	}
	//av := reflect.ValueOf(a); bv := reflect.ValueOf(b)
	//fmt.Printf("here2: %#v\n", reflect.TypeOf(a).Name())
	//switch reflect.TypeOf(a).Name() {
	switch a.(type) {
	case Symbol:
		return a.(Symbol).Val == b.(Symbol).Val
	case List:
		as, _ := GetSlice(a)
		bs, _ := GetSlice(b)
		if len(as) != len(bs) {
			return false
		}
		for i := 0; i < len(as); i += 1 {
			if !Equal_Q(as[i], bs[i]) {
				return false
			}
		}
		return true
	case Vector:
		as, _ := GetSlice(a)
		bs, _ := GetSlice(b)
		if len(as) != len(bs) {
			return false
		}
		for i := 0; i < len(as); i += 1 {
			if !Equal_Q(as[i], bs[i]) {
				return false
			}
		}
		return true
	case HashMap:
		am := a.(HashMap).Val
		bm := b.(HashMap).Val
		if len(am) != len(bm) {
			return false
		}
		for k, v := range am {
			if !Equal_Q(v, bm[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

func (hm HashMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(hm.Val)
}

func (v Vector) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val)
}

func (l List) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.Val)
}

type RuntimeError struct {
	ErrorVal error
	Cursor   *Position
}

func (e RuntimeError) Error() string {
	if e.Cursor == nil {
		return e.ErrorVal.Error()
	}
	if e.Cursor.Row == 0 {
		return e.ErrorVal.Error()
	}
	if e.Cursor.Col == 0 {
		if e.Cursor.Module != nil {
			return fmt.Sprintf("%s(L%d): %s", *e.Cursor.Module, e.Cursor.Row, e.ErrorVal)
		}
		return fmt.Sprintf("(L%d): %s", e.Cursor.Row, e.ErrorVal)
	}
	if e.Cursor.Module != nil {
		return fmt.Sprintf("%s(L%d,%d): %s", *e.Cursor.Module, e.Cursor.Row, e.Cursor.Col, e.ErrorVal)
	}
	return fmt.Sprintf("(L%d,%d): %s", e.Cursor.Row, e.Cursor.Col, e.ErrorVal)
}
