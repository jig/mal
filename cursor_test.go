package lisp

import (
	"context"
	"testing"

	"github.com/jig/lisp/env"
	"github.com/jig/lisp/lib/core"
	"github.com/jig/lisp/types"
)

func TestCursor(t *testing.T) {
	bootEnv, err := env.NewEnv(nil, nil, nil)
	if err != nil {
		panic(err)
	}
	core.Load(bootEnv)
	core.LoadInput(bootEnv)

	bootEnv.Set(types.Symbol{Val: "eval"}, types.Func{Fn: func(ctx context.Context, a []types.MalType) (types.MalType, error) {
		return EVAL(ctx, a[0], bootEnv)
	}})
	bootEnv.Set(types.Symbol{Val: "*ARGV*"}, types.List{})

	ctx := context.Background()
	// core.mal: defined using the language itself
	_, err = REPL(ctx, bootEnv, `(def *host-language* "go")`, types.NewCursorFile(t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		Module string
		Code   string
		Error  error
	}{
		{
			Module: "multiline-string",
			Code:   multiline,
			Error: types.MalError{
				Cursor: &types.Position{Row: 6},
			},
		}, {
			Module: "codeThrow",
			Code:   codeThrow,
			Error: types.MalError{
				Cursor: &types.Position{Row: 4},
			},
		},
		{
			Module: "codeTryAndThrowAndCatch",
			Code:   codeTryAndThrowAndCatch,
			Error:  nil,
		},
		{
			Module: "codeUndefinedSymbol",
			Code:   codeUndefinedSymbol,
			Error: types.MalError{
				Cursor: &types.Position{Row: 3},
			},
		},
		{
			Module: "codeLetIsBogus",
			Code:   codeLetIsBogus,
			Error: types.MalError{
				Cursor: &types.Position{Row: 4},
			},
		},
		{
			Module: "codeCorrect",
			Code:   codeCorrect,
			Error:  nil,
		},
		{
			Module: "codeMissingRightBracket",
			Code:   codeMissingRightBracket,
			Error: types.MalError{
				Cursor: &types.Position{Row: 8},
			},
		},
		{
			Module: "codeTooManyRightBrackets",
			Code:   codeTooManyRightBrackets,
			Error: types.MalError{
				Cursor: &types.Position{Row: 25},
			},
		},
	} {
		subEnv, err := env.NewEnv(bootEnv, nil, nil)
		if err != nil {
			panic(err)
		}
		ast, err := REPLPosition(ctx, subEnv, "(do\n"+testCase.Code+"\na)", &types.Position{
			Module: &testCase.Module,
			Row:    0,
		})
		switch err := err.(type) {
		case nil:
			if testCase.Error != nil {
				t.Fatalf("Expected error %q", testCase.Error)
			}
			continue
		case types.MalError:
			if err.Cursor.Row != testCase.Error.(types.MalError).Cursor.Row {
				t.Fatal(err.Error(), err.Cursor.Row, testCase.Error.(types.MalError).Cursor.Row)
			}
			continue
		default:
			//			t.Fatal(err)
		}
		if ast == "" {
			t.Error(testCase.Module, "(no error) AST is nil")
			continue
		}
		if ast != "1234" {
			t.Error(testCase.Module, "(no error) REPL didn't reach the end")
			continue
		}
	}
}

var multiline = `;; multiline strings

(def multi ¬line1
	line6¬)

(throw "pum")`

var codeCorrect = `;; prerequisites
;; Trivial but convenient functions.

;; Integer predecessor (number -> number)
(def inc (fn [a] (+ a 1)))

;; Integer predecessor (number -> number)
(def dec (fn (a) (- a 1)))

;; Integer nullity test (number -> boolean)
(def zero? (fn (n) (= 0 n)))

;; Returns the unchanged argument.
(def identity (fn (x) x))

;; Generate a hopefully unique symbol. See section "Plugging the Leaks"
;; of http://www.gigamonkeys.com/book/macros-defining-your-own.html
(def gensym
  (let [counter (atom 0)]
    (fn []
      (symbol (str "G__" (swap! counter inc))))))

(def a 1234)
`

var codeMissingRightBracket = `;; prerequisites
;; Trivial but convenient functions.

;; Integer predecessor (number -> number)
(def inc (fn [a] (+ a 1)))

;; Integer predecessor (number -> number) ;; MISSING ) ON NEXT LINE:
(def dec (fn (a) (- a 1))

;; Integer nullity test (number -> boolean)
(def zero? (fn (n) (= 0 n)))

;; Returns the unchanged argument.
(def identity (fn (x) x))

;; Generate a hopefully unique symbol. See section "Plugging the Leaks"
;; of http://www.gigamonkeys.com/book/macros-defining-your-own.html
(def gensym
  (let [counter (atom 0)]
    (fn []
      (symbol (str "G__" (swap! counter inc))))))

(def a 1234)
`

var codeTooManyRightBrackets = `;; prerequisites
;; Trivial but convenient functions.

;; Integer predecessor (number -> number)
(def inc (fn [a] (+ a 1)))

;; Integer predecessor (number -> number)
(def dec (fn (a) (- a 1))))

;; Integer nullity test (number -> boolean)
(def zero? (fn (n) (= 0 n)))

;; Returns the unchanged argument.
(def identity (fn (x) x))

;; Generate a hopefully unique symbol. See section "Plugging the Leaks"
;; of http://www.gigamonkeys.com/book/macros-defining-your-own.html
(def gensym
  (let [counter (atom 0)]
    (fn []
      (symbol (str "G__" (swap! counter inc))))))

(def a 1234)
`
var codeThrow = `;; this will throw an error
;; in a trivial way

(throw "boo")
`

var codeTryAndThrowAndCatch = `;; throwing an error and catching
;; must not involve program lines

(try
	abc
	(catch exc
		(str "exc is:" exc)))

(def a 1234)
`

// var codeTryAndThrow = `;; throwing an error and catching
// ;; must not involve program lines

// (try
// 	abc
// 	(catch exc
// 		(str "exc is:" exc)))

// (def a 1234)
// `

var codeLetIsBogus = `;; let requires a vector with even elements

(let [x 1
	y]
	y)
`

var codeUndefinedSymbol = `;; undefined-symbol is undefined

undefined-symbol
`
