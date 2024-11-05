// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	AddCheckboxFieldToForm(ctx context.Context, arg AddCheckboxFieldToFormParams) (*CheckboxField, error)
	AddCheckboxFieldToTask(ctx context.Context, arg AddCheckboxFieldToTaskParams) (*FilledCheckboxField, error)
	AddFilledFieldToTask(ctx context.Context, arg AddFilledFieldToTaskParams) (*FilledFormField, error)
	AddFormFieldToForm(ctx context.Context, arg AddFormFieldToFormParams) (*FormField, error)
	AddFormToTask(ctx context.Context, arg AddFormToTaskParams) (*FilledForm, error)
	AddRadioFieldToForm(ctx context.Context, arg AddRadioFieldToFormParams) (*RadioField, error)
	AddRadioFieldToTask(ctx context.Context, arg AddRadioFieldToTaskParams) (*FilledRadioField, error)
	AddTextFieldToForm(ctx context.Context, arg AddTextFieldToFormParams) (*TextField, error)
	AddTextFieldToTask(ctx context.Context, arg AddTextFieldToTaskParams) (*FilledTextField, error)
	AssignCurrentFormVersion(ctx context.Context, arg AssignCurrentFormVersionParams) (*CurrentFormVersion, error)
	CreateForm(ctx context.Context, arg CreateFormParams) (*CreateFormRow, error)
	CreateFormVersion(ctx context.Context, arg CreateFormVersionParams) (*FormVersion, error)
	CreateTask(ctx context.Context, arg CreateTaskParams) (*CreateTaskRow, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	DeleteForm(ctx context.Context, arg DeleteFormParams) error
	GetClientTasks(ctx context.Context, clientID uuid.UUID) ([]*GetClientTasksRow, error)
	GetCurrentFormVersionBySlug(ctx context.Context, arg GetCurrentFormVersionBySlugParams) (*GetCurrentFormVersionBySlugRow, error)
	GetFormFields(ctx context.Context, formVersionID int64) ([]*GetFormFieldsRow, error)
	GetFormHeaderBySlug(ctx context.Context, arg GetFormHeaderBySlugParams) (*GetFormHeaderBySlugRow, error)
	GetServiceTasks(ctx context.Context, formID int64) ([]*GetServiceTasksRow, error)
	GetUserByUUID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserFormHeaders(ctx context.Context, creatorID uuid.UUID) ([]*GetUserFormHeadersRow, error)
	NumTasksOnVersion(ctx context.Context, formVersionID int64) (int64, error)
}

var _ Querier = (*Queries)(nil)
