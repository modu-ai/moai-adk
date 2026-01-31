package rank

import "fmt"

// RankMedal returns a medal string for the given position
func RankMedal(position int) string {
	switch position {
	case 1:
		return "[1st]"
	case 2:
		return "[2nd]"
	case 3:
		return "[3rd]"
	default:
		return ""
	}
}

// FormatTokens formats a token count with K/M suffixes
func FormatTokens(count int64) string {
	if count >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	}
	if count >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(count)/1_000)
	}
	return fmt.Sprintf("%d", count)
}
