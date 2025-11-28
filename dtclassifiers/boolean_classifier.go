package dtclassifiers

import (
	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&BooleanClassifier{})
	pvt.RegisterDataTypeAlias(pvt.BooleanType, pvt.BoolTypeSlug)
}

var _ pvt.DataTypeClassifier = (*BooleanClassifier)(nil)

type BooleanClassifier struct {
	*pvt.BaseDataTypeClassifier
}

func (v BooleanClassifier) Validate(value string) (err error) {
	if value == "true" {
		goto end
	}
	if value == "false" {
		goto end
	}
	err = NewErr(
		pvt.ErrInvalidBooleanFormat,
		"allowed_values", "true,false",
	)
end:
	return err
}

func (v BooleanClassifier) DataType() pvt.PVDataType {
	return pvt.BooleanType
}

func (v BooleanClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &BooleanClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (BooleanClassifier) Example() any {
	return "true"
}
func (BooleanClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.BooleanTypeSlug
}

func (BooleanClassifier) DefaultValue() *string {
	f := "false"
	return &f
}
