package codexwiring

import "fmt"

// CheckStopInventory compares the Codex Stop-chain inventory (StopChainMembers)
// with the Claude Stop handler scripts of the distributed settings template,
// rendered with the hook opt-in off (plain) and on (optIn), each in
// registration order (SPEC-DUAL-HARNESS-HOOK-PARITY-001 REQ-HPR-001,
// AC-HPR-001). It returns one line per problem; none means they match.
func CheckStopInventory(plain, optIn []string) []string {
	return checkStopInventory(StopChainMembers, plain, optIn)
}

func checkStopInventory(inv []StopMember, plain, optIn []string) []string {
	var problems []string
	byScript := make(map[string]StopMember, len(inv))
	for _, m := range inv {
		if m.ClaudeScript == "" {
			problems = append(problems, fmt.Sprintf("member %d (%s) names no Claude script", m.Number, m.Name))
			continue
		}
		if _, dup := byScript[m.ClaudeScript]; dup {
			problems = append(problems, fmt.Sprintf("%s appears in more than one inventory row", m.ClaudeScript))
		}
		byScript[m.ClaudeScript] = m
		switch m.Class {
		case StopClassAdvisory, StopClassRequiredGate, StopClassGoal, StopClassFailOpenOnMissing:
		default:
			problems = append(problems, fmt.Sprintf("member %d (%s) has no valid class (%q)", m.Number, m.Name, m.Class))
		}
		switch m.Placement {
		case StopPlacementInHook:
		case StopPlacementReceipt:
			if m.ReceiptProducer == "" {
				problems = append(problems, fmt.Sprintf("member %d (%s) is receipt-placed but names no receipt producer", m.Number, m.Name))
			}
		default:
			problems = append(problems, fmt.Sprintf("member %d (%s) has no Codex placement (%q)", m.Number, m.Name, m.Placement))
		}
	}

	count := func(render []string) map[string]int {
		c := make(map[string]int, len(render))
		for _, s := range render {
			c[s]++
		}
		return c
	}
	inPlain, inOptIn := count(plain), count(optIn)
	for _, render := range []struct {
		name    string
		scripts []string
		counts  map[string]int
	}{{"opt-in off", plain, inPlain}, {"opt-in on", optIn, inOptIn}} {
		for script, n := range render.counts {
			if _, ok := byScript[script]; !ok {
				problems = append(problems, fmt.Sprintf("Stop handler %s (render %s) has no inventory row", script, render.name))
			}
			if n > 1 {
				problems = append(problems, fmt.Sprintf("Stop handler %s is registered %d times (render %s)", script, n, render.name))
			}
		}
	}

	for _, m := range inv {
		if m.ClaudeScript == "" {
			continue
		}
		p, o := inPlain[m.ClaudeScript] > 0, inOptIn[m.ClaudeScript] > 0
		switch {
		case !p && !o:
			problems = append(problems, fmt.Sprintf("inventory row %d (%s, %s) is absent from both renders", m.Number, m.Name, m.ClaudeScript))
		case m.Conditional != (o && !p):
			problems = append(problems, fmt.Sprintf("inventory row %d (%s) conditional = %v, but the renders say opt-in-only = %v", m.Number, m.Name, m.Conditional, o && !p))
		}
	}
	// Claude order: the opt-in render carries every member, so the inventory
	// numbering must follow its registration order.
	for i, s := range optIn {
		if m, ok := byScript[s]; ok && m.Number != i+1 {
			problems = append(problems, fmt.Sprintf("%s is Stop handler %d in the template but inventory member %d", s, i+1, m.Number))
		}
	}
	return problems
}
