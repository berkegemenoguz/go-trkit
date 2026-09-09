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
// A word that is already written entirely in capitals and contains no vowel is
// left alone, because Turkish words always carry a vowel and such a word is
// therefore an initialism: "TBMM" and "KDV dahildir" survive as written. An
// acronym that does contain a vowel — TÜBİTAK, ABD — cannot be told apart from
// a shouted word by its shape alone, so it is title-cased like any other unless
// named through [TitleWith].
//
// Nothing is ever promoted to capitals. A word typed in lower case stays that
// way, so "aş" does not become "AŞ" on account of an abbreviation that happens
// to be spelled the same.
//
// Title formats; it does not decide what a word means. It is the one function
// in this package whose correct output depends on that, so treat its result as
// a sensible default rather than a guarantee. It also does not apply the
// Turkish Language Institute's rules for titles, which keep conjunctions like
// "ve" in lower case — Title capitalizes every word, which is what normalizing
// a name or a place suits.
func Title(s string) string {
	return titleWith(s, nil)
}

func titleWith(s string, acronyms map[string]bool) string {
	var b strings.Builder
	b.Grow(len(s))

	var word []rune
	flush := func() {
		if len(word) > 0 {
			b.WriteString(titleWord(string(word), acronyms))
			word = word[:0]
		}
	}

	for _, r := range s {
		if isWordRune(r) {
			word = append(word, r)
			continue
		}
		flush()
		b.WriteRune(r)
	}
	flush()

	return b.String()
}

// titleWord case-converts a single word, leaving it untouched if it reads as an
// acronym.
func titleWord(w string, acronyms map[string]bool) string {
	if isAcronym(w, acronyms) {
		return w
	}

	var b strings.Builder
	b.Grow(len(w))

	for i, r := range w {
		switch {
		case !unicode.IsLetter(r):
			b.WriteRune(r)
		case i == 0:
			b.WriteRune(upperRune(r))
		default:
			b.WriteRune(lowerRune(r))
		}
	}

	return b.String()
}

// isAcronym reports whether w should keep the capitals it was written with.
//
// The first condition does the heavy lifting: the word must already be in
// capitals. That is what keeps the rule from ever promoting a lower-case word,
// and it is why a short abbreviation colliding with an ordinary word — AS, AŞ —
// costs nothing to anyone who wrote the ordinary word in lower case.
func isAcronym(w string, acronyms map[string]bool) bool {
	if !isAllUpper(w) {
		return false
	}
	return !hasVowel(w) || acronyms[w]
}

// isAllUpper reports whether every letter in w is already uppercase under
// Turkish rules, and that w has at least one letter.
//
// The comparison goes through [upperRune] rather than the unicode package:
// "istanbul" uppercases to "İSTANBUL", so testing it against unicode's
// "ISTANBUL" would wrongly report a lower-case word as already capitalized.
func isAllUpper(w string) bool {
	seenLetter := false

	for _, r := range w {
		if !unicode.IsLetter(r) {
			continue
		}
		seenLetter = true
		if r != upperRune(r) {
			return false
		}
	}

	return seenLetter
}

// hasVowel reports whether w contains a Turkish vowel, including the
// circumflexed forms that appear in words like kâğıt and mahkûm.
func hasVowel(w string) bool {
	for _, r := range w {
		switch r {
		case 'a', 'e', 'ı', 'i', 'o', 'ö', 'u', 'ü', 'â', 'î', 'û',
			'A', 'E', 'I', 'İ', 'O', 'Ö', 'U', 'Ü', 'Â', 'Î', 'Û':
			return true
		}
	}
	return false
}

// isWordRune reports whether r belongs to a word rather than separating two of
// them. Apostrophes count because Turkish attaches suffixes to proper nouns
// with them, and digits count so that "3d" does not become "3D".
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '’'
}

// turkishToASCII maps each Turkish letter that has no ASCII form to its closest
// ASCII equivalent, including the circumflexed vowels that appear in words like
// Hakkâri, kâğıt, and mahkûm.
//
// The two entries that matter most are ı to i and İ to I: they are what let a
// name typed on an ASCII keyboard fold to the same key as its properly spelled
// form. See [PlateFromCity].
var turkishToASCII = map[rune]rune{
	'ı': 'i', 'İ': 'I',
	'ş': 's', 'Ş': 'S',
	'ğ': 'g', 'Ğ': 'G',
	'ü': 'u', 'Ü': 'U',
	'ö': 'o', 'Ö': 'O',
	'ç': 'c', 'Ç': 'C',
	'â': 'a', 'Â': 'A',
	'î': 'i', 'Î': 'I',
	'û': 'u', 'Û': 'U',
}

// ToASCII returns s with Turkish letters replaced by their plain ASCII
// equivalents, so that "Şanlıurfa" becomes "Sanliurfa" and "Hakkâri" becomes
// "Hakkari".
//
// Letter case is preserved: each Turkish letter maps to the ASCII letter of the
// same case. Anything the table does not cover — including accented letters
// from other languages — passes through unchanged, so the result is not
// guaranteed to be pure ASCII for arbitrary input. Callers that need that
// guarantee should filter afterwards, as [Slugify] does.
func ToASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if ascii, ok := turkishToASCII[r]; ok {
			return ascii
		}
		return r
	}, s)
}

// Slugify returns s as a URL slug: transliterated to ASCII, lowercased, with
// runs of anything else collapsed into single hyphens and no hyphen at either
// end. "Şanlıurfa Merkez" becomes "sanliurfa-merkez".
//
// Only a-z and 0-9 survive; every other character acts as a separator. Letters
// that [ToASCII] does not cover therefore drop out rather than appearing in the
// slug, which is what keeps the result safe to put in a URL.
func Slugify(s string) string {
	return SlugifyWith(s, "-")
}

// SlugifyWith is [Slugify] with a separator of the caller's choosing, for the
// underscores some systems expect. An empty separator runs the words together.
//
// The separator is inserted verbatim and is not itself sanitized.
func SlugifyWith(s, sep string) string {
	// strings.ToLower, deliberately, not this package's ToLower: after ToASCII
	// the text is ASCII, and the Turkish rule would turn I back into ı — a
	// non-ASCII rune that the filter below would then discard.
	ascii := strings.ToLower(ToASCII(s))

	var b strings.Builder
	b.Grow(len(ascii))

	separatorPending, wroteAny := false, false
	for i := 0; i < len(ascii); i++ {
		c := ascii[i]

		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			if separatorPending && wroteAny {
				b.WriteString(sep)
			}
			b.WriteByte(c)
			separatorPending, wroteAny = false, true
			continue
		}

		// Anything else only marks a boundary. Writing the separator lazily —
		// when the next kept character arrives — is what collapses runs and
		// trims the ends in one pass.
		separatorPending = true
	}

	return b.String()
}
