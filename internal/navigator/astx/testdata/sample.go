package sample

import "fmt"

// Greeter is a sample type for AST extraction.
type Greeter struct {
	Name string
}

// Greet is a sample function for AST extraction.
func (g *Greeter) Greet() string {
	return fmt.Sprintf("hi %s", g.Name)
}

// NewGreeter constructs a Greeter.
func NewGreeter(name string) *Greeter {
	return &Greeter{Name: name}
}
