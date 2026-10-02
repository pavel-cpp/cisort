package sorter

import (
	"cmp"
	"strings"

	"github.com/pavel-cpp/cisort/internal/config"
)

// sortKey returns the string compared under order.
func sortKey(spelling string, order config.SortOrder) string {
	if order == config.Lexical {
		return spelling
	}
	return strings.ToLower(spelling)
}

// compareKeys compares keys built by sortKey.
func compareKeys(a, b string, order config.SortOrder) int {
	if order == config.Natural {
		return compareNatural(a, b)
	}
	return strings.Compare(a, b)
}

// compareNatural compares a and b treating digit runs as numbers.
func compareNatural(a, b string) int {
	for a != "" && b != "" {
		if isDigit(a[0]) && isDigit(b[0]) {
			na, restA := cutDigits(a)
			nb, restB := cutDigits(b)
			na, nb = strings.TrimLeft(na, "0"), strings.TrimLeft(nb, "0")
			if c := cmp.Compare(len(na), len(nb)); c != 0 {
				return c
			}
			if c := strings.Compare(na, nb); c != 0 {
				return c
			}
			a, b = restA, restB
			continue
		}
		if c := cmp.Compare(a[0], b[0]); c != 0 {
			return c
		}
		a, b = a[1:], b[1:]
	}
	return cmp.Compare(len(a), len(b))
}

func cutDigits(s string) (digits, rest string) {
	i := 0
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	return s[:i], s[i:]
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
