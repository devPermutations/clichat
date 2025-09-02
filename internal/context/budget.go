package ctxutil

import (
	"fmt"
)

// PercentUsed returns a human string for context usage percentage.
func PercentUsed(totalUsed, contextTokens int) string {
	if contextTokens <= 0 || totalUsed < 0 {
		return "N/A"
	}
	pct := float64(totalUsed) / float64(contextTokens) * 100.0
	return fmt.Sprintf("%.1f%%", pct)
}

// EstimateTokens approximates token count from character length.
// Rough heuristic: ~4 chars per token.
func EstimateTokens(text string) int {
	l := len(text)
	if l <= 0 {
		return 0
	}
	// ceil(l/4)
	return (l + 3) / 4
}

// EstimateTokensForContents sums EstimateTokens over a slice of strings.
func EstimateTokensForContents(contents []string) int {
	total := 0
	for _, c := range contents {
		total += EstimateTokens(c)
	}
	return total
}
