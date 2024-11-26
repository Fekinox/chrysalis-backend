// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	AcknowledgeAllUpdatesForTask(ctx context.Context, taskID int64) error
	AcknowledgeAllUpdatesForUser(ctx context.Context, username string) error
	AcknowledgeUpdate(ctx context.Context, taskUpdateID int64) error
	AddCheckboxFieldToForm(ctx context.Context, arg AddCheckboxFieldToFormParams) (*CheckboxField, error)
	AddCheckboxFieldToTask(ctx context.Context, arg AddCheckboxFieldToTaskParams) (*FilledCheckboxField, error)
	AddFilledFieldToTask(ctx context.Context, arg AddFilledFieldToTaskParams) (*FilledFormField, error)
	AddFormFieldToForm(ctx context.Context, arg AddFormFieldToFormParams) (*FormField, error)
	AddRadioFieldToForm(ctx context.Context, arg AddRadioFieldToFormParams) (*RadioField, error)
	AddRadioFieldToTask(ctx context.Context, arg AddRadioFieldToTaskParams) (*FilledRadioField, error)
	AddTextFieldToForm(ctx context.Context, arg AddTextFieldToFormParams) (*TextField, error)
	AddTextFieldToTask(ctx context.Context, arg AddTextFieldToTaskParams) (*FilledTextField, error)
	AllNewUpdatesForUser(ctx context.Context, username string) ([]*AllNewUpdatesForUserRow, error)
	AssignCurrentFormVersion(ctx context.Context, arg AssignCurrentFormVersionParams) (*CurrentFormVersion, error)
	CreateForm(ctx context.Context, arg CreateFormParams) (*CreateFormRow, error)
	CreateFormVersion(ctx context.Context, arg CreateFormVersionParams) (*FormVersion, error)
	CreateTask(ctx context.Context, arg CreateTaskParams) (*CreateTaskRow, error)
	CreateTaskState(ctx context.Context, taskID int64) (*TaskState, error)
	CreateUpdate(ctx context.Context, arg CreateUpdateParams) (*TaskUpdate, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	DeleteForm(ctx context.Context, arg DeleteFormParams) error
	FindDiscrepancies(ctx context.Context, arg FindDiscrepanciesParams) ([]*FindDiscrepanciesRow, error)
	FindIfFormUnchanged(ctx context.Context, id int64) ([]int32, error)
	GetChrysalisStats(ctx context.Context) (*GetChrysalisStatsRow, error)
	GetCurrentFormVersionBySlug(ctx context.Context, arg GetCurrentFormVersionBySlugParams) (*GetCurrentFormVersionBySlugRow, error)
	GetFilledFormFields(ctx context.Context, arg GetFilledFormFieldsParams) ([]*GetFilledFormFieldsRow, error)
	GetFormFields(ctx context.Context, formVersionID int64) ([]*GetFormFieldsRow, error)
	GetFormHeaderBySlug(ctx context.Context, arg GetFormHeaderBySlugParams) (*GetFormHeaderBySlugRow, error)
	GetFormVersionById(ctx context.Context, formVersionID int64) (*GetFormVersionByIdRow, error)
	GetInboundTasks(ctx context.Context, creatorUsername string) ([]*GetInboundTasksRow, error)
	GetOutboundTasks(ctx context.Context, clientUsername string) ([]*GetOutboundTasksRow, error)
	GetServiceTasksBySlug(ctx context.Context, arg GetServiceTasksBySlugParams) ([]*GetServiceTasksBySlugRow, error)
	GetServiceTasksWithStatus(ctx context.Context, arg GetServiceTasksWithStatusParams) ([]*GetServiceTasksWithStatusRow, error)
	GetTaskByStatusAndIndex(ctx context.Context, arg GetTaskByStatusAndIndexParams) (*GetTaskByStatusAndIndexRow, error)
	GetTaskCounts(ctx context.Context, arg GetTaskCountsParams) ([]*GetTaskCountsRow, error)
	GetTaskHeader(ctx context.Context, arg GetTaskHeaderParams) (*GetTaskHeaderRow, error)
	GetUserByUUID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserFormHeaders(ctx context.Context, creatorID uuid.UUID) ([]*GetUserFormHeadersRow, error)
	InsertTask(ctx context.Context, arg InsertTaskParams) error
	LastUpdateTime(ctx context.Context, taskID int64) (pgtype.Timestamptz, error)
	RemoveTask(ctx context.Context, taskID1 int64) error
	ReorderTaskStatuses(ctx context.Context, arg ReorderTaskStatusesParams) error
	SwapTasks(ctx context.Context, arg SwapTasksParams) error
	UpdatePositionAndStatus(ctx context.Context, arg UpdatePositionAndStatusParams) error
	UpdateTaskStatus(ctx context.Context, arg UpdateTaskStatusParams) ([]int32, error)
}

var _ Querier = (*Queries)(nil)
