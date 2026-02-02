package employee

import (
	"context"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/rohitashk/golang-rest-api/internal/domain"
	domainEmployee "github.com/rohitashk/golang-rest-api/internal/domain/employee"
)

type CreateInput struct {
	FirstName  string  `validate:"required,min=1,max=100"`
	LastName   string  `validate:"required,min=1,max=100"`
	Email      string  `validate:"required,email,max=320"`
	Department string  `validate:"required,min=1,max=120"`
	Position   string  `validate:"required,min=1,max=120"`
	Salary     float64 `validate:"gte=0,lte=1000000000"`
	Status     string  `validate:"omitempty,oneof=active inactive"`
}

type UpdateInput struct {
	FirstName  *string  `validate:"omitempty,min=1,max=100"`
	LastName   *string  `validate:"omitempty,min=1,max=100"`
	Email      *string  `validate:"omitempty,email,max=320"`
	Department *string  `validate:"omitempty,min=1,max=120"`
	Position   *string  `validate:"omitempty,min=1,max=120"`
	Salary     *float64 `validate:"omitempty,gte=0,lte=1000000000"`
	Status     *string  `validate:"omitempty,oneof=active inactive"`
}

type ListInput struct {
	Department *string
	Status     *string
	Query      *string
	Limit      int64
	Offset     int64
}

type Service struct {
	repo     domainEmployee.Repository
	validate *validator.Validate
	now      func() time.Time
}

func NewService(repo domainEmployee.Repository) *Service {
	return &Service{
		repo:     repo,
		validate: validator.New(),
		now:      time.Now,
	}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*domainEmployee.Employee, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if err := s.validate.Struct(in); err != nil {
		return nil, domain.Validation(err.Error())
	}

	existing, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.Conflict("employee with this email already exists")
	}

	status := domainEmployee.StatusActive
	if in.Status != "" {
		status = domainEmployee.Status(in.Status)
	}

	now := s.now().UTC()
	e := &domainEmployee.Employee{
		FirstName:  in.FirstName,
		LastName:   in.LastName,
		Email:      in.Email,
		Department: in.Department,
		Position:   in.Position,
		Salary:     in.Salary,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domainEmployee.Employee, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, domain.NotFound("employee not found")
	}
	return e, nil
}

func (s *Service) List(ctx context.Context, in ListInput) ([]domainEmployee.Employee, int64, error) {
	if in.Limit <= 0 || in.Limit > 200 {
		in.Limit = 20
	}
	if in.Offset < 0 {
		in.Offset = 0
	}

	var status *domainEmployee.Status
	if in.Status != nil && *in.Status != "" {
		st := domainEmployee.Status(strings.ToLower(strings.TrimSpace(*in.Status)))
		status = &st
	}

	filter := domainEmployee.ListFilter{
		Department: in.Department,
		Status:     status,
		Query:      in.Query,
	}
	page := domainEmployee.ListPage{Limit: in.Limit, Offset: in.Offset}
	return s.repo.List(ctx, filter, page)
}

func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (*domainEmployee.Employee, error) {
	if err := s.validate.Struct(in); err != nil {
		return nil, domain.Validation(err.Error())
	}

	e, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Email != nil {
		v := strings.TrimSpace(strings.ToLower(*in.Email))
		if v != "" && v != e.Email {
			existing, err := s.repo.GetByEmail(ctx, v)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != e.ID {
				return nil, domain.Conflict("employee with this email already exists")
			}
			e.Email = v
		}
	}
	if in.FirstName != nil {
		e.FirstName = *in.FirstName
	}
	if in.LastName != nil {
		e.LastName = *in.LastName
	}
	if in.Department != nil {
		e.Department = *in.Department
	}
	if in.Position != nil {
		e.Position = *in.Position
	}
	if in.Salary != nil {
		e.Salary = *in.Salary
	}
	if in.Status != nil && *in.Status != "" {
		e.Status = domainEmployee.Status(*in.Status)
	}

	e.UpdatedAt = s.now().UTC()
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	// ensure not-found is consistent
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
