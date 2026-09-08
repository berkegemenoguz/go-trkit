package trkit

import (
	"strings"
	"unicode"
)

// Turkish differs from the default Unicode case mapping in exactly two places,
// both involving the letter i: uppercase of dotted i is dotted İ, and lowercase
// of I is dotless ı. Every other Turkish letter — ş, ğ, ü, ö, ç and their
// capitals — already maps correctly, so only these two are special-cased and
// the rest is delegated to the unicode package.

// upperRune returns the Turkish uppercase form of r.
func upperRune(r rune) rune {
	if r == 'i' {
		return 'İ'
	}
	return unicode.ToUpper(r)
}

// lowerRune returns the Turkish lowercase form of r.
func lowerRune(r rune) rune {
	if r == 'I' {
		return 'ı'
	}
	return unicode.ToLower(r)
}

// ToUpper returns s with every letter uppercased using Turkish rules.
//
// Dotted i becomes dotted İ, so "izmir" yields "İZMİR" where [strings.ToUpper]
// yields "IZMIR". Dotless ı still becomes I.
func ToUpper(s string) string {
	return strings.Map(upperRune, s)
}

// ToLower returns s with every letter lowercased using Turkish rules.
//
// I becomes dotless ı, so "IĞDIR" yields "ığdır" where [strings.ToLower] yields
// "iğdir". Dotted İ still becomes i.
//
// Because of this, ToLower is the wrong tool for folding lookup keys typed on an
// ASCII keyboard: someone writing "ISTANBUL" means "istanbul", but the Turkish
// rule produces "ıstanbul". Use [ToASCII] to build such keys.
func ToLower(s string) string {
	return strings.Map(lowerRune, s)
}

// Title returns s with the first letter of each word uppercased and the rest of
// each word lowercased, both using Turkish rules — so "AHMET YILMAZ" and
// "ahmet yilmaz" both yield "Ahmet Yılmaz".
//
// A word continues through letters, digits, and apostrophes, so the suffix in
// "istanbul'un" stays lowercase and yields "İstanbul'un". Both the ASCII
// apostrophe and the typographic one are recognized.
//
// Note that lowercasing the remainder flattens acronyms: "TBMM" yields "Tbmm".
func Title(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	atWordStart := true
	for _, r := range s {
		switch {
		case !unicode.IsLetter(r):
			b.WriteRune(r)
		case atWordStart:
			b.WriteRune(upperRune(r))
		default:
			b.WriteRune(lowerRune(r))
		}
		atWordStart = !continuesWord(r)
	}

	return b.String()
}

// continuesWord reports whether r keeps a word going, so that the letter after
// it is not treated as a word start. Apostrophes count because Turkish attaches
// suffixes to proper nouns with them, and digits count so that "3d" does not
// become "3D".
func continuesWord(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '’'
}
