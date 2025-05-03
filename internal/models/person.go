package models

import (
	"fmt"
	"time"
)

// Person представляет запись о человеке в базе данных.
type Person struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Surname     *string    `json:"surname" db:"surname"`
	Patronymic  *string    `json:"patronymic" db:"patronymic"`
	Age         *int       `json:"age" db:"age"`
	Gender      *string    `json:"gender" db:"gender"`
	Nationality *string    `json:"nationality" db:"nationality"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
}

// PersonInput представляет входные данные для создания человека.
type PersonInput struct {
	Name       string  `json:"name"`
	Surname    *string `json:"surname"`
	Patronymic *string `json:"patronymic"`
}

// Validate проверяет корректность данных Person.
func (p *Person) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name обязателен")
	}
	if p.Age != nil && *p.Age < 0 {
		return fmt.Errorf("age не может быть отрицательным")
	}
	return nil
}

// Validate проверяет корректность данных PersonInput.
func (pi *PersonInput) Validate() error {
	if pi.Name == "" {
		return fmt.Errorf("name обязателен")
	}
	return nil
}
