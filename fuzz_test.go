package trkit

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// The fuzz targets below check invariants rather than only the absence of a
// panic. A validator that crashed on odd input would be a bug, but so would one
// that produced a canonical form it then rejected, or a slug with a character
// that cannot go in a URL — and those failures are the ones a panic check would
// never find.

// textInputs seeds every text target with the cases that have caused trouble
// before: both dotted and dotless i, circumflexes, suffix apostrophes, accents
// from other alphabets, and invalid UTF-8.
var textInputs = []string{
	"", "İstanbul", "ISTANBUL", "ıstanbul", "Iğdır", "Şanlıurfa", "Hakkâri",
	"istanbul'un", "ankara’nın", "AHMET YILMAZ", "3d yazıcı", "café",
	"  --İzmir--  ", "ÇANKAYA/ANKARA", "TBMM", "\xff\xfe", "\x00",
}

func FuzzIdentityValidators(f *testing.F) {
	for _, seed := range []string{
		"", "12345678950", "1234567890", "0000000001", "abcdefghijk",
		"1234567895 ", "١٢٣٤٥٦٧٨٩٥٠", "\xff",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		// Neither validator may accept the other's length.
		if IsValidTCKN(s) && IsValidVKN(s) {
			t.Errorf("%q accepted as both a TCKN and a VKN", s)
		}
		if IsValidTCKN(s) && len(s) != 11 {
			t.Errorf("IsValidTCKN(%q) = true for a string of length %d", s, len(s))
		}
		if IsValidVKN(s) && len(s) != 10 {
			t.Errorf("IsValidVKN(%q) = true for a string of length %d", s, len(s))
		}
	})
}

func FuzzIBAN(f *testing.F) {
	for _, seed := range []string{
		"", validIBAN, "TR58 0000 0011 1111 1111 1111 11", "tr580000001111111111111111",
		shortIBANPassingMod97, longIBANPassingMod97, foreignIBANPassingMod97,
		"TR58-000001111111111111111", "\xff", "TR",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		normalized, err := NormalizeIBAN(s)

		if err != nil {
			if normalized != "" {
				t.Errorf("NormalizeIBAN(%q) returned %q alongside an error", s, normalized)
			}
			if IsValidIBAN(s) {
				t.Errorf("IsValidIBAN(%q) = true but NormalizeIBAN rejected it", s)
			}
			return
		}

		// Whatever comes back must itself be acceptable, or a value could not
		// survive a round trip through storage.
		if !IsValidIBAN(normalized) {
			t.Errorf("NormalizeIBAN(%q) = %q, which IsValidIBAN rejects", s, normalized)
		}
		if again, err := NormalizeIBAN(normalized); err != nil || again != normalized {
			t.Errorf("NormalizeIBAN is not idempotent for %q: %q then %q (err %v)", s, normalized, again, err)
		}
		if len(normalized) != ibanLength {
			t.Errorf("NormalizeIBAN(%q) = %q, length %d, want %d", s, normalized, len(normalized), ibanLength)
		}

		formatted, err := FormatIBAN(s)
		if err != nil {
			t.Errorf("FormatIBAN(%q) failed while NormalizeIBAN succeeded: %v", s, err)
			return
		}
		if back, err := NormalizeIBAN(formatted); err != nil || back != normalized {
			t.Errorf("FormatIBAN(%q) = %q does not normalize back to %q (got %q, err %v)",
				s, formatted, normalized, back, err)
		}
	})
}

func FuzzTextConversions(f *testing.F) {
	for _, seed := range textInputs {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		for _, fn := range []struct {
			name string
			f    func(string) string
		}{
			{"ToUpper", ToUpper},
			{"ToLower", ToLower},
			{"Title", Title},
			{"ToASCII", ToASCII},
		} {
			out := fn.f(s)
			if !utf8.ValidString(out) {
				t.Errorf("%s(%q) = %q, which is not valid UTF-8", fn.name, s, out)
			}
			if twice := fn.f(out); twice != out {
				t.Errorf("%s is not idempotent for %q: %q then %q", fn.name, s, out, twice)
			}
		}
	})
}

func FuzzSlugify(f *testing.F) {
	for _, seed := range textInputs {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		slug := Slugify(s)

		for _, r := range slug {
			isLower := r >= 'a' && r <= 'z'
			isDigit := r >= '0' && r <= '9'
			if !isLower && !isDigit && r != '-' {
				t.Fatalf("Slugify(%q) = %q contains %q, which is not URL-safe", s, slug, r)
			}
		}
		if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
			t.Errorf("Slugify(%q) = %q has a leading or trailing separator", s, slug)
		}
		if strings.Contains(slug, "--") {
			t.Errorf("Slugify(%q) = %q has a doubled separator", s, slug)
		}
		if twice := Slugify(slug); twice != slug {
			t.Errorf("Slugify is not idempotent for %q: %q then %q", s, slug, twice)
		}
	})
}

func FuzzPlateFromCity(f *testing.F) {
	for _, seed := range append([]string{"Adana", "istanbul", "ISTANBUL", "Berlin"}, textInputs...) {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		code, err := PlateFromCity(s)

		if err != nil {
			if code != 0 {
				t.Errorf("PlateFromCity(%q) returned %d alongside an error", s, code)
			}
			return
		}
		if code < 1 || code > 81 {
			t.Fatalf("PlateFromCity(%q) = %d, outside the valid range", s, code)
		}

		// A name that resolves must resolve back to the same code through its
		// official spelling.
		city, err := CityFromPlate(code)
		if err != nil {
			t.Fatalf("CityFromPlate(%d) failed for a code PlateFromCity produced: %v", code, err)
		}
		if back, err := PlateFromCity(city); err != nil || back != code {
			t.Errorf("round trip broke for %q: code %d, city %q, back %d (err %v)", s, code, city, back, err)
		}
	})
}

func FuzzPhone(f *testing.F) {
	for _, seed := range []string{
		"", "05321234567", "0532 123 45 67", "+90 532 123 45 67", "00905321234567",
		"0212 555 12 34", "0299 555 12 34", "9053212345", "0532-ABC-4567", "\xff",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		normalized, err := NormalizePhone(s)

		if err != nil {
			if normalized != "" {
				t.Errorf("NormalizePhone(%q) returned %q alongside an error", s, normalized)
			}
			if IsValidPhone(s) {
				t.Errorf("IsValidPhone(%q) = true but NormalizePhone rejected it", s)
			}
			return
		}

		if !IsValidPhone(normalized) {
			t.Errorf("NormalizePhone(%q) = %q, which IsValidPhone rejects", s, normalized)
		}
		if again, err := NormalizePhone(normalized); err != nil || again != normalized {
			t.Errorf("NormalizePhone is not idempotent for %q: %q then %q (err %v)", s, normalized, again, err)
		}
		if want := len(countryCallingCode) + nationalNumberLength; len(normalized) != want {
			t.Errorf("NormalizePhone(%q) = %q, length %d, want %d", s, normalized, len(normalized), want)
		}
		if !strings.HasPrefix(normalized, countryCallingCode) {
			t.Errorf("NormalizePhone(%q) = %q, which does not start with %q", s, normalized, countryCallingCode)
		}

		// Exactly one of the two kinds must claim it.
		mobile, landline := IsValidMobilePhone(normalized), IsValidLandlinePhone(normalized)
		if mobile == landline {
			t.Errorf("NormalizePhone(%q) = %q: mobile=%v landline=%v, want exactly one", s, normalized, mobile, landline)
		}
	})
}

func FuzzPostalCode(f *testing.F) {
	for _, seed := range []string{
		"", "34000", "06100", "01000", "81000", "00100", "82000",
		"3400a", " 3400", "34-00", "٣٤٠٠٠", "\xff",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		city, err := CityFromPostalCode(s)

		if got, want := IsValidPostalCode(s), err == nil; got != want {
			t.Errorf("IsValidPostalCode(%q) = %v, but CityFromPostalCode error = %v", s, got, err)
		}
		if err != nil {
			if city != "" {
				t.Errorf("CityFromPostalCode(%q) returned %q alongside an error", s, city)
			}
			return
		}

		if len(s) != postalCodeLength {
			t.Fatalf("CityFromPostalCode(%q) accepted a string of length %d", s, len(s))
		}

		// The province must be exactly the one the first two digits name.
		plate := int(s[0]-'0')*10 + int(s[1]-'0')
		if want, perr := CityFromPlate(plate); perr != nil || city != want {
			t.Errorf("CityFromPostalCode(%q) = %q, but CityFromPlate(%d) = %q (err %v)",
				s, city, plate, want, perr)
		}
	})
}
