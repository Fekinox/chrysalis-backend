package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/Fekinox/chrysalis-backend/internal/db"
	"github.com/Fekinox/chrysalis-backend/internal/formfield"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserNotFound        = errors.New("User not found")
	ErrServiceNotFound     = errors.New("Service not found")
	ErrFormVersionNotFound = errors.New("Form version not found")
	ErrFieldsNotFound      = errors.New("Fields not found")
	ErrUnchangedForm       = errors.New("Form is unchanged")
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

func GetServiceFormVersion(
	ctx context.Context,
	d *db.Store,
	formVersionId int64,
) (form *ServiceForm, err error) {
	err = d.BeginFunc(ctx, func(stx *db.Store) error {
		service, err := stx.GetFormVersionById(ctx, formVersionId)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrServiceNotFound, formVersionId)
		}

		rawFields, err := stx.GetFormFields(
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

func GetServiceForm(
	ctx context.Context,
	d *db.Store,
	p ServiceFormParams,
) (form *ServiceForm, err error) {
	err = d.BeginFunc(ctx, func(stx *db.Store) error {
		user, err := stx.GetUserByUsername(ctx, p.Username)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrUserNotFound, p.Username)
		}

		params := db.GetCurrentFormVersionBySlugParams{
			Slug:      p.Service,
			CreatorID: user.ID,
		}

		service, err := stx.GetCurrentFormVersionBySlug(ctx, params)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrServiceNotFound, p.Service)
		}

		rawFields, err := stx.GetFormFields(
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

func createServiceVersion(
	ctx context.Context,
	d *db.Store,
	p CreateServiceVersionParams,
) (form *ServiceForm, err error) {
	err = d.BeginFunc(ctx, func(stx *db.Store) error {
		frm, err := stx.GetFormHeaderBySlug(ctx, db.GetFormHeaderBySlugParams{
			CreatorID: p.CreatorID,
			Slug:      p.ServiceSlug,
		})
		if err != nil {
			return err
		}

		version, err := stx.CreateFormVersion(ctx, db.CreateFormVersionParams{
			FormID:      frm.ID,
			Name:        p.Title,
			Description: p.Description,
		})
		if err != nil {
			return err
		}

		for i, f := range p.Fields {
			_, err := stx.AddFormFieldToForm(
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

			err = f.Data.Create(ctx, stx, version.ID, int64(i))
			if err != nil {
				return err
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

func CreateServiceForm(
	ctx context.Context,
	d *db.Store,
	p CreateServiceVersionParams,
) (form *ServiceForm, err error) {
	err = d.BeginFunc(ctx, func(stx *db.Store) error {
		_, err := stx.CreateForm(ctx, db.CreateFormParams{
			CreatorID: p.CreatorID,
			Slug:      p.ServiceSlug,
		})
		if err != nil {
			return err
		}

		form, err = createServiceVersion(ctx, stx, p)
		if err != nil {
			return err
		}

		_, err = stx.AssignCurrentFormVersion(
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

func UpdateServiceForm(
	ctx context.Context,
	d *db.Store,
	p CreateServiceVersionParams,
) (form *ServiceForm, err error) {
	err = d.BeginFunc(ctx, func(stx *db.Store) error {
		form, err = createServiceVersion(ctx, stx, p)
		if err != nil {
			return err
		}

		found, err := stx.FindIfFormUnchanged(ctx, form.FormVersionID)
		if err != nil {
			return err
		}
		if len(found) > 0 {
			return ErrUnchangedForm
		}

		_, err = stx.AssignCurrentFormVersion(
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
