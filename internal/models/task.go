package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/Fekinox/chrysalis-backend/internal/genbytes"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrTooManyRetryAttempts = errors.New("Too many retry attempts")
	ErrFailedValidation     = errors.New("Failed form validation")
	ErrTaskNotFound         = errors.New("Task not found")
)

const MAX_TASK_RETRY_ATTEMPTS int = 10

var (
	dehyphenize = strings.NewReplacer("-", " ", "_", " ")
	hyphenize   = strings.NewReplacer(" ", "-")
)

func Dehyphenize(s string) string {
	return dehyphenize.Replace(s)
}

func Hyphenize(s string) string {
	return hyphenize.Replace(s)
}

func generateTaskSlug() (string, error) {
	slug, err := genbytes.GenRandomBytes(4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", slug), nil
}

type Task struct {
	TaskID        int64  `json:"task_id"`
	FormID        int64  `json:"form_id"`
	FormVersionID int64  `json:"form_version_id"`
	FormSlug      string `json:"form_slug"`
	TaskSlug      string `json:"task_slug"`

	ClientID       uuid.UUID `json:"client_id"`
	ClientUsername string    `json:"client_username"`

	TaskName    string             `json:"task_name"`
	TaskSummary string             `json:"task_summary"`
	Status      db.TaskStatus      `json:"status"`
	Index       int32              `json:"index"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`

	Fields []formfield.FilledFormField `json:"fields"`
}

type CreateTaskParams struct {
	CreatorUsername string
	ClientID        uuid.UUID
	FormSlug        string
	Fields          []formfield.FilledFormField
	TaskName        string
	TaskSummary     string
}

type GetTaskParams struct {
	CreatorUsername string
	ServiceName     string
	TaskName        string
}

type UpdateTaskParams struct {
	CreatorUsername string
	ServiceName     string
	TaskName        string
	Status          db.TaskStatus
}

type SwapTasksParams struct {
	CreatorUsername string
	ServiceName     string
	Task1Name       string
	Task2Name       string
}

type SwapTasksByStatusAndIdParams struct {
	CreatorUsername string
	ServiceName     string
	Status          db.TaskStatus
	Task1Index      int
	Task2Index      int
}

type MoveTaskParams struct {
	CreatorUsername string
	ServiceName     string
	OldStatus       db.TaskStatus
	NewStatus       db.TaskStatus
	OldIndex        int
	NewIndex        int
}

func CreateTask(
	ctx context.Context,
	d *db.Store,
	p CreateTaskParams,
) (task *Task, err error) {
	err = d.BeginFunc(ctx, func(s *db.Store) error {
		form, err := GetServiceForm(
			ctx,
			s,
			ServiceFormParams{
				Username: p.CreatorUsername,
				Service:  p.FormSlug,
			})
		if err != nil {
			return err
		}

		client, err := s.GetUserByUUID(ctx, p.ClientID)
		if err != nil {
			return err
		}

		var taskHeader *db.CreateTaskRow
		var attempts int
		for {
			err := s.BeginFunc(
				ctx,
				func(loopTx *db.Store) error {
					taskSlug, err := generateTaskSlug()
					if err != nil {
						return err
					}

					taskHeader, err = loopTx.
						CreateTask(ctx, db.CreateTaskParams{
							FormVersionID: form.FormVersionID,
							ClientID:      p.ClientID,
							Slug:          taskSlug,
							TaskName:      p.TaskName,
							TaskSummary:   p.TaskSummary,
						})

					return err
				},
			)

			var pgErr *pgconn.PgError

			if err == nil {
				break
			} else if errors.As(err, &pgErr) {
				if pgErr.Code != "23505" || pgErr.ConstraintName != "task_slug_unique" {
					return err
				}
			} else {
				return err
			}

			attempts++

			if attempts >= MAX_TASK_RETRY_ATTEMPTS {
				return ErrTooManyRetryAttempts
			}

			time.Sleep(time.Millisecond * 50)
		}

		if validationErr := formfield.Validate(form.Fields, p.Fields); validationErr != nil {
			return fmt.Errorf("%w: %v", ErrFailedValidation, validationErr)
		}

		for i, f := range p.Fields {
			_, err := s.AddFilledFieldToTask(ctx, db.AddFilledFieldToTaskParams{
				TaskID: taskHeader.ID,
				Idx:    int32(i),
				Ftype:  f.FieldType,
				Filled: f.Filled,
			})
			if err != nil {
				return err
			}

			if !f.Filled {
				continue
			}

			err = f.Data.Create(ctx, s, taskHeader.ID, int32(i))
			if err != nil {
				return err
			}
		}

		taskState, err := s.CreateTaskState(ctx, taskHeader.ID)
		if err != nil {
			return err
		}

		task = &Task{
			TaskID:        taskHeader.ID,
			TaskSlug:      taskHeader.Slug,
			FormID:        form.FormID,
			FormSlug:      form.Slug,
			FormVersionID: form.FormVersionID,

			ClientID:       taskHeader.ClientID,
			ClientUsername: client.Username,
			TaskName:       taskHeader.TaskName,
			TaskSummary:    taskHeader.TaskSummary,
			Status:         taskState.Status,
			Index:          taskState.Idx,
			CreatedAt:      taskHeader.CreatedAt,
			Fields:         p.Fields,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return task, nil
}

func GetTask(
	ctx context.Context,
	d *db.Store,
	p GetTaskParams,
) (task *Task, err error) {
	err = d.BeginFunc(ctx, func(s *db.Store) error {
		form, err := GetServiceForm(
			ctx,
			s,
			ServiceFormParams{
				Username: p.CreatorUsername,
				Service:  p.ServiceName,
			})
		if err != nil {
			return fmt.Errorf("%w: %s", ErrServiceNotFound, p.ServiceName)
		}

		taskHeader, err := s.GetTaskHeader(ctx, db.GetTaskHeaderParams{
			Username: p.CreatorUsername,
			FormSlug: p.ServiceName,
			TaskSlug: p.TaskName,
		})
		if err != nil {
			return fmt.Errorf("%w: %s", ErrTaskNotFound, p.TaskName)
		}

		rawFields, err := s.GetFilledFormFields(ctx, db.GetFilledFormFieldsParams{
			FormVersionID: form.FormVersionID,
			TaskSlug:      p.TaskName,
		})
		if err != nil {
			return ErrFieldsNotFound
		}

		parsedFields := make([]formfield.FilledFormField, len(rawFields))

		for i, f := range rawFields {
			err = parsedFields[i].FromRow(f)
			if err != nil {
				return err
			}
		}

		client, err := s.GetUserByUUID(ctx, taskHeader.ClientID)
		if err != nil {
			return err
		}

		task = &Task{
			TaskID:        taskHeader.ID,
			TaskSlug:      p.TaskName,
			FormID:        form.FormID,
			FormSlug:      form.Slug,
			FormVersionID: form.FormVersionID,

			ClientID:       taskHeader.ClientID,
			ClientUsername: client.Username,
			Status:         taskHeader.Status,
			Index:          taskHeader.Idx,
			TaskName:       taskHeader.TaskName,
			TaskSummary:    taskHeader.TaskSummary,
			CreatedAt:      taskHeader.CreatedAt,
			Fields:         parsedFields,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

func SwapTasks(
	ctx context.Context,
	d *db.Store,
	p SwapTasksParams,
) error {
	return d.BeginFunc(ctx, func(s *db.Store) error {
		task1, err := s.GetTaskHeader(ctx, db.GetTaskHeaderParams{
			Username: p.CreatorUsername,
			FormSlug: p.ServiceName,
			TaskSlug: p.Task1Name,
		})
		if err != nil {
			return nil
		}
		task2, err := s.GetTaskHeader(ctx, db.GetTaskHeaderParams{
			Username: p.CreatorUsername,
			FormSlug: p.ServiceName,
			TaskSlug: p.Task2Name,
		})
		if err != nil {
			return nil
		}

		if task1.Status != task2.Status {
			return errors.New("Tasks must have same status to be swapped")
		}

		return s.SwapTasks(ctx, db.SwapTasksParams{
			TaskID1: &task1.ID,
			TaskID2: &task2.ID,
		})
	})
}

func SwapTasksByStatusAndId(
	ctx context.Context,
	d *db.Store,
	p SwapTasksByStatusAndIdParams,
) error {
	return d.BeginFunc(ctx, func(s *db.Store) error {
		task1, err := s.GetTaskByStatusAndIndex(ctx, db.GetTaskByStatusAndIndexParams{
			CreatorUsername: p.CreatorUsername,
			FormSlug:        p.ServiceName,
			Status:          p.Status,
			Idx:             int32(p.Task1Index),
		})
		if err != nil {
			return nil
		}
		task2, err := s.GetTaskByStatusAndIndex(ctx, db.GetTaskByStatusAndIndexParams{
			CreatorUsername: p.CreatorUsername,
			FormSlug:        p.ServiceName,
			Status:          p.Status,
			Idx:             int32(p.Task2Index),
		})
		if err != nil {
			return nil
		}

		if task1.Status != task2.Status {
			return errors.New("Tasks must have same status to be swapped")
		}

		return s.SwapTasks(ctx, db.SwapTasksParams{
			TaskID1: &task1.ID,
			TaskID2: &task2.ID,
		})
	})
}

func MoveTask(
	ctx context.Context,
	d *db.Store,
	p MoveTaskParams,
) error {
	return d.BeginFunc(ctx, func(s *db.Store) error {
		task, err := s.GetTaskByStatusAndIndex(ctx, db.GetTaskByStatusAndIndexParams{
			CreatorUsername: p.CreatorUsername,
			FormSlug:        p.ServiceName,
			Status:          p.OldStatus,
			Idx:             int32(p.OldIndex),
		})
		if err != nil {
			return err
		}

		err = s.RemoveTask(ctx, task.ID)
		if err != nil {
			return err
		}

		err = s.ReorderTaskStatuses(ctx, db.ReorderTaskStatusesParams{
			CreatorUsername: p.CreatorUsername,
			FormSlug:        p.ServiceName,
		})
		if err != nil {
			return err
		}

		err = s.InsertTask(ctx, db.InsertTaskParams{
			TaskID:   task.ID,
			NewIndex: int32(p.NewIndex),
			Status:   p.NewStatus,
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func UpdateTaskStatus(
	ctx context.Context,
	d *db.Store,
	p UpdateTaskParams,
) error {
	return d.BeginFunc(ctx, func(s *db.Store) error {
		task, err := s.GetTaskHeader(ctx, db.GetTaskHeaderParams{
			Username: p.CreatorUsername,
			FormSlug: p.ServiceName,
			TaskSlug: p.TaskName,
		})
		if err != nil {
			return err
		}

		err = s.RemoveTask(ctx, task.ID)
		if err != nil {
			return err
		}

		err = s.ReorderTaskStatuses(ctx, db.ReorderTaskStatusesParams{
			CreatorUsername: p.CreatorUsername,
			FormSlug:        p.ServiceName,
		})
		if err != nil {
			return err
		}

		err = s.InsertTask(ctx, db.InsertTaskParams{
			TaskID:   task.ID,
			NewIndex: 0,
			Status:   p.Status,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
