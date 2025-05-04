-- name: CreatePerson
INSERT INTO persons (name, surname, patronymic, age, gender, nationality, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: GetPersonByID
SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at
FROM persons
WHERE id = $1;

-- name: UpdatePerson
UPDATE persons
SET name = $1, surname = $2, patronymic = $3, age = $4, gender = $5, nationality = $6, updated_at = $7
WHERE id = $8;

-- name: DeletePerson
DELETE FROM persons
WHERE id = $1;

-- name: ListPersons
SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at
FROM persons
{{if .Where}}WHERE {{.Where}}{{end}}
ORDER BY id
LIMIT $1 OFFSET $2;