package trkit

import "strings"

// A Turkish IBAN is 26 characters once separators are removed: the country code
// "TR", two check digits, a five-digit bank code, one reserved digit that is
// always zero, and a sixteen-character account number.
const (
	ibanLength      = 26
	ibanReservedPos = 9

	// ibanGroupSize is how many characters go in each group of the display
	// form; the last group holds the remainder.
	ibanGroupSize = 4
)

// IsValidIBAN reports whether iban is a valid Turkish IBAN.
//
// Spaces are ignored and lowercase letters are accepted, so both the grouped
// form people write by hand and the compact form stored in databases are
// recognized. Any other character is rejected.
//
// Validation covers the whole number: length, the "TR" country code, the
// reserved zero, and the ISO 7064 mod-97 checksum. The checksum alone would not
// be enough — strings of the wrong length can satisfy mod-97, so the structural
// check is not redundant.
//
// Only Turkish IBANs are accepted. A valid IBAN from another country is
// reported as invalid here; see the package documentation for why.
func IsValidIBAN(iban string) bool {
	_, err := NormalizeIBAN(iban)
	return err == nil
}

// NormalizeIBAN returns iban in the canonical form to store and compare:
// spaces removed and letters uppercased, so that the many ways people write one
// number collapse to a single string.
//
// It validates before returning, applying exactly the checks [IsValidIBAN]
// applies — anything it returns is a valid Turkish IBAN. Invalid input yields
// an empty string and an error matching [ErrInvalidIBAN].
//
// The function is idempotent: normalizing an already-normalized IBAN returns it
// unchanged.
func NormalizeIBAN(iban string) (string, error) {
	s := cleanIBAN(iban)
	if !hasIBANStructure(s) || !ibanChecksumOK(s) {
		return "", ErrInvalidIBAN
	}
	return s, nil
}

// FormatIBAN returns iban grouped in fours for display, the form printed on
// statements and shown in interfaces:
//
//	TR33 0006 1005 1978 6457 8413 26
//
// The trailing group is shorter, since 26 does not divide evenly by four.
//
// Input is normalized and validated first, so formatting accepts every form
// [NormalizeIBAN] does and rejects everything it rejects, returning an empty
// string and an error matching [ErrInvalidIBAN].
//
// The result round-trips: passing it back through [NormalizeIBAN] returns the
// compact form again.
func FormatIBAN(iban string) (string, error) {
	s, err := NormalizeIBAN(iban)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.Grow(len(s) + (len(s)-1)/ibanGroupSize)

	for i := 0; i < len(s); i += ibanGroupSize {
		if i > 0 {
			b.WriteByte(' ')
		}

		end := i + ibanGroupSize
		if end > len(s) {
			end = len(s)
		}
		b.WriteString(s[i:end])
	}

	return b.String(), nil
}

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

// ibanChecksumOK reports whether s satisfies the ISO 7064 mod-97 check that
// every IBAN carries, treating s as already cleaned by [cleanIBAN].
//
// The check is defined over the number rotated four places, so that the country
// code and check digits move to the end, with each letter expanded to a
// two-digit value (A is 10, Z is 35). The remainder must come out as 1.
//
// The remainder is accumulated digit by digit rather than by building one huge
// integer, which keeps the arithmetic inside a machine word.
func ibanChecksumOK(s string) bool {
	if len(s) == 0 {
		return false
	}

	remainder := 0
	for i := 0; i < len(s); i++ {
		c := s[(i+4)%len(s)]

		switch {
		case c >= '0' && c <= '9':
			remainder = (remainder*10 + int(c-'0')) % 97
		case c >= 'A' && c <= 'Z':
			remainder = (remainder*100 + int(c-'A') + 10) % 97
		default:
			return false
		}
	}

	return remainder == 1
}
