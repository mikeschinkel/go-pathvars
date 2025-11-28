package dtclassifiers

import (
	"strconv"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&IntegerClassifier{})
	pvt.RegisterDataTypeAlias(pvt.IntegerType, pvt.IntTypeSlug)
}

var _ pvt.DataTypeClassifier = (*IntegerClassifier)(nil)

type IntegerClassifier struct {
	*pvt.BaseDataTypeClassifier
}

func (v IntegerClassifier) Validate(value string) (err error) {
	_, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		err = NewErr(pvt.ErrInvalidIntegerFormat, "value", value, err)
	}
	return err
}

func (v IntegerClassifier) DataType() pvt.PVDataType {
	return pvt.IntegerType
}

func (v IntegerClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &IntegerClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (IntegerClassifier) Example() any {
	return 123
}

func (IntegerClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.IntegerTypeSlug
}

func (IntegerClassifier) IndefiniteArticle() string {
	return "an"
}

func (IntegerClassifier) DefaultValue() *string {
	return nil
}
