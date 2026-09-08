package trkit

import "strings"

// A Turkish IBAN is 26 characters once separators are removed: the country code
// "TR", two check digits, a five-digit bank code, one reserved digit that is
// always zero, and a sixteen-character account number.
const (
	ibanLength      = 26
	ibanReservedPos = 9
)

// cleanIBAN removes ASCII spaces from s and uppercases its letters, turning the
// grouped form people write by hand into the compact form the checks expect.
//
// Only spaces are removed. Hyphens and other punctuation are left in place so
// that they fail the digit check below rather than being silently accepted —
// grouping IBANs in fours is conventional, hyphenating them is not.
func cleanIBAN(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' {
			continue
		}
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		b.WriteByte(c)
	}

	return b.String()
}

// hasIBANStructure reports whether s, already cleaned by [cleanIBAN], has the
// shape of a Turkish IBAN: the right length, the "TR" country code, digits
// everywhere after it, and a zero in the reserved position.
//
// This is a necessary but not sufficient condition — the checksum is verified
// separately.
func hasIBANStructure(s string) bool {
	if len(s) != ibanLength {
		return false
	}
	if s[0] != 'T' || s[1] != 'R' {
		return false
	}

	for i := 2; i < ibanLength; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return s[ibanReservedPos] == '0'
}
