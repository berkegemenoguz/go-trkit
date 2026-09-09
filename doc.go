// Package trkit provides validation, normalization, and text utilities for data
// specific to Türkiye.
//
// The package covers five areas:
//
//   - Identity numbers: checksum validation for TCKN (national ID) and VKN (tax ID).
//   - Banking: validation, normalization, and display formatting of Turkish IBANs.
//   - License plates: two-way lookup between province codes and province names.
//   - Text: Turkish-aware case conversion, ASCII transliteration, and URL slugs.
//   - Phone numbers: mobile and landline validation, plus E.164 normalization.
//
// Every function is pure: no side effects, no global state, and no network
// access. Validation is algorithmic and format-based only — the package never
// contacts official registries such as NVİ or GİB, so a number that passes
// these checks is well-formed, not necessarily issued to a real person.
//
// trkit depends on nothing outside the Go standard library.
//
// # Getting started
//
// The usual job is checking what a form submitted and putting it into the one
// shape worth storing:
//
//	if !trkit.IsValidTCKN(form.NationalID) {
//		return errors.New("invalid national identity number")
//	}
//
//	iban, err := trkit.NormalizeIBAN(form.IBAN)
//	if err != nil {
//		return fmt.Errorf("iban: %w", err)
//	}
//
//	phone, err := trkit.NormalizePhone(form.Phone)
//	if err != nil {
//		return fmt.Errorf("phone: %w", err)
//	}
//
//	// iban and phone are canonical now: "TR33..." and "+905321234567",
//	// whatever spacing or country prefix they arrived with.
//
// The Normalize functions validate as they go, so a value they return has
// already passed the matching IsValid check. Their errors are sentinel values,
// which callers compare with [errors.Is]:
//
//	if errors.Is(err, trkit.ErrInvalidIBAN) {
//		// ...
//	}
//
// # Turkish case conversion
//
// Turkish pairs i with İ and ı with I, which differs from the default Unicode
// mapping that the strings package applies. [ToUpper], [ToLower], and [Title]
// implement the Turkish rules, so trkit.ToUpper("izmir") yields "İZMİR" where
// strings.ToUpper yields "IZMIR".
//
// This correctness has a consequence worth knowing: because Turkish [ToLower]
// maps I to ı, it is the wrong tool for normalizing lookup keys typed on an
// ASCII keyboard. [PlateFromCity] therefore folds through [ToASCII] instead, so
// that "İstanbul", "ISTANBUL", and "Istanbul" all resolve to the same province.
//
// # Scope
//
// IBAN support is limited to Turkish (TR) IBANs. Validating arbitrary
// international IBANs correctly requires a per-country length table; without
// one, a mod-97 check alone accepts strings of the wrong length. Rather than
// ship that half-measure, the package validates what its name promises.
package trkit
