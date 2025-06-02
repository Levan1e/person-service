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
// @Description Создаёт запись о человеке с указанным именем, фамилией (опционально) и отчеством (опционально). Данные обогащаются через внешние API (Agify.io, Genderize.io, Nationalize.io).
// @Tags persons
// @Accept json
// @Produce json
// @Param person body models.PersonInput true "Данные человека"
// @Success 201 {object} models.Person "Созданная запись"
// @Failure 400 {object} map[string]string "Некорректный JSON или ошибка валидации, например, пустое имя"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера, например, сбой API обогащения"
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
// @Failure 400 {object} map[string]string "Некорректный ID, например, не число"
// @Failure 404 {object} map[string]string "Запись не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/persons/{id} [get]
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

// UpdatePerson обновляет запись полностью.
// @Summary Обновить запись о человеке
// @Description Полностью обновляет существующую запись о человеке по указанному ID.
// @Tags persons
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Param person body models.Person true "Обновлённые данные человека"
// @Success 200 {object} map[string]string "Запись обновлена"
// @Failure 400 {object} map[string]string "Некорректный ID, JSON или ошибка валидации"
// @Failure 404 {object} map[string]string "Запись не найдена"
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
		if err.Error() == fmt.Sprintf("не удалось найти запись: запись с id %d не найдена", id) {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Запись обновлена"}); err != nil {
		h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}

// PatchPerson частично обновляет запись.
// @Summary Частично обновить запись о человеке
// @Description Частично обновляет существующую запись о человеке по указанному ID. Возвращает сообщение об успехе, совпадении данных или ошибке.
// @Tags persons
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Param person body models.PersonUpdate true "Обновлённые данные человека"
// @Success 200 {object} map[string]string "Запись обновлена или данные не изменены"
// @Failure 400 {object} map[string]string "Некорректный ID, JSON, пустой запрос, несуществующее поле или ошибка валидации"
// @Failure 404 {object} map[string]string "Запись не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/persons/{id} [patch]
func (h *Handler) PatchPerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Некорректный ID", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный ID"}`, http.StatusBadRequest)
		return
	}

	// Читаем JSON как RawMessage для проверки полей
	var rawBody json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
		h.logger.Error("Ошибка декодирования запроса", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверяем, пустой ли запрос
	if string(rawBody) == "{}" {
		h.logger.Error("Пустой JSON в запросе")
		http.Error(w, `{"error": "Не указано ни одного поля для обновления"}`, http.StatusBadRequest)
		return
	}

	// Проверяем наличие несуществующих полей
	var tempMap map[string]interface{}
	if err := json.Unmarshal(rawBody, &tempMap); err != nil {
		h.logger.Error("Ошибка разбора JSON", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный JSON"}`, http.StatusBadRequest)
		return
	}

	validFields := map[string]bool{
		"name":        true,
		"surname":     true,
		"patronymic":  true,
		"age":         true,
		"gender":      true,
		"nationality": true,
	}
	for key := range tempMap {
		if !validFields[key] {
			h.logger.Error("Указано несуществующее поле", logger.ErrorKV("field", key))
			http.Error(w, fmt.Sprintf(`{"error": "Указано несуществующее поле: %s"}`, key), http.StatusBadRequest)
			return
		}
	}

	// Декодируем в PersonUpdate
	var update models.PersonUpdate
	if err := json.Unmarshal(rawBody, &update); err != nil {
		h.logger.Error("Ошибка декодирования в PersonUpdate", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Некорректный JSON"}`, http.StatusBadRequest)
		return
	}

	err = h.service.Patch(r.Context(), id, &update)
	if err != nil {
		h.logger.Error("Ошибка частичного обновления записи", logger.ErrorKV("error", err))
		if err.Error() == fmt.Sprintf("не удалось найти запись: запись с id %d не найдена", id) {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
			return
		}
		if err.Error() == "данные не изменены, так как они совпадают с текущими" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]string{"message": "Данные не изменены, так как они совпадают с текущими"}); err != nil {
				h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
				http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
			}
			return
		}
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
// @Failure 400 {object} map[string]string "Некорректный ID, например, не число"
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
// @Description Возвращает список записей о людях с поддержкой пагинации и фильтров по всем полям. Если записей нет, возвращается сообщение.
// @Tags persons
// @Produce json
// @Param limit query int false "Лимит записей" default(10)
// @Param offset query int false "Смещение" default(0)
// @Param name query string false "Фильтр по имени"
// @Param surname query string false "Фильтр по фамилии"
// @Param patronymic query string false "Фильтр по отчеству"
// @Param age query int false "Фильтр по возрасту"
// @Param gender query string false "Фильтр по полу (male, female)"
// @Param nationality query string false "Фильтр по национальности"
// @Success 200 {array} models.Person "Список записей или сообщение о пустом списке"
// @Failure 400 {object} map[string]string "Некорректные параметры, например, отрицательный лимит"
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
	if patronymic := r.URL.Query().Get("patronymic"); patronymic != "" {
		filters["patronymic"] = patronymic
	}
	if age := r.URL.Query().Get("age"); age != "" {
		filters["age"] = age
	}
	if gender := r.URL.Query().Get("gender"); gender != "" {
		filters["gender"] = gender
	}
	if nationality := r.URL.Query().Get("nationality"); nationality != "" {
		filters["nationality"] = nationality
	}

	persons, err := h.service.List(r.Context(), limit, offset, filters)
	if err != nil {
		h.logger.Error("Ошибка получения списка", logger.ErrorKV("error", err))
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Проверяем, пустой ли список
	if len(persons) == 0 {
		h.logger.Info("Список записей пуст")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "В базе данных нет записей"}); err != nil {
			h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
			http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("Список записей получен", logger.InfoKV("count", len(persons)))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(persons); err != nil {
		h.logger.Error("Ошибка кодирования ответа", logger.ErrorKV("error", err))
		http.Error(w, `{"error": "Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
	}
}
