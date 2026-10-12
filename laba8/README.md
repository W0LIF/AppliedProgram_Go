# Lab 8 — REST API пользователей

Учебный REST-сервис на Go: CRUD для пользователей, хранение в PostgreSQL
(или в памяти), валидация, централизованная обработка ошибок, пагинация,
фильтрация и покрытие тестами.

## Технологии

- Go 1.25+ (`net/http` с новым `ServeMux` и паттернами путей)
- PostgreSQL 18+ (через `github.com/lib/pq`)
- Docker / docker compose
- Стандартная библиотека `testing` + `net/http/httptest`

## Запуск

### 1. Поднять PostgreSQL

```bash
docker compose up -d
```

Проверить контейнер:

```bash
docker ps
```

### 2. Запустить сервер

Хранилище в памяти (быстро, данные теряются при перезапуске):

```bash
go run .
```

Хранилище в PostgreSQL:

```bash
STORAGE=postgres \
DATABASE_URL='postgres://lab:lab@localhost:5434/lab8?sslmode=disable' \
go run .
```

По умолчанию сервер слушает `:8080`.

### 3. Тесты

```bash
go test ./...
```

## Формат ответов

Успешный ответ — JSON-объект или массив. Ошибки — всегда:

```json
{ "error": "human readable message" }
```

Ошибки валидации дополнительно содержат `details` по каждому полю:

```json
{
  "error": "validation failed",
  "details": {
    "name":  "name is required",
    "email": "email is invalid",
    "age":   "age must be between 1 and 150"
  }
}
```

## Модель User

| Поле  | Тип    | Ограничения                       |
|-------|--------|-----------------------------------|
| id    | int    | генерируется сервером             |
| name  | string | 2–100 символов, обязательно       |
| email | string | корректный email, обязательно     |
| age   | int    | 1–150                             |

## Маршруты

### `GET /users`

Список пользователей. Поддерживает пагинацию и фильтры.

**Query-параметры**

| Имя        | Тип    | По умолчанию | Описание                                  |
|------------|--------|--------------|-------------------------------------------|
| `name`     | string | —            | Поиск по подстроке имени (без регистра)   |
| `min_age`  | int    | —            | Минимальный возраст (включительно)        |
| `max_age`  | int    | —            | Максимальный возраст (включительно)       |
| `limit`    | int    | 20, макс 100 | Сколько вернуть                           |
| `offset`   | int    | 0            | Сколько пропустить                        |

**Пример**

```bash
curl 'http://localhost:8080/users?name=ali&min_age=18&max_age=65&limit=10&offset=0'
```

**Ответ 200**

```json
[
  { "id": 1, "name": "Alice", "email": "a@x.com", "age": 25 },
  { "id": 3, "name": "Alicia", "email": "c@x.com", "age": 30 }
]
```

### `GET /users/{id}`

Получить пользователя по ID.

**Ответ 200**

```json
{ "id": 1, "name": "Alice", "email": "a@x.com", "age": 25 }
```

**Ошибки**

- `400` — id не число
- `404` — нет такого пользователя

### `POST /users`

Создать пользователя. ID генерируется сервером.

**Тело запроса**

```json
{ "name": "Alice", "email": "alice@example.com", "age": 25 }
```

**Ответ 201**

```json
{ "id": 1, "name": "Alice", "email": "alice@example.com", "age": 25 }
```

**Ошибки**

- `400` — битый JSON
- `400` — не прошла валидация (см. `details`)

### `PUT /users/{id}`

Полностью обновить пользователя. Все поля обязательны.

**Ответ 200** — обновлённый объект.

**Ошибки**

- `400` — битый JSON, невалидный id, ошибка валидации
- `404` — пользователя нет

### `DELETE /users/{id}`

Удалить пользователя.

**Ответ 204** — тело пустое.

**Ошибки**

- `400` — невалидный id
- `404` — пользователя нет

## Коды ответов

| Код | Когда                                              |
|-----|----------------------------------------------------|
| 200 | Успешное чтение/обновление                         |
| 201 | Ресурс создан                                      |
| 204 | Ресурс удалён                                      |
| 400 | Невалидный ввод (json, id, validation, query)      |
| 404 | Ресурс не найден                                   |
| 500 | Внутренняя ошибка (подробности только в логах)     |

## Примеры

```bash
# Создать
curl -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"a@x.com","age":25}'

# Список с фильтром и пагинацией
curl 'http://localhost:8080/users?name=ali&limit=5&offset=0'

# Получить
curl http://localhost:8080/users/1

# Обновить
curl -X PUT http://localhost:8080/users/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice Smith","email":"as@x.com","age":26}'

# Удалить
curl -X DELETE http://localhost:8080/users/1

# Ошибка валидации
curl -i -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"","email":"broken","age":-3}'
```

## Структура проекта

```
laba8/
├── main.go                 — точка входа, маршруты, выбор хранилища
├── go.mod
├── docker-compose.yml      — PostgreSQL
├── models/
│   ├── user.go             — модель User
│   └── validation.go       — валидация + ValidationError
├── storage/
│   ├── storage.go          — интерфейс UserStorage
│   ├── filter.go           — UserFilter
│   ├── memory.go           — реализация в памяти
│   └── postgres.go         — реализация на PostgreSQL
└── handlers/
    ├── errors.go           — AppError + конструкторы
    ├── middleware.go       — Wrap + handleError
    ├── user.go             — обработчики маршрутов
    └── user_test.go        — HTTP-тесты
```