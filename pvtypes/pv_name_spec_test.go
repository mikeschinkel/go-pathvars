package pvtypes

import (
	"testing"
)

func TestParsePVNameSpec(t *testing.T) {
	const varName = "name"
	tests := []struct {
		name     string
		nameSpec string
		wantSpec string
		wantErr  bool
	}{
		{
			name:     "Nominal",
			nameSpec: "name",
			wantErr:  false,
		},
		{
			name:     "Empty Name",
			nameSpec: "",
			wantErr:  true,
		},
		{
			name:     "Multi-segment",
			nameSpec: "name*",
			wantErr:  false,
		},
		{
			name:     "Optional",
			nameSpec: "name?",
			wantErr:  false,
		},
		{
			name:     "Optional but no name",
			nameSpec: "?",
			wantErr:  true,
		},
		{
			name:     "Optional with Default but no name",
			nameSpec: "?Default Value",
			wantErr:  true,
		},
		{
			name:     "Multi-segment but no name",
			nameSpec: "*",
			wantErr:  true,
		},
		{
			name:     "Optional with Default",
			nameSpec: "name?Default Value",
			wantErr:  false,
		},
		{
			name:     "Multi-segment,Optional",
			nameSpec: "name*?",
			wantErr:  false,
		},
		{
			name:     "Optional,Multi-segment",
			nameSpec: "name?*",
			wantSpec: "name*?",
			wantErr:  false,
		},
		{
			name:     "Optional,Multi-segment with Default",
			nameSpec: "name*?Default Value",
			wantErr:  false,
		},
		{
			name:     "Invalid Optional,Multi-segment with Default",
			nameSpec: "name?*Default Value",
			wantSpec: "name*?Default Value",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProps, err := ParseNameSpecProps(tt.nameSpec)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("ParsePVNameSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			wantSpec := tt.wantSpec
			if wantSpec == "" {
				wantSpec = tt.nameSpec
			}
			if gotProps == nil {
				t.Errorf("ParsePVNameSpec() props is nil for namespace %s", tt.nameSpec)
				return
			}
			if gotProps.Name != varName {
				t.Errorf("ParsePVNameSpec() name = %v, wantSpec %v", varName, gotProps.Name)
			}
			if gotProps.String() != wantSpec {
				t.Errorf("ParsePVNameSpec() gotProps = %v, wantSpec %v", gotProps.String(), wantSpec)
			}
		})
	}
}
