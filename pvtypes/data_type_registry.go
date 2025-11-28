package pvtypes

import (
	"strings"
)

type DataTypeClassifier interface {
	DataType() PVDataType
	Example() any
	Validate(value string) error
	MakeNew(args *DataTypeClassifierArgs) DataTypeClassifier
	Slug() PVDataTypeSlug
	IndefiniteArticle() string
	DefaultValue() *string
}

type BaseDataTypeClassifier struct {
	owner        DataTypeClassifier
	MultiSegment bool
}

func (c BaseDataTypeClassifier) IndefiniteArticle() string {
	return "a"
}

// DefaultValue returns the implicit default value for optional parameters without explicit defaults.
// For string-derived types, returns pointer to empty string.
// Types requiring explicit defaults should override this to return nil.
func (c BaseDataTypeClassifier) DefaultValue() *string {
	return new(string) // Returns pointer to "" for all string-derived types
}

type DataTypeClassifierArgs struct {
	MultiSegment bool
}

func NewBaseDataTypeClassifier(owner DataTypeClassifier, args *DataTypeClassifierArgs) *BaseDataTypeClassifier {
	if args == nil {
		args = &DataTypeClassifierArgs{}
	}
	return &BaseDataTypeClassifier{
		owner:        owner,
		MultiSegment: args.MultiSegment,
	}
}

var dataTypeClassifiersMap = make(map[PVDataType]DataTypeClassifier)
var dataTypeMap = make(map[PVDataTypeSlug]PVDataType)

func RegisterDataTypeClassifier(v DataTypeClassifier) {
	dt := v.DataType()
	v = v.MakeNew(nil)
	dataTypeClassifiersMap[dt] = v
	dataTypeMap[v.Slug()] = dt
}

func FindDataType(slug PVDataTypeSlug) (dt PVDataType) {
	dt, _ = dataTypeMap[PVDataTypeSlug(strings.ToLower(string(slug)))]
	return dt
}

func GetDataTypeClassifier(dt PVDataType) (v DataTypeClassifier, err error) {
	var ok bool

	if len(dataTypeClassifiersMap) == 0 {
		err = NewErr(ErrDataTypeClassifiersNotRegistered)
		goto end
	}
	v, ok = dataTypeClassifiersMap[dt]
	if !ok {
		err = NewErr(ErrDataTypeHasNoRegisteredClassifier, "data_type", dt.Slug())
		goto end
	}
end:
	return v, err
}
