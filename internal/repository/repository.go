package repository

import (
	"context"
	"person-service/internal/models"
)

// PersonRepository определяет методы для работы с записями о людях.
type PersonRepository interface {
	Create(ctx context.Context, person *models.Person) error
	GetByID(ctx context.Context, id int) (*models.Person, error)
	Update(ctx context.Context, person *models.Person) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int, filters map[string]string) ([]*models.Person, error)
}
