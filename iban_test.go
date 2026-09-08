package trkit

import "testing"

// validIBAN is the IBAN used throughout these tests: 26 characters, a zero in
// the reserved position, and a checksum that satisfies mod-97. It is a
// documentation example, not anyone's account.
const validIBAN = "TR330006100519786457841326"

func TestCleanIBAN(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"grouped in fours", "TR33 0006 1005 1978 6457 8413 26", validIBAN},
		{"already compact", validIBAN, validIBAN},
		{"lowercase", "tr330006100519786457841326", validIBAN},
		{"mixed case and spaces", "tr33 0006 1005 1978 6457 8413 26", validIBAN},
		{"surrounding spaces", "  " + validIBAN + "  ", validIBAN},
		{"irregular spacing", "TR33000  61005 19786457841326", validIBAN},
		{"empty", "", ""},

		// Hyphens survive cleaning on purpose, so that the structure check
		// rejects them instead of the cleaner quietly accepting them.
		{"hyphenated", "TR33-0006-1005", "TR33-0006-1005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanIBAN(tt.input); got != tt.want {
				t.Errorf("cleanIBAN(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHasIBANStructure(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", validIBAN, true},

		// Structure and checksum are separate concerns: a number may be shaped
		// correctly and still fail mod-97, which this function does not check.
		{"wrong check digits but right shape", "TR340006100519786457841326", true},

		{"empty", "", false},
		{"one character short", validIBAN[:25], false},
		{"one character long", validIBAN + "7", false},
		{"foreign country code", "DE330006100519786457841326", false},
		{"lowercase country code", "tr330006100519786457841326", false},
		{"letter among the digits", "TR33000610051978645784132X", false},
		{"reserved digit not zero", "TR330006110519786457841326", false},
		// Exactly 26 characters, so this fails on the digit check rather than
		// on length — which is the point of the case.
		{"hyphen where a digit belongs", "TR33-006100519786457841326", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasIBANStructure(tt.input); got != tt.want {
				t.Errorf("hasIBANStructure(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
