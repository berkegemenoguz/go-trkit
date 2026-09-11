package trkit

import (
	"fmt"
	"testing"
)

// Every identity number in this file is synthetic. Each satisfies the published
// checksum but was generated from an obviously artificial digit sequence, so no
// real person's or company's number is committed to this repository.

func TestIsValidTCKN(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"sequential digits", "12345678950", true},
		{"repeated digits", "11111111110", true},
		{"mostly zeros", "10000000078", true},
		{"descending digits", "98765432150", true},

		{"empty", "", false},
		{"one digit short", "1234567895", false},
		{"one digit long", "123456789501", false},
		{"leading zero", "01234567890", false},
		{"wrong tenth digit", "12345678940", false},
		{"wrong eleventh digit", "12345678951", false},
		{"letter for digit", "1234567895a", false},
		{"trailing space", "1234567895 ", false},
		{"leading space", " 1234567895", false},
		{"separator", "1234-567895", false},
		{"all letters", "abcdefghijk", false},
		{"arabic-indic digits", "١٢٣٤٥٦٧٨٩٥٠", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTCKN(tt.input); got != tt.want {
				t.Errorf("IsValidTCKN(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidVKN(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"sequential digits", "1234567890", true},
		{"repeated digits", "1111111114", true},
		{"mostly zeros", "1000000000", true},
		{"leading zeros", "0000000001", true},

		{"empty", "", false},
		{"one digit short", "123456789", false},
		{"one digit long", "12345678901", false},
		{"wrong check digit", "1234567891", false},
		{"letter for digit", "123456789a", false},
		{"trailing space", "123456789 ", false},
		{"separator", "1234-56789", false},
		{"all letters", "abcdefghij", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidVKN(tt.input); got != tt.want {
				t.Errorf("IsValidVKN(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// A TCKN is 11 digits and a VKN is 10, so neither validator may accept the
// other's input. This guards against a length check drifting.
func TestIdentityValidatorsDoNotOverlap(t *testing.T) {
	for _, tckn := range []string{"12345678950", "11111111110", "98765432150"} {
		if IsValidVKN(tckn) {
			t.Errorf("IsValidVKN(%q) = true, want false: that is a TCKN", tckn)
		}
	}
	for _, vkn := range []string{"1234567890", "1111111114", "0000000001"} {
		if IsValidTCKN(vkn) {
			t.Errorf("IsValidTCKN(%q) = true, want false: that is a VKN", vkn)
		}
	}
}

// For any nine-digit prefix exactly one check digit completes a valid VKN.
// This pins the algorithm's shape without needing more published examples: an
// implementation that accepts too much or too little fails here even when the
// handful of table cases above still pass.
func TestVKNCheckDigitIsUnique(t *testing.T) {
	prefixes := []string{"123456789", "111111111", "100000000", "000000000", "987654321"}

	for _, prefix := range prefixes {
		accepted := 0
		for c := '0'; c <= '9'; c++ {
			if IsValidVKN(prefix + string(c)) {
				accepted++
			}
		}
		if accepted != 1 {
			t.Errorf("prefix %q: %d check digits accepted, want exactly 1", prefix, accepted)
		}
	}
}

// The same property for TCKN: of the 100 possible checksum pairs, exactly one
// completes a valid number.
func TestTCKNCheckDigitsAreUnique(t *testing.T) {
	prefixes := []string{"123456789", "111111111", "100000000", "987654321"}

	for _, prefix := range prefixes {
		accepted := 0
		for tenth := '0'; tenth <= '9'; tenth++ {
			for eleventh := '0'; eleventh <= '9'; eleventh++ {
				if IsValidTCKN(prefix + string(tenth) + string(eleventh)) {
					accepted++
				}
			}
		}
		if accepted != 1 {
			t.Errorf("prefix %q: %d checksum pairs accepted, want exactly 1", prefix, accepted)
		}
	}
}

func ExampleIsValidTCKN() {
	fmt.Println(IsValidTCKN("12345678950"))
	fmt.Println(IsValidTCKN("12345678951")) // last digit does not check out
	// Output:
	// true
	// false
}

func ExampleIsValidVKN() {
	fmt.Println(IsValidVKN("1234567890"))
	fmt.Println(IsValidVKN("1234567891")) // check digit does not match
	// Output:
	// true
	// false
}

func TestIsValidYKN(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"sequential digits", "99123456740", true},
		{"mostly zeros", "99000000042", true},
		{"all nines", "99999999990", true},
		{"repeated ones", "99111111194", true},

		// A TCKN has the right shape and checksum but not the 99 prefix.
		{"a TCKN", "12345678950", false},
		{"TCKN starting with 98", "98765432150", false},

		{"empty", "", false},
		{"one digit short", "9912345674", false},
		{"one digit long", "991234567401", false},
		{"wrong check digit", "99123456741", false},
		{"letter for digit", "9912345674a", false},
		{"separator", "9912-345674", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidYKN(tt.input); got != tt.want {
				t.Errorf("IsValidYKN(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// The same property TestTCKNCheckDigitsAreUnique pins for TCKN: of the 100
// possible checksum pairs, exactly one completes a valid YKN.
func TestYKNCheckDigitsAreUnique(t *testing.T) {
	prefixes := []string{"991234567", "990000000", "999999999", "991111111"}

	for _, prefix := range prefixes {
		accepted := 0
		for tenth := '0'; tenth <= '9'; tenth++ {
			for eleventh := '0'; eleventh <= '9'; eleventh++ {
				if IsValidYKN(prefix + string(tenth) + string(eleventh)) {
					accepted++
				}
			}
		}
		if accepted != 1 {
			t.Errorf("prefix %q: %d checksum pairs accepted, want exactly 1", prefix, accepted)
		}
	}
}

// A YKN is 11 digits and a VKN is 10, so neither validator may accept the
// other's input.
func TestYKNAndVKNDoNotOverlap(t *testing.T) {
	for _, ykn := range []string{"99123456740", "99000000042", "99999999990"} {
		if IsValidVKN(ykn) {
			t.Errorf("IsValidVKN(%q) = true, want false: that is a YKN", ykn)
		}
	}
	for _, vkn := range []string{"1234567890", "1111111114", "0000000001"} {
		if IsValidYKN(vkn) {
			t.Errorf("IsValidYKN(%q) = true, want false: that is a VKN", vkn)
		}
	}
}

func ExampleIsValidYKN() {
	fmt.Println(IsValidYKN("99123456740"))
	fmt.Println(IsValidYKN("12345678950")) // a TCKN: right checksum, wrong prefix
	// Output:
	// true
	// false
}
