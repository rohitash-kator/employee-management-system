package employee

import "context"

type ListFilter struct {
	Department *string
	Status     *Status
	Query      *string // search in name/email
}

type ListPage struct {
	Limit  int64
	Offset int64
}

type Repository interface {
	Create(ctx context.Context, e *Employee) error
	GetByID(ctx context.Context, id string) (*Employee, error)
	GetByEmail(ctx context.Context, email string) (*Employee, error)
	List(ctx context.Context, filter ListFilter, page ListPage) ([]Employee, int64, error)
	Update(ctx context.Context, e *Employee) error
	Delete(ctx context.Context, id string) error
}
