package repos

import (
	"context"
	"errors"
	"fmt"

	"xm-companies-manager/internal/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound = errors.New("company not found")
	ErrConflict = errors.New("company already exists")
)

type CompanyRepository struct {
	queries *sqlc.Queries
}

func NewCompanyRepository(queries *sqlc.Queries) *CompanyRepository {
	return &CompanyRepository{
		queries: queries,
	}
}

func (r *CompanyRepository) Get(
	ctx context.Context,
	id uuid.UUID,
) (sqlc.Company, error) {
	company, err := r.queries.GetCompany(ctx, toPGUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Company{}, ErrNotFound
		}

		return sqlc.Company{}, fmt.Errorf("get company: %w", err)
	}

	return company, nil
}

type CreateCompanyParams struct {
	ID                uuid.UUID
	Name              string
	Description       string
	AmountOfEmployees int32
	Registered        bool
	Type              string
}

func (r *CompanyRepository) Create(
	ctx context.Context,
	params CreateCompanyParams,
) (sqlc.Company, error) {
	company, err := r.queries.CreateCompany(
		ctx,
		sqlc.CreateCompanyParams{
			ID:   toPGUUID(params.ID),
			Name: params.Name,

			Description: pgtype.Text{
				String: params.Description,
				Valid:  params.Description != "",
			},

			AmountOfEmployees: params.AmountOfEmployees,
			Registered:        params.Registered,
			Type:              params.Type,
		},
	)

	if err != nil {
		if isUniqueViolation(err) {
			return sqlc.Company{}, ErrConflict
		}

		return sqlc.Company{}, fmt.Errorf("create company: %w", err)
	}

	return company, nil
}

type PatchCompanyParams struct {
	ID                uuid.UUID
	Name              *string
	Description       *string
	AmountOfEmployees *int32
	Registered        *bool
	Type              *string
}

func (r *CompanyRepository) Patch(
	ctx context.Context,
	params PatchCompanyParams,
) (sqlc.Company, error) {
	company, err := r.queries.PatchCompany(
		ctx,
		sqlc.PatchCompanyParams{
			ID: toPGUUID(params.ID),

			Name: pgtype.Text{
				String: valueOrZero(params.Name),
				Valid:  params.Name != nil,
			},

			Description: pgtype.Text{
				String: valueOrZero(params.Description),
				Valid:  params.Description != nil,
			},

			AmountOfEmployees: pgtype.Int4{
				Int32: valueOrZero(params.AmountOfEmployees),
				Valid: params.AmountOfEmployees != nil,
			},

			Registered: pgtype.Bool{
				Bool:  valueOrZero(params.Registered),
				Valid: params.Registered != nil,
			},

			Type: pgtype.Text{
				String: valueOrZero(params.Type),
				Valid:  params.Type != nil,
			},
		},
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Company{}, ErrNotFound
		}

		// PATCHing the name can also violate UNIQUE(name).
		if isUniqueViolation(err) {
			return sqlc.Company{}, ErrConflict
		}

		return sqlc.Company{}, fmt.Errorf("patch company: %w", err)
	}

	return company, nil
}

func (r *CompanyRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	rowsAffected, err := r.queries.DeleteCompany(
		ctx,
		toPGUUID(id),
	)

	if err != nil {
		return fmt.Errorf("delete company: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toPGUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

func valueOrZero[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}

	return *value
}
