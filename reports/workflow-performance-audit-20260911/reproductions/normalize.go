package main
import("fmt";"strings")
func normalizeWhitespace(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}
func main(){a:="command: |\n  echo first\n  echo second\n";b:="command: |\n  echo first echo second\n";fmt.Printf("different_source=%t\nnormalized_equal=%t\nnormalized=%q\n",a!=b,normalizeWhitespace(a)==normalizeWhitespace(b),normalizeWhitespace(a))}
