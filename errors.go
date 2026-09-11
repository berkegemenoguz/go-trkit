package trkit

import "errors"

// Errors reported by this package.
//
// Test for these with [errors.Is] rather than by comparing message strings, so
// that callers keep working if an error is later wrapped with more context.
var (
	// ErrUnknownPlateCode reports that a license plate code does not belong to
	// any province. Valid codes run from 1 to 81.
	ErrUnknownPlateCode = errors.New("trkit: unknown plate code")

	// ErrUnknownCity reports that a province name was not recognized.
	ErrUnknownCity = errors.New("trkit: unknown city name")

	// ErrInvalidIBAN reports that a string is not a well-formed Turkish IBAN.
	ErrInvalidIBAN = errors.New("trkit: invalid IBAN")

	// ErrInvalidPhone reports that a string is not a well-formed Turkish phone
	// number.
	ErrInvalidPhone = errors.New("trkit: invalid phone number")

	// ErrInvalidPostalCode reports that a string is not a well-formed Turkish
	// postal code, or that its first two digits do not name a province.
	ErrInvalidPostalCode = errors.New("trkit: invalid postal code")
)
