package service

import (
	"context"

	"github.com/alginugraha/monify/internal/domain"
)

// UserRepo is the persistence port for users. Implementations live in package
// postgres; interfaces are defined here, on the consumer side.
type UserRepo interface {
	ByEmail(ctx context.Context, email string) (domain.User, error)
	ByID(ctx context.Context, id string) (domain.User, error)
	Create(ctx context.Context, email, passwordHash string) (domain.User, error)
}

type AccountRepo interface {
	List(ctx context.Context) ([]domain.Account, error)
	ByID(ctx context.Context, id string) (domain.Account, error)
	Create(ctx context.Context, a domain.Account) (domain.Account, error)
	Update(ctx context.Context, a domain.Account) error
	SoftDelete(ctx context.Context, id string) error
}

type CategoryRepo interface {
	List(ctx context.Context) ([]domain.Category, error)
	ByID(ctx context.Context, id string) (domain.Category, error)
	Create(ctx context.Context, c domain.Category) (domain.Category, error)
	Update(ctx context.Context, c domain.Category) error
	SoftDelete(ctx context.Context, id string) error
}
