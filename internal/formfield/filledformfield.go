package formfield

import (
	"context"
	"encoding/json"

	"github.com/Fekinox/chrysalis-backend/internal/db"
)

type FilledFormField struct {
	FieldType db.FieldType        `json:"type"`
	Filled    bool                `json:"filled"`
	Data      FilledFormFieldData `json:"data"`
}

type FilledFormFieldData interface {
	Create(
		ctx context.Context,
		db *db.Store,
		taskID int64,
		idx int32,
	) error
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

func (d *FilledCheckboxFieldData) Create(
	ctx context.Context,
	s *db.Store,
	taskID int64, idx int32) error {
	_, err := s.AddCheckboxFieldToTask(ctx, db.AddCheckboxFieldToTaskParams{
		TaskID:          taskID,
		Idx:             idx,
		SelectedOptions: d.SelectedOptions,
	})

	return err
}

func (d *FilledRadioFieldData) Create(
	ctx context.Context,
	s *db.Store,
	taskID int64, idx int32) error {
	_, err := s.AddRadioFieldToTask(ctx, db.AddRadioFieldToTaskParams{
		TaskID:         taskID,
		Idx:            idx,
		SelectedOption: &d.SelectedOption,
	})

	return err
}

func (d *FilledTextFieldData) Create(
	ctx context.Context,
	s *db.Store,
	taskID int64, idx int32) error {
	_, err := s.AddTextFieldToTask(ctx, db.AddTextFieldToTaskParams{
		TaskID:  taskID,
		Idx:     idx,
		Content: &d.Content,
	})
	return err
}
