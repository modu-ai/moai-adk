package gar

// EdgeCases tests boundary behavior of the refined rule.
//   - EdgeReturnNilInErrBlock: `return nil` inside `if err != nil` — candidate B
//     would match this (B has no name guard). Candidate C (name guard) excludes it.
//   - EdgeMultiValue: `return a, err` — multi-value return; single-metavar `return $ERR`
//     should not match (ast-grep single metavar = single expression node).
//   - EdgeBareErrReturn: bare `return err` at top level (no if guard) — only
//     candidate A (name guard, no structural guard) would match this.
func EdgeReturnNilInErrBlock(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func EdgeMultiValue() (int, error) {
	if err := bar(); err != nil {
		return 0, err
	}
	return 0, nil
}

func EdgeBareErrReturn() error {
	err := bar()
	return err
}

func bar() error { return nil }
