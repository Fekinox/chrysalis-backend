package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserNotFound    = errors.New("User not found")
	ErrServiceNotFound = errors.New("Service not found")
	ErrFieldsNotFound  = errors.New("Fields not found")
	ErrUnchangedForm   = errors.New("Form is unchanged")
)

type ServiceForm struct {
	FormID        int64                 `json:"id"`
	CreatorID     uuid.UUID             `json:"creator_id"`
	Slug          string                `json:"slug"`
	FormVersionID int64                 `json:"form_version_id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	CreatedAt     pgtype.Timestamptz    `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz    `json:"updated_at"`
	Fields        []formfield.FormField `json:"fields"`
}

type ServiceFormParams struct {
	Username string
	Service  string
}

type CreateServiceVersionParams struct {
	CreatorID   uuid.UUID
	ServiceSlug string
	Title       string
	Description string
	Fields      []formfield.FormField
}

func GetServiceForm(ctx context.Context, d interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}, q *db.Queries, p ServiceFormParams) (form *ServiceForm, err error) {
	err = pgx.BeginFunc(ctx, d, func(tx pgx.Tx) error {
		qtx := q.WithTx(tx)

		user, err := qtx.GetUserByUsername(ctx, p.Username)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrUserNotFound, p.Username)
		}

		params := db.GetCurrentFormVersionBySlugParams{
			Slug:      p.Service,
			CreatorID: user.ID,
		}

		service, err := qtx.GetCurrentFormVersionBySlug(ctx, params)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrServiceNotFound, p.Service)
		}

		rawFields, err := qtx.GetFormFields(
			ctx,
			service.FormVersionID,
		)
		if err != nil {
			return ErrFieldsNotFound
		}

		parsedFields := make([]formfield.FormField, len(rawFields))

		for i, f := range rawFields {
			err = parsedFields[i].FromRow(f)
			if err != nil {
				return err
			}
		}

		form = &ServiceForm{
			FormID:        service.ID,
			CreatorID:     service.CreatorID,
			Slug:          service.Slug,
			FormVersionID: service.FormVersionID,
			Name:          service.Name,
			Description:   service.Description,
			CreatedAt:     service.CreatedAt,
			UpdatedAt:     service.UpdatedAt,
			Fields:        parsedFields,
		}

		return nil
	})
	return form, err
}

func createServiceVersion(ctx context.Context, d interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}, q *db.Queries, p CreateServiceVersionParams) (form *ServiceForm, err error) {
	err = pgx.BeginFunc(ctx, d, func(tx pgx.Tx) error {
		qtx := q.WithTx(tx)

		frm, err := qtx.GetFormHeaderBySlug(ctx, db.GetFormHeaderBySlugParams{
			CreatorID: p.CreatorID,
			Slug:      p.ServiceSlug,
		})
		if err != nil {
			return err
		}

		version, err := qtx.CreateFormVersion(ctx, db.CreateFormVersionParams{
			FormID:      frm.ID,
			Name:        p.Title,
			Description: p.Description,
		})
		if err != nil {
			return err
		}

		for i, f := range p.Fields {
			_, err := qtx.AddFormFieldToForm(
				ctx,
				db.AddFormFieldToFormParams{
					FormVersionID: version.ID,
					Idx:           int64(i),
					Ftype:         f.FieldType,
					Prompt:        f.Prompt,
					Required:      f.Required,
				},
			)
			if err != nil {
				return err
			}

			switch f.FieldType {
			case db.FieldTypeCheckbox:
				d, ok := f.Data.(*formfield.CheckboxFieldData)
				if !ok {
					return formfield.ErrInvalidFormField
				}
				_, err := qtx.AddCheckboxFieldToForm(
					ctx,
					db.AddCheckboxFieldToFormParams{
						FormVersionID: version.ID,
						Idx:           int64(i),
						Options:       d.Options,
					},
				)
				if err != nil {
					return err
				}
			case db.FieldTypeRadio:
				d, ok := f.Data.(*formfield.RadioFieldData)
				if !ok {
					return formfield.ErrInvalidFormField
				}
				_, err := qtx.AddRadioFieldToForm(
					ctx,
					db.AddRadioFieldToFormParams{
						FormVersionID: version.ID,
						Idx:           int64(i),
						Options:       d.Options,
					},
				)
				if err != nil {
					return err
				}
			case db.FieldTypeText:
				d, ok := f.Data.(*formfield.TextFieldData)
				if !ok {
					return formfield.ErrInvalidFormField
				}
				_, err := qtx.AddTextFieldToForm(
					ctx,
					db.AddTextFieldToFormParams{
						FormVersionID: version.ID,
						Idx:           int64(i),
						Paragraph:     d.Paragraph,
					},
				)
				if err != nil {
					return err
				}
			}
		}

		form = &ServiceForm{
			FormID:        version.FormID,
			FormVersionID: version.ID,
			CreatorID:     frm.CreatorID,
			Slug:          frm.Slug,
			Name:          version.Name,
			Description:   version.Description,
			CreatedAt:     frm.CreatedAt,
			UpdatedAt:     version.CreatedAt,
			Fields:        p.Fields,
		}

		return nil
	})

	return form, err
}

func CreateServiceForm(ctx context.Context, d interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}, q *db.Queries, p CreateServiceVersionParams) (form *ServiceForm, err error) {
	err = pgx.BeginFunc(ctx, d, func(tx pgx.Tx) error {
		qtx := q.WithTx(tx)
		_ = qtx

		_, err := qtx.CreateForm(ctx, db.CreateFormParams{
			CreatorID: p.CreatorID,
			Slug:      p.ServiceSlug,
		})
		if err != nil {
			return err
		}

		form, err = createServiceVersion(ctx, tx, qtx, p)
		if err != nil {
			return err
		}

		_, err = qtx.AssignCurrentFormVersion(
			ctx,
			db.AssignCurrentFormVersionParams{
				FormID:        form.FormID,
				FormVersionID: form.FormVersionID,
			},
		)
		if err != nil {
			return err
		}

		return nil
	})

	return form, err
}

func UpdateServiceForm(ctx context.Context, d interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}, q *db.Queries, p CreateServiceVersionParams) (form *ServiceForm, err error) {
	err = pgx.BeginFunc(ctx, d, func(tx pgx.Tx) error {
		qtx := q.WithTx(tx)

		form, err = createServiceVersion(ctx, tx, qtx, p)
		if err != nil {
			return err
		}

		dupes, err := qtx.FindDuplicates(ctx, form.FormVersionID)
		if err != nil {
			return err
		}
		if len(dupes) > 0 {
			return ErrUnchangedForm
		}

		_, err = qtx.AssignCurrentFormVersion(
			ctx,
			db.AssignCurrentFormVersionParams{
				FormID:        form.FormID,
				FormVersionID: form.FormVersionID,
			},
		)
		if err != nil {
			return err
		}

		return nil
	})
	return form, err
}
