package trkit

import (
	"errors"
	"testing"
)

// mobileNational is the synthetic mobile number used throughout these tests,
// written here in the compact national form the parser produces.
const mobileNational = "5321234567"

func TestNationalNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{"trunk prefix", "05321234567", mobileNational, true},
		{"grouped with spaces", "0532 123 45 67", mobileNational, true},
		{"country code with plus", "+905321234567", mobileNational, true},
		{"country code with plus and spaces", "+90 532 123 45 67", mobileNational, true},
		{"country code without plus", "905321234567", mobileNational, true},
		{"international access prefix", "00905321234567", mobileNational, true},
		{"bare national number", mobileNational, mobileNational, true},
		{"parentheses and hyphens", "(0532) 123-45-67", mobileNational, true},
		{"dots", "0532.123.45.67", mobileNational, true},
		{"slashes", "0532/123/45/67", mobileNational, true},
		{"surrounding whitespace", "  0532 123 45 67  ", mobileNational, true},
		{"landline is parsed too", "0212 555 12 34", "2125551234", true},

		{"empty", "", "", false},
		{"one digit short", "0532123456", "", false},
		{"one digit long", "053212345678", "", false},
		{"letters", "0532-ABC-4567", "", false},
		{"plus not at the start", "0532+1234567", "", false},
		{"foreign country code", "+15551234567", "", false},
		{"separators only", "-- () --", "", false},

		// Ten digits beginning with 90 are left alone: no country code is
		// stripped, because a national number is never longer than ten digits.
		// Whether the result is a real number is for the validators to say.
		{"ten digits starting with 90", "9053212345", "9053212345", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := nationalNumber(tt.input)
			if ok != tt.ok {
				t.Fatalf("nationalNumber(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("nationalNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidMobilePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"trunk prefix", "05321234567", true},
		{"grouped with spaces", "0532 123 45 67", true},
		{"country code with plus", "+90 532 123 45 67", true},
		{"country code without plus", "905321234567", true},
		{"international access prefix", "0090 532 123 45 67", true},
		{"bare national number", mobileNational, true},
		{"parentheses and hyphens", "(0532) 123-45-67", true},

		// The prefix rule is loose on purpose: any 5xx is accepted, so a newly
		// licensed operator does not need a release of this package.
		{"unfamiliar operator prefix", "0599 123 45 67", true},

		{"empty", "", false},
		{"landline", "0212 555 12 34", false},
		{"one digit short", "0532123456", false},
		{"one digit long", "053212345678", false},
		{"letters", "0532-ABC-4567", false},
		{"foreign number", "+1 555 123 4567", false},
		{"does not start with five", "0432 123 45 67", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidMobilePhone(tt.input); got != tt.want {
				t.Errorf("IsValidMobilePhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// Every province must be reachable by phone, and every province name in the
// area code table must be spelled the same way as in the plate table. Checking
// the two tables against each other catches an omission or a typo that neither
// table could reveal on its own.
func TestAreaCodesCoverEveryProvince(t *testing.T) {
	covered := make(map[string]int, len(plateToCity))
	for _, city := range areaCodeToCity {
		covered[city]++
	}

	for code, city := range plateToCity {
		if covered[city] == 0 {
			t.Errorf("province %q (plate %d) has no area code", city, code)
		}
	}

	known := make(map[string]bool, len(plateToCity))
	for _, city := range plateToCity {
		known[city] = true
	}
	for code, city := range areaCodeToCity {
		if !known[city] {
			t.Errorf("area code %d names %q, which is not a province in plate_data.go", code, city)
		}
	}

	// İstanbul is the one province split across two codes; anything else with
	// more than one is a duplicated entry.
	for city, count := range covered {
		if count > 1 && city != "İstanbul" {
			t.Errorf("province %q has %d area codes, want 1", city, count)
		}
	}
	if covered["İstanbul"] != 2 {
		t.Errorf("İstanbul has %d area codes, want 2 (European and Asian sides)", covered["İstanbul"])
	}
}

func TestIsValidLandlinePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"istanbul european side", "0212 555 12 34", true},
		{"istanbul asian side", "0216 555 12 34", true},
		{"ankara with country code", "+90 312 555 12 34", true},
		{"bursa with parentheses", "(0224) 555-12-34", true},
		{"izmir compact", "02325551234", true},
		{"bare national number", "2125551234", true},
		{"international access prefix", "00902125551234", true},

		{"empty", "", false},
		{"mobile", "0532 123 45 67", false},
		{"unknown area code", "0299 555 12 34", false},
		{"area code that is a mobile prefix", "0500 555 12 34", false},
		{"one digit short", "0212555123", false},
		{"one digit long", "021255512345", false},
		{"letters", "0212-ABC-1234", false},
		{"foreign number", "+1 555 123 4567", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidLandlinePhone(tt.input); got != tt.want {
				t.Errorf("IsValidLandlinePhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"mobile", "0532 123 45 67", true},
		{"landline", "0212 555 12 34", true},
		{"mobile with country code", "+90 532 123 45 67", true},
		{"landline bare", "3125551234", true},

		{"empty", "", false},
		{"unknown area code", "0299 555 12 34", false},
		{"wrong length", "0532123456", false},
		{"letters", "0532-ABC-4567", false},
		{"foreign number", "+1 555 123 4567", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPhone(tt.input); got != tt.want {
				t.Errorf("IsValidPhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// IsValidPhone must be exactly the union of the two specific checks, with no
// behaviour of its own.
func TestIsValidPhoneIsTheUnion(t *testing.T) {
	inputs := []string{
		"0532 123 45 67", "0212 555 12 34", "+90 312 555 12 34", "2125551234",
		"", "0299 555 12 34", "0532123456", "0532-ABC-4567", "+1 555 123 4567",
		"9053212345", "0599 123 45 67",
	}

	for _, input := range inputs {
		mobile, landline := IsValidMobilePhone(input), IsValidLandlinePhone(input)
		if got, want := IsValidPhone(input), mobile || landline; got != want {
			t.Errorf("IsValidPhone(%q) = %v, but mobile=%v landline=%v", input, got, mobile, landline)
		}
	}
}

// No number can be both, because area codes start with 2, 3, or 4 while mobile
// prefixes start with 5. Checking every area code in the table proves the two
// sets stay disjoint even if a code is added later.
func TestMobileAndLandlineNeverOverlap(t *testing.T) {
	for code := range areaCodeToCity {
		number := formatNational(code, "5551234")

		if !IsValidLandlinePhone(number) {
			t.Errorf("area code %d: %q is not accepted as a landline", code, number)
		}
		if IsValidMobilePhone(number) {
			t.Errorf("area code %d: %q is accepted as a mobile number too", code, number)
		}
	}
}

// formatNational builds a national number from an area code and the remaining
// seven digits.
func formatNational(areaCode int, rest string) string {
	digits := []byte{
		byte('0' + areaCode/100),
		byte('0' + (areaCode/10)%10),
		byte('0' + areaCode%10),
	}
	return string(digits) + rest
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"mobile with trunk prefix", "05321234567", "+905321234567"},
		{"mobile grouped", "0532 123 45 67", "+905321234567"},
		{"mobile with country code", "+90 532 123 45 67", "+905321234567"},
		{"mobile already canonical", "+905321234567", "+905321234567"},
		{"mobile bare", mobileNational, "+905321234567"},
		{"mobile with parentheses", "(0532) 123-45-67", "+905321234567"},
		{"mobile with access prefix", "00905321234567", "+905321234567"},
		{"landline", "0212 555 12 34", "+902125551234"},
		{"landline with country code", "+90 312 555 12 34", "+903125551234"},
		{"surrounding whitespace", "  0532 123 45 67  ", "+905321234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePhone(tt.input)
			if err != nil {
				t.Fatalf("NormalizePhone(%q) returned error %v, want none", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("NormalizePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizePhoneRejects(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"one digit short", "0532123456"},
		{"one digit long", "053212345678"},
		{"letters", "0532-ABC-4567"},
		{"unknown area code", "0299 555 12 34"},
		{"foreign number", "+1 555 123 4567"},
		{"neither mobile nor landline", "9053212345"},
		{"separators only", "-- () --"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePhone(tt.input)
			if !errors.Is(err, ErrInvalidPhone) {
				t.Errorf("NormalizePhone(%q) error = %v, want %v", tt.input, err, ErrInvalidPhone)
			}
			if got != "" {
				t.Errorf("NormalizePhone(%q) = %q, want empty string on failure", tt.input, got)
			}
		})
	}
}

// Normalizing twice must change nothing, and the canonical form must itself be
// accepted — otherwise a stored number could not be revalidated on read.
func TestNormalizePhoneIsIdempotentAndValid(t *testing.T) {
	inputs := []string{
		"05321234567", "0532 123 45 67", "+90 532 123 45 67",
		"0212 555 12 34", "(0224) 555-12-34", "00905321234567",
	}

	for _, input := range inputs {
		once, err := NormalizePhone(input)
		if err != nil {
			t.Fatalf("NormalizePhone(%q) returned error %v, want none", input, err)
		}

		twice, err := NormalizePhone(once)
		if err != nil {
			t.Fatalf("NormalizePhone(%q) returned error %v on second pass, want none", once, err)
		}
		if twice != once {
			t.Errorf("NormalizePhone is not idempotent for %q: %q then %q", input, once, twice)
		}

		if !IsValidPhone(once) {
			t.Errorf("IsValidPhone(%q) = false for a normalized number", once)
		}
	}
}

// IsValidPhone is defined in terms of NormalizePhone, so the two must agree on
// every input. This pins the relationship against a later change to either one.
func TestIsValidPhoneAgreesWithNormalize(t *testing.T) {
	inputs := []string{
		"0532 123 45 67", "0212 555 12 34", "+90 312 555 12 34", "2125551234",
		"", "0299 555 12 34", "0532123456", "0532-ABC-4567", "+1 555 123 4567",
		"9053212345", "0599 123 45 67", "-- () --",
	}

	for _, input := range inputs {
		_, err := NormalizePhone(input)
		if got, want := IsValidPhone(input), err == nil; got != want {
			t.Errorf("IsValidPhone(%q) = %v, but NormalizePhone error = %v", input, got, err)
		}
	}
}

// Every area code must normalize to a canonical number that starts with the
// country calling code and carries the code's own digits.
func TestNormalizePhoneCoversEveryAreaCode(t *testing.T) {
	for code := range areaCodeToCity {
		number := formatNational(code, "5551234")

		got, err := NormalizePhone(number)
		if err != nil {
			t.Errorf("NormalizePhone(%q) returned error %v for area code %d", number, err, code)
			continue
		}
		if want := countryCallingCode + number; got != want {
			t.Errorf("NormalizePhone(%q) = %q, want %q", number, got, want)
		}
	}
}
