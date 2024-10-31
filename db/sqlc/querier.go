// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
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
	CreateForm(ctx context.Context, creatorID pgtype.UUID) (*Form, error)
	CreateFormVersion(ctx context.Context, arg CreateFormVersionParams) (*FormVersion, error)
	CreateTask(ctx context.Context, arg CreateTaskParams) (*CreateTaskRow, error)
	GetClientTasks(ctx context.Context, clientID pgtype.UUID) ([]*GetClientTasksRow, error)
	GetCurrentFormVersion(ctx context.Context, formID int64) ([]*GetCurrentFormVersionRow, error)
	GetServiceTasks(ctx context.Context, formID int64) ([]*GetServiceTasksRow, error)
	NumTasksOnVersion(ctx context.Context, formVersionID int64) (int64, error)
}

var _ Querier = (*Queries)(nil)
