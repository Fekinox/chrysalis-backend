package formfield

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Fekinox/chrysalis-backend/internal/db"
)

var (
	ErrInvalidFormFieldRow = errors.New("Invalid form field row")
	ErrInvalidFormField    = errors.New("Invalid form field")
)

type FormField struct {
	FieldType db.FieldType  `json:"type"`
	Prompt    string        `json:"prompt"`
	Required  bool          `json:"required"`
	Data      FormFieldData `json:"data"`
}

type FormFieldData interface {
	formFieldData()
}

type CheckboxFieldData struct {
	Options []string `json:"options"`
}

type RadioFieldData struct {
	Options []string `json:"options"`
}

type TextFieldData struct {
	Paragraph bool `json:"paragraph"`
}

func (*CheckboxFieldData) formFieldData() {}
func (*RadioFieldData) formFieldData()    {}
func (*TextFieldData) formFieldData()     {}

func (ff *FormField) FromRow(ffr *db.GetFormFieldsRow) error {
	ff.FieldType = ffr.Ftype
	ff.Prompt = ffr.Prompt
	ff.Required = ffr.Required

	switch ffr.Ftype {
	case db.FieldTypeCheckbox:
		if ffr.CheckboxOptions == nil {
			return ErrInvalidFormFieldRow
		}
		ff.Data = &CheckboxFieldData{
			Options: ffr.CheckboxOptions,
		}
	case db.FieldTypeRadio:
		if ffr.RadioOptions == nil {
			return ErrInvalidFormFieldRow
		}
		ff.Data = &RadioFieldData{
			Options: ffr.RadioOptions,
		}
	case db.FieldTypeText:
		if ffr.TextParagraph == nil {
			return ErrInvalidFormFieldRow
		}
		ff.Data = &TextFieldData{
			Paragraph: *ffr.TextParagraph,
		}
	}

	return nil
}

func (ff *FormField) UnmarshalJSON(data []byte) error {
	type partial struct {
		FieldType   db.FieldType    `json:"type"`
		Prompt      string          `json:"prompt"`
		Required    bool            `json:"required"`
		PartialData json.RawMessage `json:"data"`
	}
	var p partial
	err := json.Unmarshal(data, &p)
	if err != nil {
		return err
	}

	ff.FieldType = p.FieldType
	ff.Prompt = p.Prompt
	ff.Required = p.Required

	switch ff.FieldType {
	case db.FieldTypeCheckbox:
		var test any
		json.Unmarshal(p.PartialData, test)
		fmt.Println(test)
		d := new(CheckboxFieldData)
		err = json.Unmarshal(p.PartialData, d)
		if err != nil {
			return ErrInvalidFormField
		}
		ff.Data = d
	case db.FieldTypeRadio:
		d := new(RadioFieldData)
		err = json.Unmarshal(p.PartialData, d)
		if err != nil {
			return ErrInvalidFormField
		}
		ff.Data = d
	case db.FieldTypeText:
		d := new(TextFieldData)
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
