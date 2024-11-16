package models

import (
	"context"
	"errors"
	"fmt"
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

	ClientID  uuid.UUID          `json:"client_id"`
	Status    db.TaskStatus      `json:"status"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`

	Fields []formfield.FilledFormField `json:"fields"`
}

type CreateTaskParams struct {
	CreatorUsername string
	ClientID        uuid.UUID
	FormSlug        string
	Fields          []formfield.FilledFormField
}

type GetTaskParams struct {
	CreatorUsername string
	ServiceName     string
	TaskName        string
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

		var _ = taskHeader

		task = &Task{
			TaskID:        taskHeader.ID,
			TaskSlug:      taskHeader.Slug,
			FormID:        form.FormID,
			FormSlug:      form.Slug,
			FormVersionID: form.FormVersionID,

			ClientID:  taskHeader.ClientID,
			Status:    taskHeader.Status,
			CreatedAt: taskHeader.CreatedAt,
			Fields:    p.Fields,
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

		task = &Task{
			TaskID:        taskHeader.ID,
			TaskSlug:      p.TaskName,
			FormID:        form.FormID,
			FormSlug:      form.Slug,
			FormVersionID: form.FormVersionID,

			ClientID:  taskHeader.ClientID,
			Status:    taskHeader.Status,
			CreatedAt: taskHeader.CreatedAt,
			Fields:    parsedFields,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}
