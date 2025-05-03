package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"person-service/internal/config"
	"person-service/internal/models"
	"person-service/internal/repository"
	"time"

	"go.uber.org/zap"
)

// PersonService управляет бизнес-логикой для работы с записями о людях.
type PersonService struct {
	repo   repository.PersonRepository
	logger *zap.Logger
	apis   *config.APIs
}

// NewPersonService создаёт новый экземпляр PersonService.
func NewPersonService(repo repository.PersonRepository, logger *zap.Logger, apis *config.APIs) *PersonService {
	return &PersonService{
		repo:   repo,
		logger: logger,
		apis:   apis,
	}
}

// Create создаёт новую запись о человеке с обогащением данных.
func (s *PersonService) Create(ctx context.Context, input *models.PersonInput) (*models.Person, error) {
	// Валидация
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("валидация не пройдена: %w", err)
	}

	// Обогащение данных
	age, err := s.fetchAge(input.Name)
	if err != nil {
		s.logger.Warn("Ошибка получения возраста", zap.Error(err))
		age = 0
	}

	gender, err := s.fetchGender(input.Name)
	if err != nil {
		s.logger.Warn("Ошибка получения пола", zap.Error(err))
		gender = "unknown"
	}

	nationality, err := s.fetchNationality(input.Name)
	if err != nil {
		s.logger.Warn("Ошибка получения национальности", zap.Error(err))
		nationality = "unknown"
	}

	// Создание записи
	person := &models.Person{
		Name:        input.Name,
		Surname:     input.Surname,
		Patronymic:  input.Patronymic,
		Age:         &age,
		Gender:      &gender,
		Nationality: &nationality,
		CreatedAt:   time.Now(),
	}

	err = s.repo.Create(ctx, person)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания записи: %w", err)
	}

	return person, nil
}

// GetByID возвращает запись по ID.
func (s *PersonService) GetByID(ctx context.Context, id int) (*models.Person, error) {
	person, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("запись не найдена: %w", err)
	}
	return person, nil
}

// Update обновляет запись.
func (s *PersonService) Update(ctx context.Context, person *models.Person) error {
	if err := person.Validate(); err != nil {
		return fmt.Errorf("валидация не пройдена: %w", err)
	}

	person.UpdatedAt = &time.Time{}
	*person.UpdatedAt = time.Now()

	return s.repo.Update(ctx, person)
}

// Delete удаляет запись.
func (s *PersonService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// List возвращает список записей с пагинацией и фильтрами.
func (s *PersonService) List(ctx context.Context, limit, offset int, filters map[string]string) ([]*models.Person, error) {
	persons, err := s.repo.List(ctx, limit, offset, filters)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка: %w", err)
	}
	if persons == nil {
		return []*models.Person{}, nil
	}
	return persons, nil
}

// fetchAge получает возраст из API Agify.
func (s *PersonService) fetchAge(name string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s?name=%s", s.apis.Agify, name))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Age int `json:"age"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Age, nil
}

// fetchGender получает пол из API Genderize.
func (s *PersonService) fetchGender(name string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s?name=%s", s.apis.Genderize, name))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Gender string `json:"gender"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Gender, nil
}

// fetchNationality получает национальность из API Nationalize.
func (s *PersonService) fetchNationality(name string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s?name=%s", s.apis.Nationalize, name))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Country) > 0 {
		return result.Country[0].CountryID, nil
	}
	return "", fmt.Errorf("национальность не определена")
}
