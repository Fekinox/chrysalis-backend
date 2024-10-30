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
	AddFormFieldToForm(ctx context.Context, arg AddFormFieldToFormParams) (*FormField, error)
	AssignCurrentFormVersion(ctx context.Context, arg AssignCurrentFormVersionParams) (*CurrentFormVersion, error)
	CreateForm(ctx context.Context, creatorID pgtype.UUID) (*Form, error)
	CreateFormVersion(ctx context.Context, formID int64) (*CreateFormVersionRow, error)
	GetCurrentFormVersion(ctx context.Context, formID int64) ([]*GetCurrentFormVersionRow, error)
	NumTasksOnVersion(ctx context.Context, formVersionID int64) (int64, error)
}

var _ Querier = (*Queries)(nil)
