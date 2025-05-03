package postgres

import (
	"context"
	"fmt"
	"person-service/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PersonRepository предоставляет методы для работы с таблицей persons.
type PersonRepository struct {
	db *pgxpool.Pool
}

// NewPersonRepository создаёт новый репозиторий для работы с persons.
func NewPersonRepository(db *pgxpool.Pool) *PersonRepository {
	return &PersonRepository{db: db}
}

// Create создаёт новую запись в таблице persons.
func (r *PersonRepository) Create(ctx context.Context, person *models.Person) error {
	query := `
		INSERT INTO persons (name, surname, patronymic, age, gender, nationality, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		person.CreatedAt,
		person.UpdatedAt,
	).Scan(&person.ID)
	if err != nil {
		return fmt.Errorf("не удалось создать запись: %w", err)
	}
	return nil
}

// GetByID возвращает запись по ID.
func (r *PersonRepository) GetByID(ctx context.Context, id int) (*models.Person, error) {
	query := `
		SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at
		FROM persons
		WHERE id = $1
	`
	var person models.Person
	err := r.db.QueryRow(ctx, query, id).Scan(
		&person.ID,
		&person.Name,
		&person.Surname,
		&person.Patronymic,
		&person.Age,
		&person.Gender,
		&person.Nationality,
		&person.CreatedAt,
		&person.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить запись: %w", err)
	}
	return &person, nil
}

// Update обновляет запись в таблице persons.
func (r *PersonRepository) Update(ctx context.Context, person *models.Person) error {
	query := `
		UPDATE persons
		SET name = $1, surname = $2, patronymic = $3, age = $4, gender = $5, nationality = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.db.Exec(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		time.Now(),
		person.ID,
	)
	if err != nil {
		return fmt.Errorf("не удалось обновить запись: %w", err)
	}
	return nil
}

// Delete удаляет запись по ID.
func (r *PersonRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM persons WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить запись: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("запись с id %d не найдена", id)
	}
	return nil
}

// List возвращает список записей с пагинацией и фильтрами.
func (r *PersonRepository) List(ctx context.Context, limit, offset int, filters map[string]string) ([]*models.Person, error) {
	query := `SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at FROM persons`
	args := []interface{}{}
	whereClauses := []string{}
	argIndex := 1

	if name, ok := filters["name"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+name+"%")
		argIndex++
	}
	if surname, ok := filters["surname"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("surname ILIKE $%d", argIndex))
		args = append(args, "%"+surname+"%")
		argIndex++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}

	query += fmt.Sprintf(" ORDER BY id LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список записей: %w", err)
	}
	defer rows.Close()

	var persons []*models.Person
	for rows.Next() {
		var person models.Person
		if err := rows.Scan(
			&person.ID,
			&person.Name,
			&person.Surname,
			&person.Patronymic,
			&person.Age,
			&person.Gender,
			&person.Nationality,
			&person.CreatedAt,
			&person.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("не удалось отсканировать запись: %w", err)
		}
		persons = append(persons, &person)
	}

	return persons, nil
}
