package trkit

import "strings"

// yknPrefix begins every YKN. The block is set aside for foreign residents, so
// no TCKN is issued from it.
const yknPrefix = "99"

// IsValidTCKN reports whether tckn is a well-formed Turkish national identity
// number (T.C. Kimlik Numarası).
//
// The number must be exactly 11 ASCII digits and must not begin with zero. Its
// last two digits are checksums over the first nine: the tenth derives from the
// weighted difference between the odd- and even-positioned digits, the eleventh
// from the sum of the first ten.
//
// Separators and surrounding whitespace are not accepted. A number that passes
// this check is well-formed, not necessarily issued — trkit never contacts a
// registry.
func IsValidTCKN(tckn string) bool {
	return hasIdentityChecksum(tckn)
}

// IsValidYKN reports whether ykn is a well-formed foreign identity number
// (Yabancı Kimlik Numarası), the number Türkiye issues to foreign residents.
//
// A YKN has the same shape and the same two checksum digits as a TCKN. What
// sets it apart is that it begins with 99, a block reserved for foreigners.
//
// As with [IsValidTCKN], passing means well-formed, not issued.
func IsValidYKN(ykn string) bool {
	return hasIdentityChecksum(ykn) && strings.HasPrefix(ykn, yknPrefix)
}

// hasIdentityChecksum reports whether s is eleven ASCII digits, not starting
// with zero, whose last two digits satisfy the checksum that TCKN and YKN
// share.
func hasIdentityChecksum(s string) bool {
	d, ok := parseDigits(s, 11)
	if !ok || d[0] == 0 {
		return false
	}

	var odd, even int
	for i := 0; i < 9; i += 2 {
		odd += d[i] // digits 1, 3, 5, 7, 9
	}
	for i := 1; i < 8; i += 2 {
		even += d[i] // digits 2, 4, 6, 8
	}

	// odd*7-even is negative whenever the even-positioned digits dominate, and
	// Go's % keeps the sign of the dividend, so fold the result back into 0..9
	// before comparing. Skipping this rejects otherwise valid numbers.
	if ((odd*7-even)%10+10)%10 != d[9] {
		return false
	}

	sum := 0
	for _, v := range d[:10] {
		sum += v
	}
	return sum%10 == d[10]
}

// IsValidVKN reports whether vkn is a well-formed Turkish tax identification
// number (Vergi Kimlik Numarası).
//
// The number must be exactly 10 ASCII digits, the last being a check digit over
// the first nine. Each of those is offset by its distance from the check digit,
// weighted by a descending power of two, and reduced modulo 9; a term that
// reduces to zero counts as nine unless the offset digit was itself zero.
//
// Unlike a TCKN, a VKN may begin with zero.
//
// As with [IsValidTCKN], passing means well-formed, not registered.
func IsValidVKN(vkn string) bool {
	d, ok := parseDigits(vkn, 10)
	if !ok {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		tmp := (d[i] + 9 - i) % 10
		if tmp == 0 {
			continue // a zero offset digit contributes nothing
		}
		term := (tmp << (9 - i)) % 9 // tmp * 2^(9-i), reduced
		if term == 0 {
			term = 9
		}
		sum += term
	}

	return (10-sum%10)%10 == d[9]
}

// parseDigits reads s as exactly n ASCII decimal digits and returns their
// numeric values. It reports false if the length differs or any byte is not a
// digit, which also rejects separators, whitespace, and non-ASCII input —
// comparing byte length is safe here because every multi-byte UTF-8 sequence
// contains bytes outside the '0'-'9' range.
func parseDigits(s string, n int) ([]int, bool) {
	if len(s) != n {
		return nil, false
	}
	d := make([]int, n)
	for i := 0; i < n; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return nil, false
		}
		d[i] = int(c - '0')
	}
	return d, true
}
