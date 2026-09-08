# trkit

Validation, normalization, and text utilities for data specific to Türkiye —
identity numbers, IBANs, license plates, phone numbers, and Turkish-aware text
handling.

> **Status: under development.** The API below is the target for `v0.1.0`. Nothing
> is released yet; import paths and signatures may still change.

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
trkit.IsValidTCKN("12345678950")   // checksum-verified national ID
trkit.IsValidIBAN("TR330006100519786457841326")
trkit.CityFromPlate(34)            // "İstanbul", nil
trkit.ToUpper("izmir")             // "İZMİR", not "IZMIR"
trkit.Slugify("Şanlıurfa Merkez")  // "sanliurfa-merkez"
trkit.NormalizePhone("0532 123 45 67")
```

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

### Text

| Function | Description |
|---|---|
| `ToUpper(s string) string` | Uppercase with Turkish i/İ rules |
| `ToLower(s string) string` | Lowercase with Turkish I/ı rules |
| `Title(s string) string` | Capitalizes the first letter of each word |
| `ToASCII(s string) string` | Transliterates Turkish letters to plain ASCII |
| `Slugify(s string) string` | URL slug, hyphen-separated |
| `SlugifyWith(s, sep string) string` | URL slug with a custom separator |

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

## License

MIT — see [LICENSE](LICENSE).
