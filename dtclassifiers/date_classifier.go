package dtclassifiers

import (
	"fmt"
	"regexp"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&DateClassifier{})
}

var _ pvt.DataTypeClassifier = (*DateClassifier)(nil)

type DateClassifier struct {
	*pvt.BaseDataTypeClassifier
}

func (v DateClassifier) DataType() pvt.PVDataType {
	return pvt.DateType
}

func (v DateClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &DateClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

var (
	shortISO8601DateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	// multisegmentDateRegex matches yyyy, yyyy/mm, or yyyy/mm/dd in one regex
	multisegmentDateRegex = regexp.MustCompile(`^(\d{4})(?:/(\d{2})(?:/(\d{2}))?)?$`)
)

// Validate validates date values requiring date format (YYYY-MM-DD) by
// default. If multisegment then YYYY/MM/DD or partial dates (YYYY or YYYY/MM).
// More specific date format validation can be enforced via format constraints.
func (v DateClassifier) Validate(value string) (err error) {
	var matched bool
	var year, month, day int
	var scanFormat = "%d-%d-%d"
	var errMsgFormat = "yyyy-mm-dd"

	if v.MultiSegment {
		err = v.validateMultiSegmentDate(value)
		goto end
	}

	// Non-multisegment: require full date
	matched = shortISO8601DateRegex.MatchString(value)
	if !matched {
		err = NewErr(
			pvt.ErrFailedToMatchDateFormat,
			"expected_format", errMsgFormat,
		)
		goto end
	}

	// Parse and validate date components
	_, err = fmt.Sscanf(value, scanFormat, &year, &month, &day)
	if err != nil {
		err = NewErr(
			pvt.ErrFailedToParseDateFormat,
			"expected_format", errMsgFormat,
			err,
		)
		goto end
	}

	// Validate year
	if year < 1000 || year > 9999 {
		err = NewErr(
			pvt.ErrInvalidYearInDate,
			"invalid_year", year,
		)
		goto end
	}

	// Validate month range
	if month < 1 || month > 12 {
		err = NewErr(
			pvt.ErrInvalidMonthInDate,
			"invalid_month", month,
		)
		goto end
	}

	// Validate day range
	if day < 1 || day > 31 {
		err = NewErr(
			pvt.ErrInvalidDayInDate,
			"invalid_day", day,
		)
		goto end
	}

end:
	if err != nil {
		err = pvt.WithErr(err,
			pvt.ErrInvalidDateFormat,
		)
	}
	return err

}

func (v DateClassifier) validateMultiSegmentDate(value string) (err error) {
	var year, month, day int
	// Use regex with capture groups to match yyyy, yyyy/mm, or yyyy/mm/dd
	matches := multisegmentDateRegex.FindStringSubmatch(value)
	if matches == nil {
		err = NewErr(
			pvt.ErrFailedToMatchDateFormat,
			"expected_format", "yyyy, yyyy/mm, or yyyy/mm/dd",
		)
		goto end
	}

	// matches[0] is full match, matches[1] is year, matches[2] is month, matches[3] is day
	// Parse year (always present)
	_, err = fmt.Sscanf(matches[1], "%d", &year)
	if err != nil {
		err = NewErr(
			pvt.ErrFailedToParseDateFormat,
			"expected_format", "yyyy",
			err,
		)
		goto end
	}

	// Validate year
	if year < 1000 || year > 9999 {
		err = NewErr(
			pvt.ErrInvalidYearInDate,
			"invalid_year", year,
		)
		goto end
	}

	// Parse month if present
	if matches[2] != "" {
		_, err = fmt.Sscanf(matches[2], "%d", &month)
		if err != nil {
			err = NewErr(
				pvt.ErrFailedToParseDateFormat,
				"expected_format", "mm",
				err,
			)
			goto end
		}

		// Validate month range
		if month < 1 || month > 12 {
			err = NewErr(
				pvt.ErrInvalidMonthInDate,
				"invalid_month", month,
			)
			goto end
		}
	}

	// Parse day if present
	if matches[3] != "" {
		_, err = fmt.Sscanf(matches[3], "%d", &day)
		if err != nil {
			err = NewErr(
				pvt.ErrFailedToParseDateFormat,
				"expected_format", "dd",
				err,
			)
			goto end
		}

		// Validate day range
		if day < 1 || day > 31 {
			err = NewErr(
				pvt.ErrInvalidDayInDate,
				"invalid_day", day,
			)
			goto end
		}
	}

end:
	return err
}

func (DateClassifier) Example() any {
	return "1999-12-31" // TODO Might need to consider format constraints
}
func (DateClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.DateTypeSlug
}

func (DateClassifier) DefaultValue() *string {
	return nil
}
