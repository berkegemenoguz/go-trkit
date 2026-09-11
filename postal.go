package trkit

// A Turkish postal code is five digits. The first two are the plate code of the
// province it belongs to — 34 for İstanbul, 06 for Ankara — and the remaining
// three identify the district and delivery area within it.
const postalCodeLength = 5

// IsValidPostalCode reports whether code is a well-formed Turkish postal code:
// exactly five ASCII digits, the first two of which are the plate code of one
// of the 81 provinces.
//
// This is a check of form, not of existence. Whether 34999 is actually assigned
// to a delivery area can only be answered by PTT's full register, and a copy of
// that embedded here would go stale as codes are added and retired. A code that
// passes is shaped correctly and names a real province; nothing more.
//
// Separators and surrounding whitespace are not accepted.
func IsValidPostalCode(code string) bool {
	_, err := CityFromPostalCode(code)
	return err == nil
}

// CityFromPostalCode returns the province a postal code belongs to, read from
// its first two digits: "06100" yields "Ankara".
//
// A code that fails [IsValidPostalCode] returns an error matching
// [ErrInvalidPostalCode]. The code is taken as a string on purpose: parsing it
// as an integer would turn 06100 into 6100 and lose the province.
func CityFromPostalCode(code string) (string, error) {
	d, ok := parseDigits(code, postalCodeLength)
	if !ok {
		return "", ErrInvalidPostalCode
	}

	city, ok := plateToCity[d[0]*10+d[1]]
	if !ok {
		return "", ErrInvalidPostalCode
	}
	return city, nil
}
