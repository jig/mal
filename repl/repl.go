package repl

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	goreadline "github.com/chzyer/readline"
	"github.com/jig/lisp"
	"github.com/jig/lisp/types"
)

// Execute executes the main REPL loop
func Execute(repl_env types.EnvType, ctx *context.Context) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	l, err := goreadline.NewEx(&goreadline.Config{
		Prompt:          "\033[32m»\033[0m ",
		HistoryFile:     dirname + "/.lisp_history",
		AutoComplete:    &lispCompleter{repl_env},
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return err
	}
	defer l.Close()

	log.SetOutput(l.Stderr())
	var lines []string
	for {
		line, err := l.Readline()
		if err == goreadline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			// proper exit with ^D
			return nil
		}

		line = strings.TrimSpace(line)
		lines = append(lines, line)
		completeLine := strings.Join(lines, "\n")
		out, err := lisp.REPL(repl_env, completeLine, ctx)
		if err != nil {
			if err.Error() == "<empty line>" {
				continue
			}
			if err, ok := err.(types.RuntimeError); ok && err.ErrorVal != nil {
				if err.ErrorVal.Error() == "expected ')', got EOF" ||
					err.ErrorVal.Error() == "expected ']', got EOF" ||
					err.ErrorVal.Error() == "expected '}', got EOF" {
					l.SetPrompt("\033[31m›\033[0m ")
					continue
				}
			}
			lines = []string{}
			l.SetPrompt("\033[32m»\033[0m ")
			switch err := err.(type) {
			case types.RuntimeError:
				fmt.Printf("\033[31mError:\033[0m %s\n", err.Error())
				continue
			case types.MalError:
				errorString, err2 := lisp.PRINT(err.Obj)
				if err2 != nil {
					fmt.Printf("\033[31mMalError:\033[0m %s\n", "UNPRINTABLE-ERROR")
					continue
				}
				fmt.Printf("\033[31mMalError:\033[0m %s\n", errorString)
				continue
			default:
				fmt.Printf("Error: %s\n", err)
				continue
			}
		}
		lines = []string{}
		l.SetPrompt("\033[32m»\033[0m ")
		fmt.Printf("%v\n", out)
	}
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case goreadline.CharCtrlZ:
		return r, false
	}
	return r, true
}

type lispCompleter struct {
	repl_env types.EnvType
}

var re = regexp.MustCompile(`[\t\r\n \(\)\[\]\{\}]`)

func (l *lispCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	partial := re.Split(string(line[:pos]), -1)
	lastPartial := partial[len(partial)-1]
	l.repl_env.Map().Range(func(_key, value interface{}) bool {
		key := _key.(string)
		if strings.HasPrefix(key, lastPartial) {
			newLine = append(newLine, []rune(key[len(lastPartial):]))
		}
		return true
	})
	for _, form := range []string{
		"try", "try*",
		"finally", "finally*",
		"catch", "catch*",
		"fn", "fn*",
		"context", "context*",
		"let", "let*",
		"def", "def!",
		"defmacro", "defmacro!",

		"do",
		"macroexpand",
		"if",
		"trace",
		"quote",
		"quasiquote",
		"quasiquoteexpand",
	} {
		if strings.HasPrefix(form, lastPartial) {
			newLine = append(newLine, []rune(form[len(lastPartial):]))
		}
	}
	return newLine, len(lastPartial)
}
