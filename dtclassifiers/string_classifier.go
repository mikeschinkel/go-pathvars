package dtclassifiers

import (
	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&StringClassifier{})
}

var _ pvt.DataTypeClassifier = (*StringClassifier)(nil)

type StringClassifier struct {
	*pvt.BaseDataTypeClassifier
}

func (v StringClassifier) Validate(value string) error {
	// Always valid
	return nil
}

func (v StringClassifier) DataType() pvt.PVDataType {
	return pvt.StringType
}

func (v StringClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &StringClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (StringClassifier) Example() any {
	return "abc"
}

func (StringClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.StringTypeSlug
}
