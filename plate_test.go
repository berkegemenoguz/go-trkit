package trkit

import (
	"errors"
	"testing"
)

func TestCityFromPlate(t *testing.T) {
	tests := []struct {
		name string
		code int
		want string
	}{
		{"first code", 1, "Adana"},
		{"capital", 6, "Ankara"},
		{"dotted capital I", 34, "İstanbul"},
		{"renamed province", 33, "Mersin"},
		{"circumflex", 30, "Hakkâri"},
		{"dotless I", 76, "Iğdır"},
		{"last code", 81, "Düzce"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CityFromPlate(tt.code)
			if err != nil {
				t.Fatalf("CityFromPlate(%d) returned error %v, want none", tt.code, err)
			}
			if got != tt.want {
				t.Errorf("CityFromPlate(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestCityFromPlateRejects(t *testing.T) {
	for _, code := range []int{0, -1, 82, 100, 1000} {
		got, err := CityFromPlate(code)
		if !errors.Is(err, ErrUnknownPlateCode) {
			t.Errorf("CityFromPlate(%d) error = %v, want %v", code, err, ErrUnknownPlateCode)
		}
		if got != "" {
			t.Errorf("CityFromPlate(%d) = %q, want empty string on failure", code, got)
		}
	}
}

// The point of folding lookup keys through ToASCII: however a province name is
// typed — properly spelled, shouted, or hammered out on an ASCII keyboard — it
// must find the same province.
func TestPlateFromCityIgnoresSpellingAndCase(t *testing.T) {
	tests := []struct {
		want     int
		variants []string
	}{
		{34, []string{"İstanbul", "ISTANBUL", "istanbul", "Istanbul", "ıstanbul", "İSTANBUL", " istanbul "}},
		{76, []string{"Iğdır", "IĞDIR", "iğdır", "IGDIR", "igdir", "Igdir"}},
		{63, []string{"Şanlıurfa", "ŞANLIURFA", "sanliurfa", "SANLIURFA", "sanliURFA"}},
		{30, []string{"Hakkâri", "HAKKARİ", "hakkari", "Hakkari", "HAKKARI"}},
		{35, []string{"İzmir", "IZMIR", "izmir", "ızmır"}},
		{46, []string{"Kahramanmaraş", "KAHRAMANMARAS", "kahramanmaras"}},
	}

	for _, tt := range tests {
		for _, variant := range tt.variants {
			got, err := PlateFromCity(variant)
			if err != nil {
				t.Errorf("PlateFromCity(%q) returned error %v, want none", variant, err)
				continue
			}
			if got != tt.want {
				t.Errorf("PlateFromCity(%q) = %d, want %d", variant, got, tt.want)
			}
		}
	}
}

func TestPlateFromCityRejects(t *testing.T) {
	tests := []struct {
		name string
		city string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"not a province", "Kadıköy"},
		{"colloquial short form", "Afyon"},
		{"foreign city", "Berlin"},
		{"partial name", "İstanb"},
		{"digits", "34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PlateFromCity(tt.city)
			if !errors.Is(err, ErrUnknownCity) {
				t.Errorf("PlateFromCity(%q) error = %v, want %v", tt.city, err, ErrUnknownCity)
			}
			if got != 0 {
				t.Errorf("PlateFromCity(%q) = %d, want 0 on failure", tt.city, got)
			}
		})
	}
}

// Every code must survive a trip through its province name and back. This
// covers all eighty-one entries rather than the handful spelled out above, and
// it fails if two provinces ever fold to the same lookup key — a collision
// would otherwise lose one of them silently.
func TestPlateLookupsRoundTrip(t *testing.T) {
	for code := 1; code <= 81; code++ {
		city, err := CityFromPlate(code)
		if err != nil {
			t.Fatalf("CityFromPlate(%d) returned error %v, want none", code, err)
		}

		back, err := PlateFromCity(city)
		if err != nil {
			t.Fatalf("PlateFromCity(%q) returned error %v, want none", city, err)
		}
		if back != code {
			t.Errorf("round trip for code %d gave %d via %q", code, back, city)
		}
	}
}

// The reverse index must hold every province. If it is short, two names folded
// to the same key and one overwrote the other.
func TestCityIndexIsComplete(t *testing.T) {
	if len(plateToCity) != 81 {
		t.Fatalf("plateToCity has %d entries, want 81", len(plateToCity))
	}
	if len(cityToPlate) != len(plateToCity) {
		t.Errorf("cityToPlate has %d entries, want %d: two province names fold to one key",
			len(cityToPlate), len(plateToCity))
	}
}
