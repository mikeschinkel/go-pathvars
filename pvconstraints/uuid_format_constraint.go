package pvconstraints

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

// Note: UUIDFormatConstraint is not registered directly.
// It's handled by DateFormatConstraint.ParseBytes() based on dataType.

var _ pvtypes.Constraint = (*UUIDFormatConstraint)(nil)

// Package-level regex patterns compiled once for efficiency
var (
	ulidRegex      = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
	ksuidRegex     = regexp.MustCompile(`^[0-9A-Za-z]{27}$`)
	nanoidRegex    = regexp.MustCompile(`^[A-Za-z0-9_-]{21}$`)
	cuidRegex      = regexp.MustCompile(`^c[a-z0-9]{24}$`)
	snowflakeRegex = regexp.MustCompile(`^[0-9]{1,19}$`)
)

// Snowflake ID constants
const (
	// DefaultSnowflakeEpoch is Twitter's custom epoch: 2010-11-04 01:42:54.657 UTC
	DefaultSnowflakeEpoch int64 = 1288834974657

	// Snowflake bit structure
	snowflakeTimestampBits = 41
	snowflakeMachineIDBits = 10
	snowflakeSequenceBits  = 12

	// Masks for extracting components
	snowflakeSequenceMask  = (1 << snowflakeSequenceBits) - 1  // 4095
	snowflakeMachineIDMask = (1 << snowflakeMachineIDBits) - 1 // 1023
)

func init() {
	pvtypes.RegisterConstraint(&UUIDFormatConstraint{})
}

// UUIDFormatConstraint validates UUID formats
type UUIDFormatConstraint struct {
	pvtypes.BaseConstraint
	format      string
	formatParam string // Optional parameter (e.g., epoch for snowflake)
	validator   func(string) error
}

func NewUUIDFormatConstraint(format string, validator func(string) error) *UUIDFormatConstraint {
	return NewUUIDFormatConstraintWithParam(format, "", validator)
}

func NewUUIDFormatConstraintWithParam(format, formatParam string, validator func(string) error) *UUIDFormatConstraint {
	c := &UUIDFormatConstraint{
		format:      format,
		formatParam: formatParam,
		validator:   validator,
	}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *UUIDFormatConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{pvtypes.UUIDType}
}

func (c *UUIDFormatConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.FormatConstraintType
}

// ValidatesType returns true because format constraints perform their own type validation.
func (c *UUIDFormatConstraint) ValidatesType() bool {
	return true
}

// Example returns an Easter egg example UUID for the specific format.
// These examples have stories and meaning - see xmluisvr/pathvars/UUID.md for details.
// The error parameter is currently unused but maintains interface consistency.
func (c *UUIDFormatConstraint) Example(err error) any {
	switch c.format {
	case "v1":
		return "f81d4fae-7dec-11d0-a765-00a0c91e6bf6" // Canonical RFC example
	case "v2":
		return "02010203-0405-2607-8809-0a0b0c0d0e0f" // Crafted representative example
	case "v3":
		return "b1e42d72-b5b4-3286-9a25-e78457c1543b" // DNS namespace + www.example.com
	case "v4":
		return "deadbeef-cafe-4011-8123-b1d5c0d51234" // Hexadecimal "leet speak"
	case "v5":
		return "2a98f1f0-0a71-50e5-9d51-8650e68d9518" // URL namespace + https://www.example.com
	case "v6":
		return "1ec7c816-e5f6-6b2d-98ac-b0b3d64c1533" // Handcrafted sortable example
	case "v7":
		return "018d9f10-5341-7c91-9e73-b3c14d9b4b0e" // Common implementation example
	case "v8":
		return "20251018-b26a-8025-a12b-4c5d6e7f8a9b" // Date-encoded (Oct 18, 2025)
	case "v1-5", "v1to5":
		return "f81d4fae-7dec-11d0-a765-00a0c91e6bf6" // IETF example from UUIDv1
	case "v6-8", "v6to8":
		return "1ec7c816-e5f6-6b2d-98ac-b0b3d64c1533" // Version 6 example
	case "any", "generic":
		return "01234567-890A-BCDE-F012-34567890ABCD" // Incrementing hex digits
	case "ulid":
		return "01ARZ3NDEKTSV4RRFFQ69G5FAV" // From ULID spec README (July 30, 2017)
	case "ksuid":
		return "0ujsswThIGTUYm2K8FjOOfXtY1K" // Jan 1, 2025 timestamp
	case "nanoid":
		return "V1StGXR8_Z5jdHi6B-myT" // 21-char URL-safe example
	case "cuid":
		return "ckf0f9e5x0000q3yz4dq7a1qf" // From paralleldrive/cuid repo (2023-06-29 18:45:00 UTC)
	case "snowflake":
		return "1888944671579078978" // Wikipedia example (@Wikipedia tweet, 2025-02-10 13:34:39.256 UTC)
	default:
		return nil // No specific example
	}
}

func (c *UUIDFormatConstraint) Parse(value string, dataType pvtypes.PVDataType) (pvtypes.Constraint, error) {
	return ParseUUIDFormatConstraint(value)
}

func (c *UUIDFormatConstraint) Validate(value string) error {
	return c.validator(value)
}

func (c *UUIDFormatConstraint) Rule() string {
	return c.format
}

// ParseUUIDFormatConstraint parses UUID format specifications
// Supports format[value] and format[value:param] syntax (e.g., format[snowflake:1288834974657])
func ParseUUIDFormatConstraint(spec string) (constraint *UUIDFormatConstraint, err error) {
	var validator func(string) error
	var format, formatParam string

	// Check if spec contains a colon (e.g., "snowflake:1288834974657")
	parts := strings.SplitN(spec, ":", 2)
	format = strings.ToLower(parts[0])
	if len(parts) > 1 {
		formatParam = parts[1]
	}

	switch format {
	case "v1":
		validator = validateUUIDv1
	case "v2":
		validator = validateUUIDv2
	case "v3":
		validator = validateUUIDv3
	case "v4":
		validator = validateUUIDv4
	case "v5":
		validator = validateUUIDv5
	case "v6":
		validator = validateUUIDv6
	case "v7":
		validator = validateUUIDv7
	case "v8":
		validator = validateUUIDv8
	case "v1-5", "v1to5":
		validator = validateUUIDv1to5
	case "v6-8", "v6to8":
		validator = validateUUIDv6to8
	case "any", "generic":
		validator = validateUUIDGeneric
	case "ulid":
		validator = validateULID
	case "ksuid":
		validator = validateKSUID
	case "nanoid":
		validator = validateNanoID
	case "cuid":
		validator = validateCUID
	case "snowflake":
		// Snowflake with optional epoch parameter
		validator = createSnowflakeValidator(formatParam)
	default:
		err = pvtypes.NewErr(
			ErrUnsupportedUUIDFormat,
			"uuid_spec", spec,
		)
		goto end
	}

	constraint = NewUUIDFormatConstraintWithParam(format, formatParam, validator)

end:
	return constraint, err
}

// validateUUIDGeneric validates any standard UUID format (versions 1-8)
func validateUUIDGeneric(value string) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version < 1 || version > 8 {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrUUIDVersionOutOfRange1to8,
			"value", value,
			"version", version,
		)
	}
	return nil
}

// validateUUIDv1to5 validates UUID versions 1-5
func validateUUIDv1to5(value string) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version < 1 || version > 5 {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrUUIDVersionOutOfRange1to5,
			"value", value,
			"version", version,
		)
	}
	return nil
}

// validateUUIDv6to8 validates UUID versions 6-8
func validateUUIDv6to8(value string) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version < 6 || version > 8 {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrUUIDVersionOutOfRange6to8,
			"value", value,
			"version", version,
		)
	}
	return nil
}

// validateUUIDv1 validates UUID version 1 (time + MAC)
func validateUUIDv1(value string) error {
	return validateSpecificUUIDVersion(value, 1)
}

// validateUUIDv2 validates UUID version 2 (time + POSIX UID/GID)
func validateUUIDv2(value string) error {
	return validateSpecificUUIDVersion(value, 2)
}

// validateUUIDv3 validates UUID version 3 (name-based, MD5)
func validateUUIDv3(value string) error {
	return validateSpecificUUIDVersion(value, 3)
}

// validateUUIDv4 validates UUID version 4 (random)
func validateUUIDv4(value string) error {
	return validateSpecificUUIDVersion(value, 4)
}

// validateUUIDv5 validates UUID version 5 (name-based, SHA-1)
func validateUUIDv5(value string) error {
	return validateSpecificUUIDVersion(value, 5)
}

// validateUUIDv6 validates UUID version 6 (reordered v1)
func validateUUIDv6(value string) error {
	return validateSpecificUUIDVersion(value, 6)
}

// validateUUIDv7 validates UUID version 7 (Unix timestamp + random)
func validateUUIDv7(value string) error {
	return validateSpecificUUIDVersion(value, 7)
}

// validateUUIDv8 validates UUID version 8 (custom/experimental)
func validateUUIDv8(value string) error {
	return validateSpecificUUIDVersion(value, 8)
}

// validateSpecificUUIDVersion validates a specific UUID version
func validateSpecificUUIDVersion(value string, expectedVersion int) error {
	version, err := parseStandardUUID(value)
	if err != nil {
		return err
	}
	if version != expectedVersion {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrUUIDVersionMismatch,
			"value", value,
			"expected_version", expectedVersion,
			"actual_version", version,
		)
	}
	return nil
}

// parseStandardUUID parses and validates a standard UUID, returning the version
func parseStandardUUID(value string) (version int, err error) {
	var b [16]byte
	var hexStr string
	var variant int

	// Check basic shape: 36 chars with hyphens at correct positions
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		err = pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidUUIDShape,
			"value", value,
		)
		goto end
	}

	// Remove hyphens and decode hex
	hexStr = strings.ReplaceAll(value, "-", "")
	_, err = hex.Decode(b[:], []byte(hexStr))
	if err != nil {
		err = pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidUUIDHexEncoding,
			"value", value,
			err,
		)
		goto end
	}

	// Check variant bits (must be RFC 4122/9562: bits 10xx)
	variant = int((b[8] & 0xC0) >> 6)
	if variant != 0b10 {
		err = pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidUUIDVariant,
			"value", value,
			fmt.Errorf("variant=%08b", variant),
		)
		goto end
	}

	// Extract version from upper 4 bits of byte 6
	version = int(b[6] >> 4)
	if version < 1 || version > 8 {
		err = pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidUUIDVersion,
			"value", value,
			"version", version,
		)
		goto end
	}

end:
	return version, err
}

// validateULID validates ULID format (26 chars, Crockford Base32)
func validateULID(value string) error {
	// ULID: 26 characters, Crockford Base32 alphabet
	if !ulidRegex.MatchString(value) {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidULIDFormat,
			"value", value,
			"pattern", "26 chars of Crockford Base32",
		)
	}
	return nil
}

// validateKSUID validates KSUID format (27 chars, Base62)
func validateKSUID(value string) error {
	// KSUID: 27 characters, Base62 alphabet
	if !ksuidRegex.MatchString(value) {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidKSUIDFormat,
			"value", value,
			"pattern", "27 chars of Base62",
		)
	}
	return nil
}

// validateNanoID validates NanoID format (21 chars by default, URL-safe)
func validateNanoID(value string) error {
	// NanoID: 21 characters (default), URL-safe alphabet
	if !nanoidRegex.MatchString(value) {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidNanoIDFormat,
			"value", value,
			"pattern", "21 chars of URL-safe alphabet",
		)
	}
	return nil
}

// validateCUID validates CUID format (25 chars, starts with 'c', lowercase alphanumeric)
func validateCUID(value string) error {
	// CUID: 25 characters, starts with 'c', lowercase letters and digits
	if !cuidRegex.MatchString(value) {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidCUIDFormat,
			"value", value,
			"pattern", "25 chars starting with 'c', lowercase alphanumeric",
		)
	}
	return nil
}

// createSnowflakeValidator creates a Snowflake validator with optional custom epoch
func createSnowflakeValidator(epochParam string) func(string) error {
	// Parse custom epoch if provided, otherwise use Twitter's default
	epoch := DefaultSnowflakeEpoch
	if epochParam != "" {
		var err error
		var customEpoch int64
		_, err = fmt.Sscanf(epochParam, "%d", &customEpoch)
		if err != nil {
			// Return a validator that always fails with epoch parse error
			return func(value string) error {
				return pvtypes.NewErr(
					ErrParameterValidationFailed,
					ErrInvalidSnowflakeEpoch,
					"epoch_param", epochParam,
					err,
				)
			}
		}
		epoch = customEpoch
	}

	// Return validator closure with captured epoch
	return func(value string) error {
		return validateSnowflakeWithEpoch(value, epoch)
	}
}

// validateSnowflakeWithEpoch validates Snowflake ID format and components
func validateSnowflakeWithEpoch(value string, epoch int64) error {
	var id uint64
	var timestamp, machineID, sequence int64
	var err error

	// Basic format check: 1-19 digits (64-bit unsigned integer range)
	if !snowflakeRegex.MatchString(value) {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidSnowflakeFormat,
			"value", value,
			"pattern", "1-19 digits (64-bit integer)",
		)
	}

	// Parse as uint64
	_, err = fmt.Sscanf(value, "%d", &id)
	if err != nil {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrInvalidSnowflakeFormat,
			"value", value,
			err,
		)
	}

	// Extract components using bit manipulation
	sequence = int64(id & snowflakeSequenceMask)
	id >>= snowflakeSequenceBits

	machineID = int64(id & snowflakeMachineIDMask)
	id >>= snowflakeMachineIDBits

	timestamp = int64(id) // Remaining 41 bits

	// Validate timestamp is not negative (before epoch)
	if timestamp < 0 {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrSnowflakeTimestampNegative,
			"value", value,
			"timestamp_offset", timestamp,
			"epoch", epoch,
		)
	}

	// Validate timestamp is not unreasonably in the future (more than 1 year from now)
	// timestamp is milliseconds since epoch
	currentTime := time.Now().UnixMilli()
	actualTimestamp := epoch + timestamp
	oneYearFromNow := currentTime + (365 * 24 * 60 * 60 * 1000)

	if actualTimestamp > oneYearFromNow {
		return pvtypes.NewErr(
			ErrParameterValidationFailed,
			ErrSnowflakeTimestampInFuture,
			"value", value,
			"timestamp_offset", timestamp,
			"epoch", epoch,
			"actual_timestamp", actualTimestamp,
			"current_time", currentTime,
		)
	}

	// Components are valid (machineID and sequence are automatically within range due to bit masks)
	_ = machineID // 0-1023
	_ = sequence  // 0-4095

	return nil
}
