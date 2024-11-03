// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type FieldType string

const (
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeRadio    FieldType = "radio"
	FieldTypeText     FieldType = "text"
)

func (e *FieldType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = FieldType(s)
	case string:
		*e = FieldType(s)
	default:
		return fmt.Errorf("unsupported scan type for FieldType: %T", src)
	}
	return nil
}

type NullFieldType struct {
	FieldType FieldType `json:"field_type"`
	Valid     bool      `json:"valid"` // Valid is true if FieldType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullFieldType) Scan(value interface{}) error {
	if value == nil {
		ns.FieldType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.FieldType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullFieldType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.FieldType), nil
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusApproved   TaskStatus = "approved"
	TaskStatusInprogress TaskStatus = "in progress"
	TaskStatusDelayed    TaskStatus = "delayed"
	TaskStatusComplete   TaskStatus = "complete"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

func (e *TaskStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TaskStatus(s)
	case string:
		*e = TaskStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for TaskStatus: %T", src)
	}
	return nil
}

type NullTaskStatus struct {
	TaskStatus TaskStatus `json:"task_status"`
	Valid      bool       `json:"valid"` // Valid is true if TaskStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTaskStatus) Scan(value interface{}) error {
	if value == nil {
		ns.TaskStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.TaskStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTaskStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.TaskStatus), nil
}

type CheckboxField struct {
	FormVersionID int64    `json:"form_version_id"`
	Idx           int64    `json:"idx"`
	Options       []string `json:"options"`
}

type CurrentFormVersion struct {
	FormID        int64 `json:"form_id"`
	FormVersionID int64 `json:"form_version_id"`
}

type FilledCheckboxField struct {
	TaskID          int64    `json:"task_id"`
	Idx             int32    `json:"idx"`
	SelectedOptions []string `json:"selected_options"`
}

type FilledForm struct {
	TaskID        int64  `json:"task_id"`
	FormVersionID *int64 `json:"form_version_id"`
}

type FilledFormField struct {
	TaskID int64     `json:"task_id"`
	Idx    int32     `json:"idx"`
	Ftype  FieldType `json:"ftype"`
	Filled bool      `json:"filled"`
}

type FilledRadioField struct {
	TaskID         int64   `json:"task_id"`
	Idx            int32   `json:"idx"`
	SelectedOption *string `json:"selected_option"`
}

type FilledTextField struct {
	TaskID  int64   `json:"task_id"`
	Idx     int32   `json:"idx"`
	Content *string `json:"content"`
}

type Form struct {
	ID        int64     `json:"id"`
	CreatorID uuid.UUID `json:"creator_id"`
}

type FormField struct {
	FormVersionID int64     `json:"form_version_id"`
	Idx           int64     `json:"idx"`
	Ftype         FieldType `json:"ftype"`
	Prompt        string    `json:"prompt"`
	Required      bool      `json:"required"`
}

type FormVersion struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	FormID      int64              `json:"form_id"`
}

type RadioField struct {
	FormVersionID int64    `json:"form_version_id"`
	Idx           int64    `json:"idx"`
	Options       []string `json:"options"`
}

type Task struct {
	ID        int64              `json:"id"`
	ClientID  uuid.UUID          `json:"client_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	Status    TaskStatus         `json:"status"`
	Slug      string             `json:"slug"`
}

type TextField struct {
	FormVersionID int64 `json:"form_version_id"`
	Idx           int64 `json:"idx"`
	Paragraph     bool  `json:"paragraph"`
}

type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
}
