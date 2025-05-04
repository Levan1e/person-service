# Person Service

It's a service you can use for links shortening, that is written in Go.

## Project description:

Person Service is a service for creating, retrieving, updating, deleting, and listing person records with data enrichment from external APIs (Agify.io, Genderize.io, Nationalize.io).

### Features:
+ Create person records with name, surname, and optional patronymic.
+ Enrich records with age, gender, and nationality.
+ Retrieve, update, or delete records by ID.
+ List records with pagination.
+ Swagger UI for API documentation.
+ Graceful shutdown with 10-second timeout.
+ PostgreSQL storage with migration support.
+ Docker and Docker Compose support.

### Technology Stack:
+ Language: Go 1.22
+ Framework: Chi (github.com/go-chi/chi/v5)
+ Database: PostgreSQL 17
+ Logging: Zap (go.uber.org/zap)
+ API Docs: Swagger (github.com/swaggo/http-swagger)
+ Migrations: Goose (github.com/pressly/goose/v3)
+ Containerization: Docker

## Launch options: 
        go run .\cmd\person-service\main.go

## Docker launch options: 
      docker build -t person-service .
      docker run -p 8081:8081 --env-file .env person-service
## Docker Compose (recommended):
      docker-compose up -d

## API endpoints:
+ ### POST /api/v1/persons
  Create a person record.
  
  ### Request:
      { "name": "Иван", "surname": "Иванов", "patronymic": "Иванович" }

  ### Response:
      { "id": 1, "name": "Иван", "surname": "Иванов", "patronymic": "Иванович", "age": 30, "gender": "male", "nationality": "RU", "created_at": "2025-05-04T12:00:00Z", "updated_at": "2025-05-04T12:00:00Z" }

+ ### GET /api/v1/persons
  List persons with pagination/filters.
  
  ### Query Parameters: limit (default: 10), offset (default: 0), name, surname

  ### Response:
      [{ "id": 1, "name": "Иван", "surname": "Иванов", "patronymic": "Иванович", "age": 30, "gender": "male", "nationality": "RU", "created_at": "2025-05-04T12:00:00Z", "updated_at": "2025-05-04T12:00:00Z" }]

+ ### GET /api/v1/persons/{id}
  Get a person by ID.

  ### Response:
      { "id": 1, "name": "Иван", "surname": "Иванов", "patronymic": "Иванович", "age": 30, "gender": "male", "nationality": "RU", "created_at": "2025-05-04T12:00:00Z", "updated_at": "2025-05-04T12:00:00Z" }

  + ### PUT /api/v1/persons/{id}
  Update a person record.
  
  ### Request:
      { "name": "Иван", "surname": "Петров", "patronymic": "Иванович", "age": 31, "gender": "male", "nationality": "RU" }

  ### Response:
      { "message": "Запись обновлена" }

+ ### DELETE /api/v1/persons/{id}
  Delete a person record.

  ### Response:
      { "message": "Запись удалена" }
+ ### GET /swagger/index.html
  Swagger UI for API documentation.
