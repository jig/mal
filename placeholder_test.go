package lisp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jig/lisp/env"
	"github.com/jig/lisp/lib/core"
	"github.com/jig/lisp/reader"

	. "github.com/jig/lisp/lnotation"
	. "github.com/jig/lisp/types"
)

type Example struct {
	A int
	B string
}

func TestPlaceholders(t *testing.T) {
	repl_env, _ := env.NewEnv(nil, nil, nil)
	for k, v := range core.NS {
		repl_env.Set(
			Symbol{Val: k},
			Func{Fn: v.(func([]MalType, *context.Context) (MalType, error))},
		)
	}

	str := `(do
				(def! v0 $0)
				(def! v1 $1)
				(def! vNUMBER $NUMBER)
				(def! v3 $3)
				(def! v4 $4)
				true)`

	exp, err := reader.Read_str(
		str,
		nil,
		&HashMap{
			Val: map[string]MalType{
				"$0":      "hello",
				"$1":      "{\"key\": \"value\"}",
				"$NUMBER": 44,
				"$3":      LS("+", 1, 1),
				"$4": LS("jsonencode",
					Example{A: 3, B: "blurp"}),
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(PRINT(exp))

	res, err := EVAL(exp, repl_env, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.(bool) {
		v0, err := repl_env.Get(Symbol{Val: "v0"})
		if err != nil {
			t.Fatal(err)
		}
		if v0.(string) != "hello" {
			t.Fatal("no hello")
		}

		v1, err := repl_env.Get(Symbol{Val: "v1"})
		if err != nil {
			t.Fatal(err)
		}
		if v1.(string) != "{\"key\": \"value\"}" {
			t.Fatal("no {\"key\": \"value\"}")
		}

		v2, err := repl_env.Get(Symbol{Val: "vNUMBER"})
		if err != nil {
			t.Fatal(err)
		}
		if v2.(int) != 44 {
			t.Fatal("no 44")
		}

		v3, err := repl_env.Get(Symbol{Val: "v3"})
		if err != nil {
			t.Fatal(err)
		}
		if v3.(int) != 2 {
			t.Fatal("no 2")
		}

		v4, err := repl_env.Get(Symbol{Val: "v4"})
		if err != nil {
			t.Fatal(err)
		}
		switch v4 := v4.(type) {
		case string:
			if v4 != "{\"A\":3,\"B\":\"blurp\"}" {
				t.Fatal("invalid value")
			}
		default:
			t.Fatal("invalid type")
		}
	}
}

func TestREADWithPreamble(t *testing.T) {
	repl_env, _ := env.NewEnv(nil, nil, nil)
	for k, v := range core.NS {
		repl_env.Set(
			Symbol{Val: k},
			Func{Fn: v.(func([]MalType, *context.Context) (MalType, error))},
		)
	}

	str :=
		`;; $0 "hello"
;; $1 {"key" "value"}
;; $NUMBER 44
;; $4 (+ 1 1)

(do
	(def! v0 $0)
	(def! v1 $1)
	(def! v2 $NUMBER)
	(def! v3 $3) ;; this is nil
	(def! v4 '$4)
	true)
`

	// exp, err := READ_WithPlaceholders(str, nil, []MalType{"hello", "{\"key\": \"value\"}", 44, List{Val: []MalType{Symbol{Val: "quote"}, List{Val: []MalType{23, 37}}}}})
	exp, err := READWithPreamble(str, nil)
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(PRINT(exp))

	res, err := EVAL(exp, repl_env, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.(bool) {
		v0, err := repl_env.Get(Symbol{Val: "v0"})
		if err != nil {
			t.Fatal(err)
		}
		if v0.(string) != "hello" {
			t.Fatal("no hello")
		}

		v1, err := repl_env.Get(Symbol{Val: "v1"})
		if err != nil {
			t.Fatal(err)
		}
		h, ok := v1.(HashMap)
		if !ok {
			t.Fatal("no {\"key\": \"value\"}")
		}
		if len(h.Val) != 1 {
			t.Fatal("pum")
		}
		if h.Val["key"].(string) != "value" {
			t.Fatal("pum2")
		}

		v2, err := repl_env.Get(Symbol{Val: "v2"})
		if err != nil {
			t.Fatal(err)
		}
		if v2.(int) != 44 {
			t.Fatal("no 44")
		}

		v3, err := repl_env.Get(Symbol{Val: "v3"})
		if err != nil {
			t.Fatal(err)
		}
		if v3 != nil {
			t.Fatal("no 2")
		}

		v4, err := repl_env.Get(Symbol{Val: "v4"})
		if err != nil {
			t.Fatal(err)
		}
		l, ok := v4.(List)
		if !ok {
			t.Fatal("no (+ 1 1)")
		}
		if len(l.Val) != 3 {
			t.Fatal("pum3")
		}
		if l.Val[0].(Symbol).Val != "+" {
			t.Fatal("pum4")
		}
		if l.Val[1].(int) != 1 {
			t.Fatal("pum5")
		}
		if l.Val[2].(int) != 1 {
			t.Fatal("pum6")
		}
	}
}

func TestAddPreamble(t *testing.T) {
	repl_env, _ := env.NewEnv(nil, nil, nil)
	for k, v := range core.NS {
		repl_env.Set(
			Symbol{Val: k},
			Func{Fn: v.(func([]MalType, *context.Context) (MalType, error))},
		)
	}

	str := `(do
	(def! v0 $EXAMPLESTRING)
	(def! v1 $EXAMPLESTRUCT)
	(def! v2 $EXAMPLEINTEGER)
	(def! v3 $UNDEFINED) ;; this is nil
	(def! v4 '$EXAMPLEAST)
	(def! v5 $EXAMPLEBYTESTRING)
	true)`

	eb, err := json.Marshal(Example{A: 1234, B: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	source, err := AddPreamble(str, map[string]MalType{
		"$EXAMPLESTRING":  "hello",
		"$EXAMPLESTRUCT":  string(eb),
		"$EXAMPLEINTEGER": 44,
		"$EXAMPLEAST":     LS("+", 1, 1),
		// byte array is handled as string
		"$EXAMPLEBYTESTRING": []byte("byte-array"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(source)

	exp, err := READWithPreamble(source, nil)
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(PRINT(exp))

	res, err := EVAL(exp, repl_env, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.(bool) {
		v0, err := repl_env.Get(Symbol{Val: "v0"})
		if err != nil {
			t.Fatal(err)
		}
		if v0.(string) != "hello" {
			t.Fatal("no hello")
		}

		v1, err := repl_env.Get(Symbol{Val: "v1"})
		if err != nil {
			t.Fatal(err)
		}
		v1Str, ok := v1.(string)
		if !ok {
			t.Fatal("no {\"key\": \"value\"}")
		}
		if v1Str != `{"A":1234,"B":"hello"}` {
			t.Fatal(v1Str)
		}

		v2, err := repl_env.Get(Symbol{Val: "v2"})
		if err != nil {
			t.Fatal(err)
		}
		if v2.(int) != 44 {
			t.Fatal("no 44")
		}

		v3, err := repl_env.Get(Symbol{Val: "v3"})
		if err != nil {
			t.Fatal(err)
		}
		if v3 != nil {
			t.Fatal("no 2")
		}

		v4, err := repl_env.Get(Symbol{Val: "v4"})
		if err != nil {
			t.Fatal(err)
		}
		l, ok := v4.(List)
		if !ok {
			t.Fatal("no (+ 1 1)")
		}
		if len(l.Val) != 3 {
			t.Fatal("pum3")
		}
		if l.Val[0].(Symbol).Val != "+" {
			t.Fatal("pum4")
		}
		if l.Val[1].(int) != 1 {
			t.Fatal("pum5")
		}
		if l.Val[2].(int) != 1 {
			t.Fatal("pum6")
		}
	}
}

func TestPlaceholdersEmbeddedWrong1(t *testing.T) {
	repl_env, _ := env.NewEnv(nil, nil, nil)
	for k, v := range core.NS {
		repl_env.Set(
			Symbol{Val: k},
			Func{Fn: v.(func([]MalType, *context.Context) (MalType, error))},
		)
	}

	str :=
		`$0 "hello"
;; $1 {"key" "value"}
;; $NUMBER 44
;; $4 (+ 1 1)

(do
	(def! v0 $0)
	(def! v1 $1)
	(def! v2 $NUMBER)
	(def! v3 $3) ;; this is nil
	(def! v4 '$4)
	true)
`

	// exp, err := READ_WithPlaceholders(str, nil, []MalType{"hello", "{\"key\": \"value\"}", 44, List{Val: []MalType{Symbol{Val: "quote"}, List{Val: []MalType{23, 37}}}}})
	_, err := READWithPreamble(str, nil)
	if err == nil {
		t.Fatal("error expected but err was nil")
	}
	if err.Error() != "Error: not all tokens where parsed" {
		t.Fatal(err)
	}
}

func TestPlaceholdersEmbeddedNoBlankLine(t *testing.T) {
	repl_env, _ := env.NewEnv(nil, nil, nil)
	for k, v := range core.NS {
		repl_env.Set(
			Symbol{Val: k},
			Func{Fn: v.(func([]MalType, *context.Context) (MalType, error))},
		)
	}

	// missing blank line must fail
	str :=
		`;; $0 73
;; $1 27
(= (+ $0 $1) 100)
`
	exp, err := READWithPreamble(str, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := EVAL(exp, repl_env, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.(bool) {
		t.Fatal("failed")
	}
}

var notOptimiseBenchFunc string

func BenchmarkAddPreamble(b *testing.B) {
	source := `(do
		(def! v0 $EXAMPLESTRING)
		(def! v2 $EXAMPLEINTEGER)
		(def! v3 $UNDEFINED) ;; this is nil
		(def! v4 '$EXAMPLEAST)
		(def! v5 $EXAMPLEBYTESTRING)
		true)`

	for n := 0; n < b.N; n++ {
		var err error
		notOptimiseBenchFunc, err = AddPreamble(source, map[string]MalType{
			"$EXAMPLESTRING":  "hello",
			"$EXAMPLEINTEGER": 44,
			"$EXAMPLEAST":     LS("+", 1, 1),
			// byte array is handled as string
			"$EXAMPLEBYTESTRING": []byte("byte-array"),
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkREADWithPreamble(b *testing.B) {
	source := `(do
		(def! v0 $EXAMPLESTRING)
		(def! v2 $EXAMPLEINTEGER)
		(def! v3 $UNDEFINED) ;; this is nil
		(def! v4 '$EXAMPLEAST)
		(def! v5 $EXAMPLEBYTESTRING)
		true)`
	codePreamble, err := AddPreamble(source, map[string]MalType{
		"$EXAMPLESTRING":  "hello",
		"$EXAMPLEINTEGER": 44,
		"$EXAMPLEAST":     LS("+", 1, 1),
		// byte array is handled as string
		"$EXAMPLEBYTESTRING": []byte("byte-array"),
	})
	if err != nil {
		b.Fatal(err)
	}
	for n := 0; n < b.N; n++ {
		res, err := READWithPreamble(codePreamble, nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = res
	}
}

func BenchmarkNewEnv(b *testing.B) {
	repl_env, _ := env.NewEnv(nil, nil, nil)
	for k, v := range core.NS {
		repl_env.Set(Symbol{Val: k}, Func{Fn: v.(func([]MalType, *context.Context) (MalType, error))})
	}
	source := `(do
		(def! v0 $EXAMPLESTRING)
		(def! v2 $EXAMPLEINTEGER)
		(def! v3 $UNDEFINED) ;; this is nil
		(def! v4 '$EXAMPLEAST)
		(def! v5 $EXAMPLEBYTESTRING)
		true)`
	codePreamble, err := AddPreamble(source, map[string]MalType{
		"$EXAMPLESTRING":  "hello",
		"$EXAMPLEINTEGER": 44,
		"$EXAMPLEAST":     LS("+", 1, 1),
		// byte array is handled as string
		"$EXAMPLEBYTESTRING": []byte("byte-array"),
	})
	if err != nil {
		b.Fatal(err)
	}

	ast, err := READWithPreamble(codePreamble, nil)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		res, err := EVAL(ast, repl_env, nil)
		if err != nil {
			b.Fatal(err)
		}
		if !res.(bool) {
			b.Fatal(err)
		}
	}
}

func BenchmarkComplete(b *testing.B) {
	repl_env, _ := env.NewEnv(nil, nil, nil)
	for k, v := range core.NS {
		repl_env.Set(Symbol{Val: k}, Func{Fn: v.(func([]MalType, *context.Context) (MalType, error))})
	}

	for n := 0; n < b.N; n++ {
		source := `(do
			(def! v0 $EXAMPLESTRING)
			(def! v2 $EXAMPLEINTEGER)
			(def! v3 $UNDEFINED) ;; this is nil
			(def! v4 '$EXAMPLEAST)
			(def! v5 $EXAMPLEBYTESTRING)

			(def! not (fn* (a) (if a false true)))
			(def! b (not $TESTRESULT))
			(not b))`
		codePreamble, err := AddPreamble(source, map[string]MalType{
			"$TESTRESULT":     true,
			"$EXAMPLESTRING":  "hello",
			"$EXAMPLEINTEGER": 44,
			"$EXAMPLEAST":     LS("+", 1, 1),
			// byte array is handled as string
			"$EXAMPLEBYTESTRING": []byte("byte-array"),
		})
		if err != nil {
			b.Fatal(err)
		}

		ast, err := READWithPreamble(codePreamble, nil)
		if err != nil {
			b.Fatal(err)
		}

		res, err := EVAL(ast, repl_env, nil)
		if err != nil {
			b.Fatal(err)
		}
		if !res.(bool) {
			b.Fatal(err)
		}
	}
}
