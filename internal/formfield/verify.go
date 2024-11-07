package formfield

import (
	"errors"
	"fmt"
	"slices"
)

func FieldsEqual(fields1, fields2 []FormField) bool {
	if len(fields1) != len(fields2) {
		return false
	}

	for i := range len(fields1) {
		if fields1[i].FieldType != fields2[i].FieldType ||
			fields1[i].Prompt != fields2[i].Prompt ||
			fields1[i].Required != fields2[i].Required {
			return false
		}
		switch l := fields1[i].Data.(type) {
		case *CheckboxFieldData:
			r := fields2[i].Data.(*CheckboxFieldData)
			if !slices.Equal(l.Options, r.Options) {
				return false
			}
		case *RadioFieldData:
			r := fields2[i].Data.(*RadioFieldData)
			if !slices.Equal(l.Options, r.Options) {
				return false
			}
		case *TextFieldData:
			r := fields2[i].Data.(*TextFieldData)
			if l.Paragraph != r.Paragraph {
				return false
			}
		}
	}

	return true
}

func Validate(fields []FormField, filledFields []FilledFormField) error {
	if len(fields) != len(filledFields) {
		return errors.New("Field count mismatch")
	}

	for i := range len(fields) {
		if fields[i].FieldType != filledFields[i].FieldType {
			return errors.New("Field type mismatch")
		}

		if fields[i].Required && !fields[i].Required {
			return fmt.Errorf("Required field %q is missing", fields[i].Prompt)
		}

		switch l := fields[i].Data.(type) {
		case *CheckboxFieldData:
			r := filledFields[i].Data.(*FilledCheckboxFieldData)
			for _, opt := range r.SelectedOptions {
				if !slices.Contains(l.Options, opt) {
					return fmt.Errorf("Option %q not valid for field", opt)
				}
			}
		case *RadioFieldData:
			r := filledFields[i].Data.(*FilledRadioFieldData)
			if !slices.Contains(l.Options, r.SelectedOption) {
				return fmt.Errorf("Option %q not valid for field", r.SelectedOption)
			}
		case *TextFieldData:
		}
	}

	return nil
}
