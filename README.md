Derived from [kanaka/mal](https://github.com/kanaka/mal) Go implementation. 
`kanaka/mal` is inspired on Clojure.

Keeping 100% backwards compatibility with `kanaka/mal`. 
There almost 100 implementations on almost 100 languages available on [kanaka/mal](https://github.com/kanaka/mal). 

This variant implementation is focused on reusability in Go projects. 
See [mal main](./cmd/mal) for an example on how to embed it in Go code.

Tested on Go version 1.17.

This implementation uses [chzyer/readline](https://github.com/chzyer/readline) instead of C implented readline or libedit, making this implementation pure Go.

# Changes

- `atom` is multithread
- Tests implemented in Go. Original implementation uses a `runtest.py` in Python to keep all implementations compatible. But it makes the Go development less enjoyable. Tests files are the original ones, there is simply a new `runtest_test.go` that substitutes the original Python script
- Some tests are actually in mal, using the macros commented in _Additions_ section (now only the test library itself). Well, actually not many at this moment, see "Test file specs" below
- Reader regexp's are compiled once
- `core` library moved to `lib/core`
- Using [chzyer/readline](https://github.com/chzyer/readline) instead of C `readline` for the mal REPL
- Multiline REPL
- REPL history stored in `~/.mal_history` (instead of kanaka/mal's `~/.mal-history`)
- `(let* () A B C)` returns `C` as Clojure `let` instead of `A`, and evaluates `A`, `B` and `C`
- `(do)` returns nil as Clojure instead of panicking

To test the implementation use:

```bash
go test ./...
```

`go test` actually validates the `step*.mal` files.

There are some benchmarks as well:

```bash
go test -benchmem -benchtime 5s -bench '^.+$' github.com/jig/mal
```

# Additions

- `(range a b)` returns a vector of integers from `a` to `b-1`
- `(unbase64 string)`, `(unbase64 byteString)`, `(str2binary string)`, `(binary2str byteString)`
- `(sleep ms)` sleeps 1 second
- Support of `¬` as string terminator to simplify JSON strings. Strings that start with `{"` and end with `"}` are printed using `¬`, otherwise strings are printed as usual (with `"`). To escape a `¬` character in a `¬` delimited string you must escape it by doubling it: `¬Hello¬¬World!¬` would be printed as `Hello¬World`. This behaviour allows to not to have to escape `"` nor `\` characters.
- `(jsondecode ¬{"key": "value"}¬)` to decode JSON to MAL data and `(jsonencode ...)` does the opposite. Example: `(jsonencode (jsondecode  ¬{"key":"value","key1": [{"a":"b","c":"d"},2,3]}¬))`. Note that MAL vectors (e.g. `[1 2 3]`) and MAL lists (e.g. `(list 1 2 3)` are both converted to JSON vectors always. Decoding a JSON vector is done on a MAL vector always though
- `(context* (do ...))` provides a Go context. Context contents depend on Go, and might be passed to specific functions context compatible
- Test minimal library to be used with `maltest` interpreter (see [./cmd/maltest/](./cmd/maltest/) folder). See below test specs
- Project compatible with GitHub CodeSpaces. Press `.` on your keyboard and you are ready to deploy a CodeSpace with mal in it.
- `(trace expr)` to trace the `expr` code.

# Test file specs

Execute the testfile with:

```bash
$ mal --test .
```

And a minimal test example `sample_test.mal`:

```lisp
(test.suite "complete tests"
    (assert-true "2 + 2 = 4 is true" (= 4 (+ 2 2)))
    (assert-false "2 + 2 = 5 is false" (= 5 (+ 2 2)))
    (assert-throws "0 / 0 throws an error" (/ 0 0)))
```

Some benchmark of the implementations:

```bash
$ go test -bench ".+" -benchtime 2s
```
