package compiler

type Compiler interface {
	CompileFromString(string) (map[string]*Contract, error)
}
