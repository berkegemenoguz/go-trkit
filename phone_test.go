package trkit

import "testing"

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
