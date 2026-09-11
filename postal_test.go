package trkit

import (
	"errors"
	"fmt"
	"testing"
)

func TestCityFromPostalCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"first province", "01000", "Adana"},
		{"leading zero kept", "06100", "Ankara"},
		{"dotted capital I", "34000", "İstanbul"},
		{"dotless capital I", "76000", "Iğdır"},
		{"last province", "81000", "Düzce"},

		// Only the first two digits are read; the delivery area is not checked.
		{"any delivery area digits", "35999", "İzmir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CityFromPostalCode(tt.code)
			if err != nil {
				t.Fatalf("CityFromPostalCode(%q) returned error %v, want none", tt.code, err)
			}
			if got != tt.want {
				t.Errorf("CityFromPostalCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestCityFromPostalCodeRejects(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"empty", ""},
		{"one digit short", "3400"},
		{"one digit long", "340000"},
		{"province 00", "00100"},
		{"province past 81", "82000"},
		{"province 99", "99000"},

		// Five characters each, so these fail on the digit check rather than
		// on length — which is what the cases are about.
		{"letter", "3400a"},
		{"leading space", " 3400"},
		{"separator", "34-00"},

		{"arabic-indic digits", "٣٤٠٠٠"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CityFromPostalCode(tt.code)
			if !errors.Is(err, ErrInvalidPostalCode) {
				t.Errorf("CityFromPostalCode(%q) error = %v, want %v", tt.code, err, ErrInvalidPostalCode)
			}
			if got != "" {
				t.Errorf("CityFromPostalCode(%q) = %q, want empty string on failure", tt.code, got)
			}
		})
	}
}

// Every province must be reachable from its postal codes, and must come back
// under exactly the name CityFromPlate gives it. This covers all eighty-one
// rather than the handful above, and ties postal codes to the plate table so
// the two cannot drift apart.
func TestPostalCodeCoversEveryProvince(t *testing.T) {
	for plate := 1; plate <= 81; plate++ {
		code := fmt.Sprintf("%02d000", plate)

		got, err := CityFromPostalCode(code)
		if err != nil {
			t.Errorf("CityFromPostalCode(%q) returned error %v, want none", code, err)
			continue
		}

		want, err := CityFromPlate(plate)
		if err != nil {
			t.Fatalf("CityFromPlate(%d) returned error %v, want none", plate, err)
		}
		if got != want {
			t.Errorf("CityFromPostalCode(%q) = %q, but CityFromPlate(%d) = %q", code, got, plate, want)
		}
	}
}

// The accepted province range must be exactly the plate table's: every
// two-digit prefix from 00 to 99, with only 01 through 81 allowed.
func TestPostalCodeProvinceRange(t *testing.T) {
	for prefix := 0; prefix <= 99; prefix++ {
		code := fmt.Sprintf("%02d123", prefix)
		want := prefix >= 1 && prefix <= 81

		if got := IsValidPostalCode(code); got != want {
			t.Errorf("IsValidPostalCode(%q) = %v, want %v", code, got, want)
		}
	}
}

// IsValidPostalCode is defined in terms of CityFromPostalCode, so the two must
// agree on every input.
func TestIsValidPostalCodeAgreesWithLookup(t *testing.T) {
	inputs := []string{
		"34000", "06100", "81000", "", "3400", "340000",
		"00100", "82000", "3400a", " 3400", "٣٤٠٠٠",
	}

	for _, code := range inputs {
		_, err := CityFromPostalCode(code)
		if got, want := IsValidPostalCode(code), err == nil; got != want {
			t.Errorf("IsValidPostalCode(%q) = %v, but CityFromPostalCode error = %v", code, got, err)
		}
	}
}

func ExampleCityFromPostalCode() {
	city, err := CityFromPostalCode("06100")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(city)
	// Output: Ankara
}

func ExampleIsValidPostalCode() {
	fmt.Println(IsValidPostalCode("34000"))
	fmt.Println(IsValidPostalCode("82000")) // there is no province 82
	fmt.Println(IsValidPostalCode("3400"))  // four digits
	// Output:
	// true
	// false
	// false
}

func ExampleCityFromPostalCode_invalid() {
	_, err := CityFromPostalCode("00100")
	fmt.Println(errors.Is(err, ErrInvalidPostalCode))
	// Output: true
}
