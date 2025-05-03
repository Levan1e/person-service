package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"person-service/internal/models"
	"person-service/internal/service"
	"person-service/pkg/logger"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Handler предоставляет обработчики для REST API.
type Handler struct {
	service *service.PersonService
	logger  *zap.Logger
}

// NewHandler создаёт новый экземпляр Handler.
func NewHandler(service *service.PersonService, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// CreatePerson создаёт новую запись о человеке.
// @Summary Создать новую запись о человеке
// @Description Создаёт запись о человеке с указанным именем, фамилией (опционально) и отчеством (опционально). Данные обогащаются через внешние API.
// @Tags persons
// @Accept json
// @Produce json
// @Param person body models.PersonInput true "Данные человека"
// @Success 201 {object} models.Person "Созданная запись"
// @Failure 400 {object} map[string]string "Некорректный JSON или ошибка валидации"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/persons [post]
func (h *Handler) CreatePerson(w http.ResponseWriter, r *http.Request) {
	var input models.PersonInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Ошибка декодирования запроса", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный JSON"}`, http.StatusBadRequest)
		return
	}

	person, err := h.service.Create(r.Context(), &input)
	if err != nil {
		h.logger.Error("Ошибка создания записи", logger.ErrorKV("error", err))
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(person); err != nil {
		h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}

// GetPerson возвращает запись по ID.
// @Summary Получить запись о человеке по ID
// @Description Возвращает запись о человеке по указанному ID.
// @Tags persons
// @Produce json
// @Param id path int true "ID человека"
// @Success 200 {object} models.Person "Запись найдена"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 404 {object} map[string]string "Запись не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/persons/{id} [get]
func (h *Handler) GetPerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Некорректный ID", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный ID"}`, http.StatusBadRequest)
		return
	}

	person, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Ошибка получения записи", logger.ErrorKV("error", err))
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(person); err != nil {
		h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}

// UpdatePerson обновляет запись.
// @Summary Обновить запись о человеке
// @Description Обновляет существующую запись о человеке по указанному ID.
// @Tags persons
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Param person body models.Person true "Обновлённые данные человека"
// @Success 200 {object} map[string]string "Запись обновлена"
// @Failure 400 {object} map[string]string "Некорректный ID или JSON"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/persons/{id} [put]
func (h *Handler) UpdatePerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Некорректный ID", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный ID"}`, http.StatusBadRequest)
		return
	}

	var person models.Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		h.logger.Error("Ошибка декодирования запроса", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный JSON"}`, http.StatusBadRequest)
		return
	}
	person.ID = id

	if err := h.service.Update(r.Context(), &person); err != nil {
		h.logger.Error("Ошибка обновления записи", logger.ErrorKV("error", err))
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Запись обновлена"}); err != nil {
		h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}

// DeletePerson удаляет запись.
// @Summary Удалить запись о человеке
// @Description Удаляет запись о человеке по указанному ID.
// @Tags persons
// @Produce json
// @Param id path int true "ID человека"
// @Success 200 {object} map[string]string "Запись удалена"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 404 {object} map[string]string "Запись не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/persons/{id} [delete]
func (h *Handler) DeletePerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Некорректный ID", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.logger.Error("Ошибка удаления записи", logger.ErrorKV("error", err))
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Запись удалена"}); err != nil {
		h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}

// ListPersons возвращает список записей с пагинацией и фильтрами.
// @Summary Получить список записей о людях
// @Description Возвращает список записей о людях с поддержкой пагинации и фильтров по имени и фамилии.
// @Tags persons
// @Produce json
// @Param limit query int false "Лимит записей" default(10)
// @Param offset query int false "Смещение" default(0)
// @Param name query string false "Фильтр по имени"
// @Param surname query string false "Фильтр по фамилии"
// @Success 200 {array} models.Person "Список записей"
// @Failure 400 {object} map[string]string "Некорректные параметры"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/persons [get]
func (h *Handler) ListPersons(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	filters := map[string]string{}
	if name := r.URL.Query().Get("name"); name != "" {
		filters["name"] = name
	}
	if surname := r.URL.Query().Get("surname"); surname != "" {
		filters["surname"] = surname
	}

	persons, err := h.service.List(r.Context(), limit, offset, filters)
	if err != nil {
		h.logger.Error("Ошибка получения списка", logger.ErrorKV("error", err))
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if persons == nil {
		persons = []*models.Person{}
	}

	h.logger.Info("Список записей получен", logger.InfoKV("count", len(persons)))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(persons); err != nil {
		h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}
