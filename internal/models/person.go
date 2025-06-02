package models

import (
	"fmt"
	"time"
)

// GenderType представляет допустимые значения пола.
type GenderType string

const (
	GenderMale   GenderType = "male"
	GenderFemale GenderType = "female"
)

// Person представляет запись о человеке в базе данных.
type Person struct {
	ID          int         `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Surname     *string     `json:"surname" db:"surname"`
	Patronymic  *string     `json:"patronymic" db:"patronymic"`
	Age         *int        `json:"age" db:"age"`
	Gender      *GenderType `json:"gender" db:"gender"`
	Nationality *string     `json:"nationality" db:"nationality"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time  `json:"updated_at" db:"updated_at"`
}

// PersonInput представляет входные данные для создания человека.
type PersonInput struct {
	Name       string  `json:"name"`
	Surname    *string `json:"surname"`
	Patronymic *string `json:"patronymic"`
}

// PersonUpdate представляет входные данные для обновления человека.
type PersonUpdate struct {
	Name        *string     `json:"name"`
	Surname     *string     `json:"surname"`
	Patronymic  *string     `json:"patronymic"`
	Age         *int        `json:"age"`
	Gender      *GenderType `json:"gender"`
	Nationality *string     `json:"nationality"`
}

// Validate проверяет корректность данных Person.
func (p *Person) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name обязателен")
	}
	if p.Age != nil && *p.Age < 0 {
		return fmt.Errorf("age не может быть отрицательным")
	}
	if p.Gender != nil && *p.Gender != GenderMale && *p.Gender != GenderFemale {
		return fmt.Errorf("gender должен быть 'male' или 'female'")
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

func (pu *PersonUpdate) Validate() error {
	if pu.Name != nil && *pu.Name == "" {
		return fmt.Errorf("name не может быть пустым")
	}
	if pu.Age != nil && *pu.Age < 0 {
		return fmt.Errorf("age не может быть отрицательным")
	}
	if pu.Gender != nil {
		gender := *pu.Gender
		if gender != GenderMale && gender != GenderFemale {
			return fmt.Errorf("gender должен быть 'male' или 'female'")
		}
	}
	return nil
}
