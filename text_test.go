package trkit

import (
	"fmt"
	"strings"
	"testing"
)

func TestToUpper(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"dotted i becomes dotted capital", "izmir", "İZMİR"},
		{"dotless i becomes plain capital", "ığdır", "IĞDIR"},
		{"both kinds of i in one word", "istanbul", "İSTANBUL"},
		{"other Turkish letters", "çanakkale", "ÇANAKKALE"},
		{"sharp letters", "gümüşhane", "GÜMÜŞHANE"},
		{"already uppercase", "İZMİR", "İZMİR"},
		{"digits and punctuation pass through", "abc-123", "ABC-123"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUpper(tt.input); got != tt.want {
				t.Errorf("ToUpper(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain capital becomes dotless", "IĞDIR", "ığdır"},
		{"dotted capital becomes dotted", "İZMİR", "izmir"},
		{"both kinds of I in one word", "İSTANBUL", "istanbul"},
		{"other Turkish letters", "ÇANAKKALE", "çanakkale"},
		{"sharp letters", "ŞIRNAK", "şırnak"},
		{"already lowercase", "izmir", "izmir"},
		{"digits and punctuation pass through", "ABC-123", "abc-123"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLower(tt.input); got != tt.want {
				t.Errorf("ToLower(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// The whole reason these functions exist is that the standard library gets
// Turkish wrong. This pins the difference so nobody "simplifies" the package by
// delegating to strings.
func TestCaseConversionDiffersFromStdlib(t *testing.T) {
	if got, std := ToUpper("izmir"), strings.ToUpper("izmir"); got == std {
		t.Errorf("ToUpper(%q) = %q, same as strings.ToUpper — the Turkish rule is missing", "izmir", got)
	}
	if got, std := ToLower("IĞDIR"), strings.ToLower("IĞDIR"); got == std {
		t.Errorf("ToLower(%q) = %q, same as strings.ToLower — the Turkish rule is missing", "IĞDIR", got)
	}
}

func TestTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase name", "ahmet yılmaz", "Ahmet Yılmaz"},
		{"shouting name", "AHMET YILMAZ", "Ahmet Yılmaz"},
		{"dotted i at word start", "istanbul", "İstanbul"},
		{"several words", "izmir büyükşehir belediyesi", "İzmir Büyükşehir Belediyesi"},

		// A suffix after an apostrophe continues the word.
		{"ascii apostrophe", "istanbul'un", "İstanbul'un"},
		{"typographic apostrophe", "ankara’nın", "Ankara’nın"},

		// A letter after a digit continues the word too.
		{"digit before letter", "3d yazıcı", "3d Yazıcı"},

		// An all-capital word with no vowel is an initialism, since Turkish
		// words always carry one.
		{"vowelless acronym kept", "TBMM", "TBMM"},
		{"acronym among words", "KDV dahildir", "KDV Dahildir"},
		{"acronym before a word", "TBMM üyesi", "TBMM Üyesi"},

		// An acronym that does contain a vowel cannot be told from a shouted
		// word by shape alone, so it is title-cased. TitleWith names them.
		{"acronym with vowels is not detected", "TÜBİTAK", "Tübitak"},

		// Capitals are never handed out, only kept.
		{"lowercase is never promoted", "aş pişirdim", "Aş Pişirdim"},
		{"mixed case is not an acronym", "Tbmm", "Tbmm"},

		{"extra spacing preserved", "  bolu   düzce  ", "  Bolu   Düzce  "},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Title(tt.input); got != tt.want {
				t.Errorf("Title(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Applying a case conversion twice must change nothing the second time.
func TestCaseConversionsAreIdempotent(t *testing.T) {
	inputs := []string{
		"istanbul", "İSTANBUL", "ığdır", "IĞDIR",
		"ahmet yılmaz", "istanbul'un", "", "abc-123",
	}

	for _, input := range inputs {
		for _, fn := range []struct {
			name string
			f    func(string) string
		}{
			{"ToUpper", ToUpper},
			{"ToLower", ToLower},
			{"Title", Title},
		} {
			once := fn.f(input)
			if twice := fn.f(once); twice != once {
				t.Errorf("%s is not idempotent for %q: %q then %q", fn.name, input, once, twice)
			}
		}
	}
}

func TestToASCII(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"dotted capital I", "İstanbul", "Istanbul"},
		{"dotless i", "Iğdır", "Igdir"},
		{"several Turkish letters", "Şanlıurfa", "Sanliurfa"},
		{"umlauts and sharp s", "Gümüşhane", "Gumushane"},
		{"cedilla", "Çanakkale", "Canakkale"},
		{"circumflex", "Hakkâri", "Hakkari"},
		{"lowercase throughout", "kırşehir", "kirsehir"},
		{"already ascii", "Ankara", "Ankara"},
		{"digits and punctuation", "34-İstanbul!", "34-Istanbul!"},
		{"empty", "", ""},

		// Only Turkish letters are in the table; other alphabets' accents are
		// left alone, so the result is not guaranteed to be pure ASCII.
		{"foreign accents pass through", "café", "café"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToASCII(tt.input); got != tt.want {
				t.Errorf("ToASCII(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Every Turkish letter must map to the ASCII letter of the same case, so that
// transliterating does not quietly change capitalization.
func TestToASCIIPreservesCase(t *testing.T) {
	pairs := []struct{ lower, upper string }{
		{"ı", "I"}, {"i", "İ"},
		{"ş", "Ş"}, {"ğ", "Ğ"}, {"ü", "Ü"},
		{"ö", "Ö"}, {"ç", "Ç"}, {"â", "Â"},
	}

	for _, p := range pairs {
		gotLower, gotUpper := ToASCII(p.lower), ToASCII(p.upper)
		if gotLower != strings.ToLower(gotLower) {
			t.Errorf("ToASCII(%q) = %q, want a lowercase result", p.lower, gotLower)
		}
		if gotUpper != strings.ToUpper(gotUpper) {
			t.Errorf("ToASCII(%q) = %q, want an uppercase result", p.upper, gotUpper)
		}
	}
}

// This is why PlateFromCity folds lookup keys through ToASCII rather than
// ToLower, and it is worth pinning down: the Turkish rule is right for display
// and wrong for lookup.
func TestToASCIIFoldsWhatToLowerCannot(t *testing.T) {
	const want = "istanbul"
	variants := []string{"İstanbul", "ISTANBUL", "istanbul", "Istanbul", "ıstanbul", "İSTANBUL"}

	// Folding through ToASCII collapses every spelling onto one key.
	for _, v := range variants {
		if got := strings.ToLower(ToASCII(v)); got != want {
			t.Errorf("strings.ToLower(ToASCII(%q)) = %q, want %q", v, got, want)
		}
	}

	// Folding through the Turkish ToLower does not: someone typing "ISTANBUL"
	// on an ASCII keyboard would produce "ıstanbul" and match nothing.
	if got := ToLower("ISTANBUL"); got == want {
		t.Fatalf("ToLower(%q) = %q, expected the Turkish dotless ı — has the I/ı rule gone?", "ISTANBUL", got)
	}
}

func TestToASCIIIsIdempotent(t *testing.T) {
	inputs := []string{"İstanbul", "Şanlıurfa", "Hakkâri", "Ankara", "café", ""}

	for _, input := range inputs {
		once := ToASCII(input)
		if twice := ToASCII(once); twice != once {
			t.Errorf("ToASCII is not idempotent for %q: %q then %q", input, once, twice)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"two words", "Şanlıurfa Merkez", "sanliurfa-merkez"},
		{"dotted capital I", "İstanbul Büyükşehir", "istanbul-buyuksehir"},
		{"dotless i", "Iğdır", "igdir"},
		{"circumflex", "Hakkâri Merkez", "hakkari-merkez"},
		{"already a slug", "ankara-cankaya", "ankara-cankaya"},
		{"digits kept", "34 İstanbul 2026", "34-istanbul-2026"},

		{"runs of separators collapse", "a  --  b", "a-b"},
		{"punctuation becomes separator", "Kadıköy, İstanbul!", "kadikoy-istanbul"},
		{"leading and trailing trimmed", "  --İzmir--  ", "izmir"},
		{"apostrophe is a separator", "İstanbul'un", "istanbul-un"},

		{"empty", "", ""},
		{"only separators", "---", ""},
		{"nothing keepable", "!!!", ""},

		// Letters ToASCII does not cover drop out instead of reaching the URL.
		{"foreign accent dropped", "café", "caf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.input); got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSlugifyWith(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		separator string
		want      string
	}{
		{"underscore", "Şanlıurfa Merkez", "_", "sanliurfa_merkez"},
		{"empty separator runs words together", "Şanlıurfa Merkez", "", "sanliurfamerkez"},
		{"multi-character separator", "a b c", "::", "a::b::c"},
		{"hyphen matches Slugify", "İstanbul Büyükşehir", "-", "istanbul-buyuksehir"},
		{"separator not added to empty result", "!!!", "_", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SlugifyWith(tt.input, tt.separator); got != tt.want {
				t.Errorf("SlugifyWith(%q, %q) = %q, want %q", tt.input, tt.separator, got, tt.want)
			}
		})
	}
}

// Slugify must be SlugifyWith bound to a hyphen, with no behaviour of its own.
func TestSlugifyMatchesSlugifyWithHyphen(t *testing.T) {
	inputs := []string{
		"Şanlıurfa Merkez", "İstanbul'un", "  --İzmir--  ", "34 İstanbul 2026", "", "!!!",
	}

	for _, input := range inputs {
		if got, want := Slugify(input), SlugifyWith(input, "-"); got != want {
			t.Errorf("Slugify(%q) = %q, but SlugifyWith(%q, \"-\") = %q", input, got, input, want)
		}
	}
}

// A slug must contain nothing but lowercase ASCII letters, digits, and the
// separator — that is the whole point of producing one.
func TestSlugifyProducesURLSafeOutput(t *testing.T) {
	inputs := []string{
		"Şanlıurfa Merkez", "İSTANBUL", "Iğdır", "Hakkâri", "café",
		"Kadıköy, İstanbul!", "  --İzmir--  ", "ÇANKAYA/ANKARA",
	}

	for _, input := range inputs {
		slug := Slugify(input)
		for _, r := range slug {
			isLower := r >= 'a' && r <= 'z'
			isDigit := r >= '0' && r <= '9'
			if !isLower && !isDigit && r != '-' {
				t.Errorf("Slugify(%q) = %q contains %q, which is not URL-safe", input, slug, r)
			}
		}
		if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
			t.Errorf("Slugify(%q) = %q has a leading or trailing separator", input, slug)
		}
		if strings.Contains(slug, "--") {
			t.Errorf("Slugify(%q) = %q has a doubled separator", input, slug)
		}
	}
}

func TestSlugifyIsIdempotent(t *testing.T) {
	inputs := []string{"Şanlıurfa Merkez", "İstanbul'un", "  --İzmir--  ", "café", ""}

	for _, input := range inputs {
		once := Slugify(input)
		if twice := Slugify(once); twice != once {
			t.Errorf("Slugify is not idempotent for %q: %q then %q", input, once, twice)
		}
	}
}

func ExampleToUpper() {
	fmt.Println(ToUpper("izmir"))

	// What the standard library does with the same word, for contrast.
	fmt.Println(strings.ToUpper("izmir"))
	// Output:
	// İZMİR
	// IZMIR
}

func ExampleToLower() {
	fmt.Println(ToLower("IĞDIR"))
	fmt.Println(strings.ToLower("IĞDIR"))
	// Output:
	// ığdır
	// iğdir
}

func ExampleTitle() {
	fmt.Println(Title("ahmet yılmaz"))
	fmt.Println(Title("AHMET YILMAZ"))

	// A word in capitals with no vowel is an initialism, and is left alone.
	fmt.Println(Title("KDV dahildir"))

	// A suffix after an apostrophe stays part of the word.
	fmt.Println(Title("istanbul'un"))
	// Output:
	// Ahmet Yılmaz
	// Ahmet Yılmaz
	// KDV Dahildir
	// İstanbul'un
}

func ExampleToASCII() {
	fmt.Println(ToASCII("Şanlıurfa"))
	fmt.Println(ToASCII("Hakkâri"))
	fmt.Println(ToASCII("Iğdır"))
	// Output:
	// Sanliurfa
	// Hakkari
	// Igdir
}

func ExampleSlugify() {
	fmt.Println(Slugify("Şanlıurfa Merkez"))
	fmt.Println(Slugify("Kadıköy, İstanbul!"))
	// Output:
	// sanliurfa-merkez
	// kadikoy-istanbul
}

func ExampleSlugifyWith() {
	fmt.Println(SlugifyWith("Şanlıurfa Merkez", "_"))
	// Output: sanliurfa_merkez
}

// The rule that keeps Title safe: it may preserve capitals a word already had,
// but it must never introduce them. Without this, an abbreviation spelled like
// an ordinary word — AS, AŞ — would corrupt every lower-case use of that word.
func TestTitleNeverPromotesToCapitals(t *testing.T) {
	inputs := []string{
		"aş pişirdim", "as kartı", "kdv dahildir", "tbmm üyesi",
		"tübitak projesi", "ptt şubesi",
	}

	for _, input := range inputs {
		got := Title(input)
		for _, word := range strings.Fields(got) {
			if isAllUpper(word) {
				t.Errorf("Title(%q) = %q: word %q was promoted to capitals", input, got, word)
			}
		}
	}
}

// Detecting "already in capitals" has to use the Turkish mapping. Comparing
// against strings.ToUpper would call "istanbul" capitalized, since the standard
// library uppercases it to "ISTANBUL" rather than "İSTANBUL".
func TestIsAllUpperUsesTurkishRules(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"TBMM", true},
		{"İSTANBUL", true},
		{"IĞDIR", true},
		{"ŞIRNAK", true},
		{"A", true},
		{"34-TR", true},

		{"istanbul", false},
		{"ığdır", false},
		{"Tbmm", false},
		{"İstanbul", false},

		// No letters at all: nothing to be capitalized.
		{"", false},
		{"1234", false},
		{"---", false},
	}

	for _, tt := range tests {
		if got := isAllUpper(tt.input); got != tt.want {
			t.Errorf("isAllUpper(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestHasVowel(t *testing.T) {
	withVowel := []string{
		"TÜBİTAK", "ABD", "AHMET", "aş", "Iğdır", "kâğıt", "mahkûm", "îman",
	}
	withoutVowel := []string{
		"TBMM", "KDV", "TCK", "PTT", "THY", "SGK", "SMS", "Ş", "1234", "",
	}

	for _, w := range withVowel {
		if !hasVowel(w) {
			t.Errorf("hasVowel(%q) = false, want true", w)
		}
	}
	for _, w := range withoutVowel {
		if hasVowel(w) {
			t.Errorf("hasVowel(%q) = true, want false", w)
		}
	}
}

func TestTitleWith(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		acronyms []string
		want     string
	}{
		{"names a vowel-bearing acronym", "TÜBİTAK projesi", []string{"TÜBİTAK"}, "TÜBİTAK Projesi"},
		{"several acronyms", "ABD ve AB", []string{"ABD", "AB"}, "ABD Ve AB"},

		// The caller may write the acronym however they like; matching folds
		// both sides with the Turkish uppercase rule.
		{"caller writes it lowercase", "TÜBİTAK", []string{"tübitak"}, "TÜBİTAK"},
		{"caller writes it title-cased", "TÜBİTAK", []string{"Tübitak"}, "TÜBİTAK"},

		// Naming an acronym does not make Title promote a lower-case word: the
		// writer did not capitalize it, so they did not mean the acronym.
		{"lowercase stays lowercase", "tübitak projesi", []string{"TÜBİTAK"}, "Tübitak Projesi"},

		// The vowelless rule keeps working alongside the named ones.
		{"heuristic still applies", "TBMM ve TÜBİTAK", []string{"TÜBİTAK"}, "TBMM Ve TÜBİTAK"},

		{"acronym absent from the text", "ahmet yılmaz", []string{"TÜBİTAK"}, "Ahmet Yılmaz"},
		{"unlisted acronym is title-cased", "ASELSAN", []string{"TÜBİTAK"}, "Aselsan"},
		{"empty input", "", []string{"TÜBİTAK"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TitleWith(tt.input, tt.acronyms...); got != tt.want {
				t.Errorf("TitleWith(%q, %q) = %q, want %q", tt.input, tt.acronyms, got, tt.want)
			}
		})
	}
}

// With no acronyms named, TitleWith must be Title exactly — no behaviour of its
// own, so a caller can reach for either without wondering which is which.
func TestTitleWithNoAcronymsMatchesTitle(t *testing.T) {
	inputs := []string{
		"ahmet yılmaz", "AHMET YILMAZ", "TBMM", "KDV dahildir",
		"istanbul'un", "3d yazıcı", "aş pişirdim", "", "  bolu   düzce  ",
	}

	for _, input := range inputs {
		if got, want := TitleWith(input), Title(input); got != want {
			t.Errorf("TitleWith(%q) = %q, but Title(%q) = %q", input, got, input, want)
		}
	}
}

// Naming acronyms must not open a way around the never-promote rule.
func TestTitleWithNeverPromotesToCapitals(t *testing.T) {
	acronyms := []string{"TÜBİTAK", "ABD", "AŞ", "AS", "KDV"}
	inputs := []string{
		"tübitak projesi", "abd doları", "aş pişirdim", "as kartı", "kdv dahildir",
	}

	for _, input := range inputs {
		got := TitleWith(input, acronyms...)
		for _, word := range strings.Fields(got) {
			if isAllUpper(word) {
				t.Errorf("TitleWith(%q, ...) = %q: word %q was promoted to capitals", input, got, word)
			}
		}
	}
}

func ExampleTitleWith() {
	// TÜBİTAK has vowels, so it cannot be told from a shouted word by shape
	// alone — naming it is what keeps its capitals.
	fmt.Println(TitleWith("TÜBİTAK projesi", "TÜBİTAK"))

	// Naming it does not capitalize a word the writer left in lower case.
	fmt.Println(TitleWith("tübitak projesi", "TÜBİTAK"))
	// Output:
	// TÜBİTAK Projesi
	// Tübitak Projesi
}
