-- +goose Up
-- +goose StatementBegin
CREATE TYPE gender_type AS ENUM ('male', 'female');

CREATE TABLE IF NOT EXISTS persons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    surname VARCHAR(255),
    patronymic VARCHAR(255),
    age INT CHECK (age >= 0),
    gender gender_type,
    nationality VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_persons_name ON persons (name);
CREATE INDEX idx_persons_surname ON persons (surname);
CREATE INDEX idx_persons_nationality ON persons (nationality);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS persons;
DROP TYPE IF EXISTS gender_type;
-- +goose StatementEnd