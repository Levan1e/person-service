package postgres

import (
	"context"
	"embed"
	"fmt"
	"person-service/internal/models"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed queries.sql
var queriesFS embed.FS

// PersonRepository предоставляет методы для работы с таблицей persons.
type PersonRepository struct {
	db      *pgxpool.Pool
	queries map[string]string
}

// NewPersonRepository создаёт новый репозиторий для работы с persons.
func NewPersonRepository(db *pgxpool.Pool) (*PersonRepository, error) {
	// Загрузка SQL-запросов
	content, err := queriesFS.ReadFile("queries.sql")
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать queries.sql: %w", err)
	}

	queries := make(map[string]string)
	lines := strings.Split(string(content), "\n")
	var currentQuery string
	var currentName string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "-- name: ") {
			if currentName != "" && currentQuery != "" {
				queries[currentName] = strings.TrimSpace(currentQuery)
			}
			currentName = strings.TrimPrefix(line, "-- name: ")
			currentQuery = ""
		} else {
			currentQuery += line + "\n"
		}
	}
	if currentName != "" && currentQuery != "" {
		queries[currentName] = strings.TrimSpace(currentQuery)
	}

	return &PersonRepository{db: db, queries: queries}, nil
}

// Create создаёт новую запись в таблице persons.
func (r *PersonRepository) Create(ctx context.Context, person *models.Person) error {
	query := r.queries["CreatePerson"]
	err := r.db.QueryRow(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
		person.CreatedAt,
	).Scan(&person.ID)
	if err != nil {
		return fmt.Errorf("не удалось создать запись: %w", err)
	}
	return nil
}

// GetByID возвращает запись по ID.
func (r *PersonRepository) GetByID(ctx context.Context, id int) (*models.Person, error) {
	query := r.queries["GetPersonByID"]
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
		return nil, fmt.Errorf("запись с id %d не найдена", id)
	}
	return &person, nil
}

// Update обновляет запись в таблице persons.
func (r *PersonRepository) Update(ctx context.Context, person *models.Person) error {
	query := r.queries["UpdatePerson"]
	result, err := r.db.Exec(ctx, query,
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
	if result.RowsAffected() == 0 {
		return fmt.Errorf("запись с id %d не найдена", person.ID)
	}
	return nil
}

// Delete удаляет запись по ID.
func (r *PersonRepository) Delete(ctx context.Context, id int) error {
	query := r.queries["DeletePerson"]
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
	queryTemplate := r.queries["ListPersons"]
	args := []interface{}{limit, offset}
	whereClauses := []string{}

	if name, ok := filters["name"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE '%s'", "%"+name+"%"))
	}
	if surname, ok := filters["surname"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("surname ILIKE '%s'", "%"+surname+"%"))
	}
	if patronymic, ok := filters["patronymic"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("patronymic ILIKE '%s'", "%"+patronymic+"%"))
	}
	if age, ok := filters["age"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("age = %s", age))
	}
	if gender, ok := filters["gender"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("gender = '%s'", gender))
	}
	if nationality, ok := filters["nationality"]; ok {
		whereClauses = append(whereClauses, fmt.Sprintf("nationality ILIKE '%s'", "%"+nationality+"%"))
	}

	var where string
	if len(whereClauses) > 0 {
		where = strings.Join(whereClauses, " AND ")
	}

	query := strings.Replace(queryTemplate, "{{if .Where}}WHERE {{.Where}}{{end}}", where, 1)
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

func (r *PersonRepository) Patch(ctx context.Context, id int, update *models.PersonUpdate) error {
	var fields []string
	var args []interface{}
	argIndex := 1

	// Формируем список полей для обновления
	if update.Name != nil {
		fields = append(fields, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *update.Name)
		argIndex++
	}
	if update.Surname != nil {
		fields = append(fields, fmt.Sprintf("surname = $%d", argIndex))
		args = append(args, *update.Surname)
		argIndex++
	}
	if update.Patronymic != nil {
		fields = append(fields, fmt.Sprintf("patronymic = $%d", argIndex))
		args = append(args, *update.Patronymic)
		argIndex++
	}
	if update.Age != nil {
		fields = append(fields, fmt.Sprintf("age = $%d", argIndex))
		args = append(args, *update.Age)
		argIndex++
	}
	if update.Gender != nil {
		fields = append(fields, fmt.Sprintf("gender = $%d", argIndex))
		args = append(args, *update.Gender)
		argIndex++
	}
	if update.Nationality != nil {
		fields = append(fields, fmt.Sprintf("nationality = $%d", argIndex))
		args = append(args, *update.Nationality)
		argIndex++
	}

	// Добавляем updated_at
	fields = append(fields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Добавляем ID
	args = append(args, id)

	// Формируем SQL-запрос
	query := fmt.Sprintf(`
        UPDATE persons
        SET %s
        WHERE id = $%d
        RETURNING id
    `, strings.Join(fields, ", "), argIndex)

	// Выполняем запрос
	var updatedID int
	err := r.db.QueryRow(ctx, query, args...).Scan(&updatedID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return fmt.Errorf("запись с id %d не найдена", id)
		}
		return fmt.Errorf("не удалось обновить запись: %w", err)
	}

	return nil
}
