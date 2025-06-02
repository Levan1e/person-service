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

// PersonService предоставляет бизнес-логику для работы с записями о людях.
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

// fetchAge запрашивает возраст по имени через API.
func (s *PersonService) fetchAge(ctx context.Context, name string) (*int, error) {
	url := fmt.Sprintf("%s/?name=%s", s.apis.Agify, name)
	resp, err := http.Get(url)
	if err != nil {
		s.logger.Error("Ошибка запроса к Agify API", zap.Error(err))
		return nil, fmt.Errorf("не удалось получить возраст: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Неуспешный ответ от Agify API", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("неуспешный ответ от Agify API: %d", resp.StatusCode)
	}

	var result struct {
		Age *int `json:"age"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("Ошибка декодирования ответа Agify API", zap.Error(err))
		return nil, fmt.Errorf("не удалось декодировать ответ: %w", err)
	}
	return result.Age, nil
}

// fetchGender запрашивает пол по имени через API.
func (s *PersonService) fetchGender(ctx context.Context, name string) (*models.GenderType, error) {
	url := fmt.Sprintf("%s/?name=%s", s.apis.Genderize, name)
	resp, err := http.Get(url)
	if err != nil {
		s.logger.Error("Ошибка запроса к Genderize API", zap.Error(err))
		return nil, fmt.Errorf("не удалось получить пол: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Неуспешный ответ от Genderize API", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("неуспешный ответ от Genderize API: %d", resp.StatusCode)
	}

	var result struct {
		Gender *string `json:"gender"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("Ошибка декодирования ответа Genderize API", zap.Error(err))
		return nil, fmt.Errorf("не удалось декодировать ответ: %w", err)
	}
	if result.Gender == nil {
		return nil, nil
	}
	gender := models.GenderType(*result.Gender)
	if gender != models.GenderMale && gender != models.GenderFemale {
		return nil, nil
	}
	return &gender, nil
}

// fetchNationality запрашивает национальность по имени через API.
func (s *PersonService) fetchNationality(ctx context.Context, name string) (*string, error) {
	url := fmt.Sprintf("%s/?name=%s", s.apis.Nationalize, name)
	resp, err := http.Get(url)
	if err != nil {
		s.logger.Error("Ошибка запроса к Nationalize API", zap.Error(err))
		return nil, fmt.Errorf("не удалось получить национальность: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Неуспешный ответ от Nationalize API", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("неуспешный ответ от Nationalize API: %d", resp.StatusCode)
	}

	var result struct {
		Country []struct {
			CountryID string `json:"country_id"`
		} `json:"country"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("Ошибка декодирования ответа Nationalize API", zap.Error(err))
		return nil, fmt.Errorf("не удалось декодировать ответ: %w", err)
	}
	if len(result.Country) == 0 {
		return nil, nil
	}
	return &result.Country[0].CountryID, nil
}

// Create создаёт новую запись о человеке с обогащением данных.
func (s *PersonService) Create(ctx context.Context, input *models.PersonInput) (*models.Person, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("валидация входных данных: %w", err)
	}

	person := &models.Person{
		Name:       input.Name,
		Surname:    input.Surname,
		Patronymic: input.Patronymic,
		CreatedAt:  time.Now(),
	}

	// Обогащение данных
	age, err := s.fetchAge(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения возраста: %w", err)
	}
	person.Age = age

	gender, err := s.fetchGender(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пола: %w", err)
	}
	person.Gender = gender

	nationality, err := s.fetchNationality(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения национальности: %w", err)
	}
	person.Nationality = nationality

	if err := person.Validate(); err != nil {
		return nil, fmt.Errorf("валидация обогащённых данных: %w", err)
	}

	if err := s.repo.Create(ctx, person); err != nil {
		return nil, fmt.Errorf("не удалось создать запись: %w", err)
	}

	s.logger.Info("Запись создана", zap.Int("id", person.ID))
	return person, nil
}

// GetByID возвращает запись по ID.
func (s *PersonService) GetByID(ctx context.Context, id int) (*models.Person, error) {
	person, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить запись: %w", err)
	}
	s.logger.Info("Запись получена", zap.Int("id", id))
	return person, nil
}

// Update обновляет запись.
func (s *PersonService) Update(ctx context.Context, person *models.Person) error {
	if err := person.Validate(); err != nil {
		return fmt.Errorf("валидация данных: %w", err)
	}
	if err := s.repo.Update(ctx, person); err != nil {
		return fmt.Errorf("не удалось обновить запись: %w", err)
	}
	s.logger.Info("Запись обновлена", zap.Int("id", person.ID))
	return nil
}

// Patch частично обновляет запись.
func (s *PersonService) Patch(ctx context.Context, id int, update *models.PersonUpdate) error {
	// Проверяем, указано ли хотя бы одно поле для обновления
	if update.Name == nil && update.Surname == nil && update.Patronymic == nil &&
		update.Age == nil && update.Gender == nil && update.Nationality == nil {
		return fmt.Errorf("не указано ни одного поля для обновления")
	}

	// Валидация входных данных
	if err := update.Validate(); err != nil {
		return fmt.Errorf("валидация данных: %w", err)
	}

	// Получаем существующую запись
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("не удалось найти запись: %w", err)
	}

	// Проверяем, есть ли изменения
	noChanges := true
	if update.Name != nil && *update.Name != existing.Name {
		noChanges = false
	}
	if update.Surname != nil && (existing.Surname == nil || *update.Surname != *existing.Surname) {
		noChanges = false
	}
	if update.Patronymic != nil && (existing.Patronymic == nil || *update.Patronymic != *existing.Patronymic) {
		noChanges = false
	}
	if update.Age != nil && (existing.Age == nil || *update.Age != *existing.Age) {
		noChanges = false
	}
	if update.Gender != nil && (existing.Gender == nil || *update.Gender != *existing.Gender) {
		noChanges = false
	}
	if update.Nationality != nil && (existing.Nationality == nil || *update.Nationality != *existing.Nationality) {
		noChanges = false
	}

	if noChanges {
		s.logger.Info("Данные не изменены, так как они совпадают с текущими", zap.Int("id", id))
		return fmt.Errorf("данные не изменены, так как они совпадают с текущими")
	}

	// Вызываем метод Patch в репозитории
	if err := s.repo.Patch(ctx, id, update); err != nil {
		return fmt.Errorf("не удалось обновить запись: %w", err)
	}

	s.logger.Info("Запись частично обновлена", zap.Int("id", id))
	return nil
}

// Delete удаляет запись.
func (s *PersonService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("не удалось удалить запись: %w", err)
	}
	s.logger.Info("Запись удалена", zap.Int("id", id))
	return nil
}

// List возвращает список записей с пагинацией и фильтрами.
func (s *PersonService) List(ctx context.Context, limit, offset int, filters map[string]string) ([]*models.Person, error) {
	persons, err := s.repo.List(ctx, limit, offset, filters)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список: %w", err)
	}
	s.logger.Info("Список записей получен", zap.Int("count", len(persons)))
	return persons, nil
}
