package trkit

import "strings"

// cityToPlate is the reverse of [plateToCity], keyed by the folded form of each
// province name so that lookups tolerate the many ways people type them.
var cityToPlate = buildCityIndex()

func buildCityIndex() map[string]int {
	index := make(map[string]int, len(plateToCity))
	for code, city := range plateToCity {
		index[cityLookupKey(city)] = code
	}
	return index
}

// cityLookupKey folds a province name into the form used as a map key.
//
// The order matters, and so does the choice of lowercaser. Transliterating
// first turns every Turkish letter into ASCII, and only then is the text
// lowercased with strings.ToLower rather than this package's [ToLower].
//
// Using the Turkish rule here would be actively wrong. It maps I to dotless ı,
// so "ISTANBUL" — what someone types on an ASCII keyboard — would fold to
// "ıstanbul" while the stored name "İstanbul" folds to "istanbul", and the two
// would never meet. The rule that is correct for displaying Turkish is the
// wrong rule for matching it.
func cityLookupKey(city string) string {
	return strings.ToLower(ToASCII(strings.TrimSpace(city)))
}

// CityFromPlate returns the province that a license plate code belongs to.
//
// Codes run from 1 to 81. Anything outside that returns an error matching
// [ErrUnknownPlateCode].
func CityFromPlate(code int) (string, error) {
	city, ok := plateToCity[code]
	if !ok {
		return "", ErrUnknownPlateCode
	}
	return city, nil
}

// PlateFromCity returns the license plate code of a province.
//
// Matching ignores case, Turkish letters, and surrounding whitespace, so
// "İstanbul", "ISTANBUL", "istanbul", and " Istanbul " all resolve to 34. An
// unrecognized name returns an error matching [ErrUnknownCity].
//
// Only official province names are recognized; colloquial short forms such as
// "Afyon" for Afyonkarahisar are not.
func PlateFromCity(city string) (int, error) {
	code, ok := cityToPlate[cityLookupKey(city)]
	if !ok {
		return 0, ErrUnknownCity
	}
	return code, nil
}
