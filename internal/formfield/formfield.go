package formfield

import (
	"context"
	"encoding/json"
	"errors"

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
	Create(
		ctx context.Context,
		db *db.Store,
		formVersionID, idx int64,
	) error
	ExecTemplate(name string) string
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

func (c *CheckboxFieldData) Create(
	ctx context.Context,
	store *db.Store,
	formVersionID int64, idx int64,
) error {
	_, err := store.AddCheckboxFieldToForm(
		ctx,
		db.AddCheckboxFieldToFormParams{
			FormVersionID: formVersionID,
			Idx:           idx,
			Options:       c.Options,
		},
	)
	return err
}

func (r *RadioFieldData) Create(
	ctx context.Context,
	store *db.Store,
	formVersionID int64, idx int64,
) error {
	_, err := store.AddRadioFieldToForm(
		ctx,
		db.AddRadioFieldToFormParams{
			FormVersionID: formVersionID,
			Idx:           idx,
			Options:       r.Options,
		},
	)
	return err
}

func (t *TextFieldData) Create(
	ctx context.Context,
	store *db.Store,
	formVersionID int64, idx int64,
) error {
	_, err := store.AddTextFieldToForm(
		ctx,
		db.AddTextFieldToFormParams{
			FormVersionID: formVersionID,
			Idx:           idx,
			Paragraph:     t.Paragraph,
		},
	)
	return err
}

func (c *CheckboxFieldData) ExecTemplate(name string) string {
	return ""
}

func (r *RadioFieldData) ExecTemplate(name string) string {
	return ""
}

func (t *TextFieldData) ExecTemplate(name string) string {
	return ""
}
