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

		{"acronyms are flattened", "TBMM", "Tbmm"},
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

	// A suffix after an apostrophe stays part of the word.
	fmt.Println(Title("istanbul'un"))
	// Output:
	// Ahmet Yılmaz
	// Ahmet Yılmaz
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
