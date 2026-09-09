package trkit

import "strings"

// A Turkish national number is ten digits: a three-digit prefix — an operator
// prefix for mobiles, an area code for landlines — followed by seven more.
const nationalNumberLength = 10

// IsValidMobilePhone reports whether phone is a well-formed Turkish mobile
// number.
//
// It accepts the ways numbers are actually written — "0532 123 45 67",
// "+90 532 123 45 67", "(0532) 123-45-67", "905321234567", and the bare
// "5321234567" — and requires ten national digits beginning with 5.
func IsValidMobilePhone(phone string) bool {
	n, ok := nationalNumber(phone)
	return ok && isMobileNational(n)
}

// IsValidLandlinePhone reports whether phone is a well-formed Turkish landline
// number.
//
// It accepts the same written forms as [IsValidMobilePhone] and requires ten
// national digits whose first three are a real area code. Unlike the mobile
// check, this one is strict: area codes belong to provinces and have not
// changed since 1999, so an unknown code means a wrong number rather than a
// stale table.
func IsValidLandlinePhone(phone string) bool {
	n, ok := nationalNumber(phone)
	return ok && isLandlineNational(n)
}

// IsValidPhone reports whether phone is a well-formed Turkish number of either
// kind, mobile or landline.
//
// The two categories cannot overlap: area codes begin with 2, 3, or 4, and
// mobile prefixes begin with 5.
func IsValidPhone(phone string) bool {
	n, ok := nationalNumber(phone)
	return ok && (isMobileNational(n) || isLandlineNational(n))
}

// isMobileNational reports whether a ten-digit national number is a mobile one.
//
// The test is deliberately loose: any prefix starting with 5, rather than a
// list of the prefixes operators hold today. Those change as new operators are
// licensed, and a hard-coded list would start rejecting valid numbers the
// moment it went stale — a failure that only shows up as real users being
// turned away. Landline area codes get the opposite treatment, because they are
// tied to provinces and effectively frozen.
func isMobileNational(n string) bool {
	return n[0] == '5'
}

// isLandlineNational reports whether a ten-digit national number begins with a
// known area code.
//
// The code is read digit by digit rather than through strconv: the three bytes
// are already known to be digits, so parsing them cannot fail, and doing it by
// hand avoids an error branch that no input could ever reach.
func isLandlineNational(n string) bool {
	code := int(n[0]-'0')*100 + int(n[1]-'0')*10 + int(n[2]-'0')
	_, known := areaCodeToCity[code]
	return known
}

// nationalNumber strips formatting and any country prefix from phone and
// returns the ten-digit national number.
//
// Spaces, hyphens, parentheses, dots, and slashes are treated as formatting; a
// leading "+" is allowed. Any other character rejects the input outright rather
// than being skipped, so "0532-ABC-4567" does not quietly pass.
func nationalNumber(phone string) (string, bool) {
	var digits strings.Builder
	digits.Grow(len(phone))

	for i, r := range strings.TrimSpace(phone) {
		switch {
		case r >= '0' && r <= '9':
			digits.WriteRune(r)
		case r == '+' && i == 0:
			// The international prefix carries no digits of its own.
		case isPhoneSeparator(r):
			// Formatting, ignored.
		default:
			return "", false
		}
	}

	s := digits.String()

	// Peel off whichever country prefix is present. Nothing here is ambiguous:
	// no national number begins with 0 or 9, so a leading 0 is always the trunk
	// prefix and a leading 90 on an over-long string is always the country code.
	switch {
	case strings.HasPrefix(s, "0090"):
		s = s[4:]
	case strings.HasPrefix(s, "90") && len(s) > nationalNumberLength:
		s = s[2:]
	case strings.HasPrefix(s, "0"):
		s = s[1:]
	}

	if len(s) != nationalNumberLength {
		return "", false
	}
	return s, true
}

// isPhoneSeparator reports whether r is punctuation people use to group the
// digits of a phone number.
func isPhoneSeparator(r rune) bool {
	switch r {
	case ' ', '-', '(', ')', '.', '/':
		return true
	}
	return false
}
