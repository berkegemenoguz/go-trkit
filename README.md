# trkit

[![Go Reference](https://pkg.go.dev/badge/github.com/berkegemenoguz/go-trkit.svg)](https://pkg.go.dev/github.com/berkegemenoguz/go-trkit)
[![CI](https://github.com/berkegemenoguz/go-trkit/actions/workflows/ci.yml/badge.svg)](https://github.com/berkegemenoguz/go-trkit/actions/workflows/ci.yml)

Validation, normalization, and text utilities for data specific to Türkiye —
identity numbers, IBANs, license plates, postal codes, phone numbers, and
Turkish-aware text handling.

> **Status: `v0.3.0`.** The API is complete and covered by tests, but the version is
> still `v0`, which under semantic versioning means it may change in a minor release
> while it settles. Pin a version if that matters to you.

## Why

Turkish validation helpers exist in Go, but scattered across single-purpose,
largely unmaintained packages — one for IBAN, another for TCKN, a third for
plates. `trkit` is a single, tested, documented package that covers them
together, with no dependencies outside the standard library.

## Install

```
go get github.com/berkegemenoguz/go-trkit
```

```go
import "github.com/berkegemenoguz/go-trkit"
```

The module path ends in `go-trkit`, but the package is named `trkit`.

## Quick start

```go
trkit.IsValidTCKN("12345678950")                       // true
trkit.IsValidIBAN("TR58 0000 0011 1111 1111 1111 11")  // true — spacing is ignored
trkit.ToUpper("izmir")                                 // "İZMİR", where strings.ToUpper gives "IZMIR"
trkit.Slugify("Şanlıurfa Merkez")                      // "sanliurfa-merkez"

city, _ := trkit.CityFromPlate(34)                      // "İstanbul"
phone, _ := trkit.NormalizePhone("0532 123 45 67")      // "+905321234567"
```

The `Normalize` functions validate as they go, so anything they return has already
passed the matching `IsValid` check — one call does both jobs.

## API

### Identity

| Function | Description |
|---|---|
| `IsValidTCKN(tckn string) bool` | TCKN (national ID) checksum validation |
| `IsValidVKN(vkn string) bool` | VKN (tax ID) checksum validation |

### Banking

| Function | Description |
|---|---|
| `IsValidIBAN(iban string) bool` | Turkish IBAN: structure, length, and mod-97 |
| `NormalizeIBAN(iban string) (string, error)` | Strips spaces, uppercases, validates |
| `FormatIBAN(iban string) (string, error)` | Groups in fours for display |

### License plates

| Function | Description |
|---|---|
| `CityFromPlate(code int) (string, error)` | Province name for a plate code |
| `PlateFromCity(city string) (int, error)` | Plate code for a province name |

### Postal codes

| Function | Description |
|---|---|
| `IsValidPostalCode(code string) bool` | Five digits, the first two a province plate code |
| `CityFromPostalCode(code string) (string, error)` | Province a postal code belongs to |

This checks form, not existence: a code that passes is shaped correctly and names a real
province, but only PTT's register can say whether it is actually assigned to a delivery
area.

### Text

| Function | Description |
|---|---|
| `ToUpper(s string) string` | Uppercase with Turkish i/İ rules |
| `ToLower(s string) string` | Lowercase with Turkish I/ı rules |
| `Title(s string) string` | Capitalizes the first letter of each word, lowercases the rest, keeps acronyms |
| `TitleWith(s string, acronyms ...string) string` | `Title`, told which vowel-bearing acronyms to keep |
| `ToASCII(s string) string` | Transliterates Turkish letters to plain ASCII |
| `Slugify(s string) string` | URL slug, hyphen-separated |
| `SlugifyWith(s, sep string) string` | URL slug with a custom separator |

`Title` lowercases the remainder of each word, so it normalizes both `ahmet yılmaz`
and `AHMET YILMAZ` to `Ahmet Yılmaz`. A word continues through letters, digits, and
apostrophes, so Turkish suffixes stay lowercase: `istanbul'un` becomes `İstanbul'un`.

**Acronyms.** A word already written in capitals with no vowel is left alone, since
Turkish words always carry one — `KDV dahildir` stays `KDV Dahildir`. An acronym that
does contain a vowel cannot be told from a shouted word by its shape, so name it:

```go
trkit.Title("TÜBİTAK projesi")                        // "Tübitak Projesi"
trkit.TitleWith("TÜBİTAK projesi", "TÜBİTAK")         // "TÜBİTAK Projesi"
```

Capitals are only ever kept, never introduced. `trkit.TitleWith("tübitak", "TÜBİTAK")`
returns `Tübitak`: a word written in lower case was not meant as an acronym. This is
what stops an abbreviation from corrupting an ordinary word spelled the same way — `aş`
stays `Aş`, whatever is on the acronym list.

`Title` formats; it does not decide what a word means. Treat its output as a sensible
default rather than a guarantee. It also does not apply the Turkish Language Institute's
rules for *titles*, which keep conjunctions like `ve` in lower case — `Title` capitalizes
every word, which is what normalizing a name or a place calls for.

### Phone numbers

| Function | Description |
|---|---|
| `IsValidMobilePhone(phone string) bool` | Turkish mobile number |
| `IsValidLandlinePhone(phone string) bool` | Turkish landline, checked against area codes |
| `IsValidPhone(phone string) bool` | Either of the above |
| `NormalizePhone(phone string) (string, error)` | Canonical E.164 form, `+905321234567` |

## Errors

Lookup and normalization functions return sentinel errors. Test them with
`errors.Is`:

```go
city, err := trkit.CityFromPlate(99)
if errors.Is(err, trkit.ErrUnknownPlateCode) {
    // ...
}
```

## Scope

- **No network access.** Validation is algorithmic and format-based. The package
  never contacts NVİ, GİB, or any other registry, so a value that passes these
  checks is well-formed — not necessarily issued to a real person or account.
- **IBAN support is Turkish only.** Validating arbitrary international IBANs
  requires a per-country length table; a mod-97 check on its own accepts strings
  of the wrong length. The package validates what its name promises rather than
  shipping that half-measure.
- **Zero dependencies.** Standard library only, by design.

## Testing

```
go test ./...
```

Beyond table-driven tests, the suite pins the properties that matter and would
otherwise rot quietly:

- **Round trips.** All 81 provinces resolve from code to name and back; a formatted
  IBAN normalizes to the number it came from.
- **Idempotence.** Normalizing or slugifying an already-processed value changes
  nothing, so callers need not track whether a function has run before.
- **Agreement.** `IsValidIBAN` and `NormalizeIBAN` must accept exactly the same
  inputs, as must `IsValidPhone` and `NormalizePhone`.
- **Cross-checked data.** The area code table is validated against the plate table:
  every province must be reachable by phone, spelled identically in both.
- **Fuzz targets** for every function that takes a string, asserting output
  invariants — a slug contains only URL-safe characters, a normalized value is
  itself valid input — rather than only the absence of a panic:

```
go test -fuzz=FuzzSlugify -fuzztime=30s
```

CI runs `gofmt`, `go vet`, and `go test -race -cover` on Go 1.21 (the version
`go.mod` declares) and on current stable.

## License

MIT — see [LICENSE](LICENSE).
