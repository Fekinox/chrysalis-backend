package formfield

import (
	"encoding/json"

	"github.com/Fekinox/chrysalis-backend/internal/db"
)

type FilledFormField struct {
	FieldType db.FieldType        `json:"type"`
	Filled    bool                `json:"filled"`
	Data      FilledFormFieldData `json:"data"`
}

type FilledFormFieldData interface {
	filledFormFieldData()
}

type FilledCheckboxFieldData struct {
	SelectedOptions []string `json:"selectedOptions"`
}

type FilledRadioFieldData struct {
	SelectedOption string `json:"selectedOption"`
}

type FilledTextFieldData struct {
	Content string `json:"content"`
}

func (*FilledCheckboxFieldData) filledFormFieldData() {}
func (*FilledRadioFieldData) filledFormFieldData()    {}
func (*FilledTextFieldData) filledFormFieldData()     {}

func (ff *FilledFormField) FromRow(ffr *db.GetFilledFormFieldsRow) error {
	ff.FieldType = ffr.Ftype
	ff.Filled = ffr.Filled

	if !ff.Filled {
		return nil
	}

	switch ffr.Ftype {
	case db.FieldTypeCheckbox:
		if ffr.CheckboxOptions == nil {
			return ErrInvalidFormFieldRow
		}
		ff.Data = &FilledCheckboxFieldData{
			SelectedOptions: ffr.CheckboxOptions,
		}
	case db.FieldTypeRadio:
		if ffr.RadioOption == nil {
			return ErrInvalidFormFieldRow
		}
		ff.Data = &FilledRadioFieldData{
			SelectedOption: *ffr.RadioOption,
		}
	case db.FieldTypeText:
		if ffr.TextContent == nil {
			return ErrInvalidFormFieldRow
		}
		ff.Data = &FilledTextFieldData{
			Content: *ffr.TextContent,
		}
	}

	return nil
}

func (ff *FilledFormField) UnmarshalJSON(data []byte) error {
	type partial struct {
		FieldType   db.FieldType    `json:"type"`
		Filled      bool            `json:"filled"`
		PartialData json.RawMessage `json:"data"`
	}
	var p partial
	err := json.Unmarshal(data, &p)
	if err != nil {
		return err
	}

	ff.FieldType = p.FieldType
	ff.Filled = p.Filled

	if !ff.Filled {
		return nil
	}

	switch ff.FieldType {
	case db.FieldTypeCheckbox:
		d := new(FilledCheckboxFieldData)
		err = json.Unmarshal(p.PartialData, d)
		if err != nil {
			return ErrInvalidFormField
		}
		ff.Data = d
	case db.FieldTypeRadio:
		d := new(FilledRadioFieldData)
		err = json.Unmarshal(p.PartialData, d)
		if err != nil {
			return ErrInvalidFormField
		}
		ff.Data = d
	case db.FieldTypeText:
		d := new(FilledTextFieldData)
		err = json.Unmarshal(p.PartialData, d)
		if err != nil {
			return ErrInvalidFormField
		}
		ff.Data = d
	default:
		return ErrInvalidFormField
	}

	return nil
}
