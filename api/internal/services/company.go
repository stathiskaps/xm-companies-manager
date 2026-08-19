package services

import (
	"context"
	"fmt"

	"xm-companies-manager/internal/database/sqlc"
	"xm-companies-manager/internal/repos"

	"github.com/google/uuid"
)

const (
	EventCompanyCreated = "company.created"
	EventCompanyUpdated = "company.updated"
	EventCompanyDeleted = "company.deleted"
)

type EventProducer interface {
	PublishCompanyEvent(
		ctx context.Context,
		eventType string,
		companyID uuid.UUID,
	) error
}

type CompanyService struct {
	repo     *repos.CompanyRepository
	producer EventProducer
}

func NewCompanyService(repo *repos.CompanyRepository, producer EventProducer) *CompanyService {
	return &CompanyService{
		repo:     repo,
		producer: producer,
	}
}

func (s *CompanyService) Get(ctx context.Context, id uuid.UUID) (sqlc.Company, error) {
	return s.repo.Get(ctx, id)
}

func (s *CompanyService) Create(ctx context.Context, params repos.CreateCompanyParams) (sqlc.Company, error) {
	company, err := s.repo.Create(ctx, params)
	if err != nil {
		return sqlc.Company{}, err
	}

	if err := s.producer.PublishCompanyEvent(
		ctx,
		EventCompanyCreated,
		params.ID,
	); err != nil {
		return sqlc.Company{}, fmt.Errorf(
			"publish company created event: %w",
			err,
		)
	}

	return company, nil
}

func (s *CompanyService) Patch(ctx context.Context, params repos.PatchCompanyParams) (sqlc.Company, error) {
	company, err := s.repo.Patch(ctx, params)
	if err != nil {
		return sqlc.Company{}, err
	}

	if err := s.producer.PublishCompanyEvent(
		ctx,
		EventCompanyUpdated,
		params.ID,
	); err != nil {
		return sqlc.Company{}, fmt.Errorf(
			"publish company updated event: %w",
			err,
		)
	}

	return company, nil
}

func (s *CompanyService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.producer.PublishCompanyEvent(
		ctx,
		EventCompanyDeleted,
		id,
	); err != nil {
		return fmt.Errorf(
			"publish company deleted event: %w",
			err,
		)
	}

	return nil
}
