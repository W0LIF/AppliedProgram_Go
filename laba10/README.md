# Лабораторная работа №10 — защита информации в Go

Практическое применение криптографических примитивов стандартной библиотеки Go:
хэширование, симметричное и асимметричное шифрование, цифровая подпись, TLS с
взаимной аутентификацией, JWT-авторизация, роли, CSRF-защита и сессии.

Работа продолжает лабораторные №8–9 (REST API пользователей + CLI-клиент):
к существующему серверу добавлены все механизмы безопасности из задания.

---

## Содержание

- [Требования](#требования)
- [Структура проекта](#структура-проекта)
- [Быстрый старт](#быстрый-старт)
- [Задание 1. Хэширование данных](#задание-1-хэширование-данных)
- [Задание 2. Симметричное шифрование (AES-256-GCM)](#задание-2-симметричное-шифрование-aes-256-gcm)
- [Задание 3. Цифровая подпись (RSA-2048 + SHA-256)](#задание-3-цифровая-подпись-rsa-2048--sha-256)
- [Задание 4. TLS с взаимной аутентификацией](#задание-4-tls-с-взаимной-аутентификацией)
- [Задание 5. Защита REST API: JWT и роли](#задание-5-защита-rest-api-jwt-и-роли)
- [Задание 6. CSRF-защита и сессии](#задание-6-csrf-защита-и-сессии)
- [REST API](#rest-api)
- [Тестирование](#тестирование)
- [Известные ограничения](#известные-ограничения)
- [Типичные проблемы](#типичные-проблемы)

---

## Требования

- **Go** ≥ 1.26 (`go.mod` использует `go 1.26.0`; нужны method+path паттерны
  `http.ServeMux` и `r.PathValue`)
- **OpenSSL** — для генерации сертификатов
- **Docker** и **Docker Compose** — для PostgreSQL (необязательно: по умолчанию
  используется in-memory хранилище)
- **Windows**: Git Bash или MSYS2 для запуска `certs/gen.sh`

---

## Структура проекта

```
laba10/
├── main.go                 # HTTP-сервер: маршруты, mTLS, выбор хранилища
├── auth/
│   ├── jwt.go              # TokenManager: выпуск и проверка JWT (HS256)
│   ├── session.go          # SessionStore: in-memory сессии
│   └── csrf.go             # генерация CSRF-токенов
├── handlers/
│   ├── auth.go             # регистрация, JWT-логин, session-логин/логаут
│   ├── auth_middleware.go  # RequireAuth, RequireRole, контекст пользователя
│   ├── csrf.go             # EnsureCSRFCookie, проверка X-CSRF-Token
│   ├── user.go             # CRUD пользователей, фильтры и пагинация
│   ├── middleware.go       # HandlerFunc, Wrap, обработка ошибок, JSON
│   └── errors.go           # AppError: 400/401/403/404/500
├── models/
│   ├── user.go             # User
│   ├── auth.go             # RegisterRequest, LoginRequest, AuthResponse
│   └── validation.go       # валидация пользователя и пароля
├── storage/
│   ├── storage.go          # интерфейс UserStorage
│   ├── memory.go           # in-memory реализация
│   ├── postgres.go         # PostgreSQL-реализация и миграции
│   └── filter.go           # UserFilter
├── cmd/
│   ├── hash/               # утилита хэширования
│   ├── aes/                # AES-256-GCM
│   ├── rsa/                # RSA-2048: ключи, подпись, проверка
│   └── client/             # интерактивный CLI-клиент (mTLS + JWT)
├── certs/
│   └── gen.sh              # CA + серверный и клиентский сертификаты
├── docker-compose.yml      # PostgreSQL 18 на localhost:5434
├── go.mod
└── go.sum
```

---

## Быстрый старт

Все команды выполняются из каталога `laba10`, если не указано иное.

### 1. PostgreSQL (опционально)

```bash
docker compose up -d
```

- контейнер `lab10-postgres`;
- пользователь `lab`, пароль `lab`, база `lab10`;
- порт **5434** на хосте (внутри контейнера — 5432).

Если раньше в томе `pgdata` была создана другая база (например, `lab8`),
пересоздайте том:

```bash
docker compose down -v
docker compose up -d
```

### 2. TLS-сертификаты

```bash
cd certs
bash gen.sh
cd ..
```

Скрипт создаёт собственный CA (`CN=lab10-ca`) и подписывает им:

- `server.crt` / `server.key` — `CN=localhost`, SAN `DNS:localhost, IP:127.0.0.1`;
- `client.crt` / `client.key` — `CN=lab10-client` (клиентский сертификат для mTLS).

Все файлы остаются в `certs/`. Сертификаты и ключи добавлены в `.gitignore`.

### 3. Запуск сервера

```bash
# Linux/macOS
export JWT_SECRET="change-me-in-prod"
export STORAGE=postgres
export DATABASE_URL="postgres://lab:lab@localhost:5434/lab10?sslmode=disable"
go run .
```

```powershell
# PowerShell
$env:JWT_SECRET = "change-me-in-prod"
$env:STORAGE = "postgres"
$env:DATABASE_URL = "postgres://lab:lab@localhost:5434/lab10?sslmode=disable"
go run .
```

Без `STORAGE=postgres` сервер поднимется на in-memory хранилище — удобно для
быстрой проверки. Сервер слушает **https://localhost:8443** и требует
клиентский сертификат (mTLS) для любого запроса.

### 4. Запуск CLI-клиента

Клиент ищет сертификаты по относительным путям `certs/...`, поэтому запускать
его нужно **из корня `laba10`**:

```bash
API_URL=https://localhost:8443 go run ./cmd/client
```

```powershell
$env:API_URL = "https://localhost:8443"
go run ./cmd/client
```

JWT сохраняется в файл `.lab10-token` в текущей директории и подхватывается при
следующем запуске.

Меню клиента:

| Пункт | Действие |
|-------|----------|
| 1 | Список пользователей (фильтры, пагинация) |
| 2 | Пользователь по ID |
| 3 | Создать пользователя (только admin) |
| 4 | Обновить пользователя (только admin) |
| 5 | Удалить пользователя (только admin) |
| 6 | Войти (JWT) |
| 7 | Зарегистрироваться |
| 8 | Выйти (локально забыть JWT) |
| 0 | Выход из программы |

---

## Задание 1. Хэширование данных

**Файл:** `cmd/hash/main.go`

Утилита поддерживает **MD5**, **SHA-256**, **SHA-512** и режим проверки
целостности по известному хэшу.

```bash
go build -o hash ./cmd/hash
```

**Примеры:**

```bash
# SHA-256 от строки
./hash -algo=sha256 -text "hello"
# 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824

# MD5 и SHA-512
./hash -algo=md5    -text "hello"
./hash -algo=sha512 -text "hello"

# Проверка целостности
./hash -algo=sha256 -text "hello" \
       -check=2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
# result:    OK — хэши совпадают
```

- Сравнение выполняется без учёта регистра hex-строки.
- Коды возврата: `0` — успех, `1` — ошибка, `2` — неверные аргументы,
  `3` — хэши не совпали.

> ⚠️ MD5/SHA-256 не подходят для хранения паролей — только для контроля
> целостности. Пароли в API хранятся через `bcrypt`.

---

## Задание 2. Симметричное шифрование (AES-256-GCM)

**Файл:** `cmd/aes/main.go`

- Алгоритм: **AES-256-GCM**.
- Ключ выводится из пароля как `SHA-256(password)` (32 байта).
- Nonce генерируется из `crypto/rand`.
- Формат результата: `Base64( nonce || ciphertext || tag )`.
- GCM аутентифицирует данные: при неверном ключе или повреждении шифротекста
  расшифровка завершится ошибкой.

```bash
go build -o aes ./cmd/aes
```

**Примеры:**

```bash
# Шифрование
./aes -mode=enc -text "секретное сообщение" -key "my-strong-pass"
# -> "XjyOetLAw8WdtJOLyABJzmoO569s7rr7/2zMvfpBSZRUavt2+4rXZQ=="

# Расшифровка тем же ключом
./aes -mode=dec -text "XjyOetLAw8WdtJOLyABJzmoO569s7rr7/2zMvfpBSZRUavt2+4rXZQ==" \
      -key "my-strong-pass"
# -> "секретное сообщение"

# Неверный ключ
./aes -mode=dec -text "<тот же шифротекст>" -key "другой пароль"
# error: decrypt: cipher: message authentication failed ...
```

> ⚠️ В продакшене ключ из пароля выводят через `scrypt` или `argon2id`
> (`golang.org/x/crypto/scrypt`), а не через одиночный SHA-256.

---

## Задание 3. Цифровая подпись (RSA-2048 + SHA-256)

**Файл:** `cmd/rsa/main.go`

- Генерация: `rsa.GenerateKey(rand.Reader, 2048)`.
- Приватный ключ: PKCS#8 + PEM (`PRIVATE KEY`, права `0600`).
- Публичный ключ: PKIX + PEM (`PUBLIC KEY`).
- Подпись: `RSA-PKCS1v15` над SHA-256.
- Проверка: `rsa.VerifyPKCS1v15`.

```bash
go build -o rsa ./cmd/rsa
```

**Сценарий «две стороны»:**

```bash
# 1. Alice генерирует пару ключей
./rsa -mode=gen -priv=priv.pem -pub=pub.pem

# 2. Alice подписывает сообщение закрытым ключом
SIG=$(./rsa -mode=sign -priv=priv.pem -text "привет, мир")

# 3. Bob проверяет подпись открытым ключом
./rsa -mode=verify -pub=pub.pem -text "привет, мир" -sig="$SIG"
# подпись верна

# 4. Сообщение изменено — проверка не проходит
./rsa -mode=verify -pub=pub.pem -text "привет, МИР" -sig="$SIG"
# подпись НЕВЕРНА: crypto/rsa: verification error
```

Подпись кодируется в Base64, поэтому её удобно передавать в текстовом виде.
Код возврата при неверной подписи — `3`.

---

## Задание 4. TLS с взаимной аутентификацией

### Сервер

**Файл:** `main.go`

```go
tlsCfg := &tls.Config{
    ClientAuth: tls.RequireAndVerifyClientCert,
    ClientCAs:  pool, // доверяем только lab10-ca
    MinVersion: tls.VersionTLS12,
}
srv.ListenAndServeTLS("certs/server.crt", "certs/server.key")
```

Сервер разрывает соединение на этапе handshake, если клиент не предъявил
сертификат, подписанный `lab10-ca`.

### Клиент

**Файл:** `cmd/client/api.go`

```go
tlsCfg := &tls.Config{
    Certificates: []tls.Certificate{cert}, // клиентский сертификат
    RootCAs:      pool,                    // доверяем только своему CA
    ServerName:   "localhost",
    MinVersion:   tls.VersionTLS12,
}
```

- Загружаются `certs/client.crt`, `certs/client.key`, `certs/ca.crt`.
- `InsecureSkipVerify` **не используется**: сертификат сервера проверяется
  по CA и имени `localhost`.
- Клиент обязан предъявить свой сертификат — это и есть **mTLS**.

### Ручная проверка

Linux/macOS (curl с OpenSSL):

```bash
curl --cacert certs/ca.crt \
     --cert   certs/client.crt \
     --key    certs/client.key \
     https://localhost:8443/users
```

Без `--cert/--key` сервер вернёт ошибку TLS handshake.

> **Windows:** системный `curl.exe` собран с Schannel и не умеет принимать
> PEM-сертификат и отдельный ключ. Для ручных проверок используйте Python
> (пример ниже) или WSL/Linux.

```python
import http.client, ssl

ctx = ssl.create_default_context(cafile="certs/ca.crt")
ctx.load_cert_chain("certs/client.crt", "certs/client.key")
conn = http.client.HTTPSConnection("localhost", 8443, context=ctx)
conn.request("GET", "/users")
print(conn.getresponse().status)
```

---

## Задание 5. Защита REST API: JWT и роли

### JWT-менеджер

**Файл:** `auth/jwt.go`

```go
type Claims struct {
    UserID int    `json:"uid"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}
```

- Алгоритм подписи: **HS256**.
- TTL: 1 час, issuer: `"laba10"`.
- `Parse` принимает только `HS256` (`jwt.WithValidMethods`) и проверяет issuer —
  это защищает от подмены `alg` и чужих токенов.
- Секрет берётся из `JWT_SECRET`; при пустом значении используется
  dev-секрет и печатается предупреждение.

### Middleware

**Файл:** `handlers/auth_middleware.go`

```go
func RequireAuth(tm *auth.TokenManager, ss *auth.SessionStore, next HandlerFunc) HandlerFunc
func RequireRole(next HandlerFunc, roles ...string) HandlerFunc
```

Логика `RequireAuth`:

1. Если есть `Authorization: Bearer <jwt>` — проверяет токен и кладёт
   `uid`/`role` в `context.Context`.
2. Иначе берёт cookie `session_id` и ищет сессию в `SessionStore`.
3. Для небезопасных методов (не `GET`/`HEAD`/`OPTIONS`) при session-аутентификации
   проверяет CSRF.
4. Если аутентификация не удалась — `401 Unauthorized`.

`RequireRole` сравнивает `role` из контекста со списком разрешённых ролей и
возвращает `403 Forbidden` при несовпадении.

**Маршруты в `main.go`:**

```go
mux.HandleFunc("GET /users", ...)                    // любой авторизованный
mux.HandleFunc("POST /users", ...RequireRole(..., "admin"))
mux.HandleFunc("PUT /users/{id}", ...RequireRole(..., "admin"))
mux.HandleFunc("DELETE /users/{id}", ...RequireRole(..., "admin"))
```

### Поток аутентификации по JWT

```bash
# Регистрация (сразу выдаёт токен, роль всегда "user")
curl -X POST https://localhost:8443/auth/register \
     --cacert certs/ca.crt --cert certs/client.crt --key certs/client.key \
     -H 'Content-Type: application/json' \
     -d '{"name":"Alice","email":"a@x.com","age":25,"password":"secret1"}'
# -> { "token": "...", "user": { ... } }

# Логин
curl -X POST https://localhost:8443/auth/login \
     --cacert certs/ca.crt --cert certs/client.crt --key certs/client.key \
     -H 'Content-Type: application/json' \
     -d '{"email":"a@x.com","password":"secret1"}'

# Запрос с токеном
curl https://localhost:8443/users \
     --cacert certs/ca.crt --cert certs/client.crt --key certs/client.key \
     -H "Authorization: Bearer <token>"
```

### Как получить первого администратора

Регистрация всегда создаёт пользователя с ролью `user`. Первого администратора
нужно назначить вручную в PostgreSQL:

```bash
docker exec -it lab10-postgres psql -U lab -d lab10 \
  -c "UPDATE users SET role = 'admin' WHERE email = 'a@x.com';"
```

После этого нужно заново войти, чтобы получить JWT с новой ролью.

> При in-memory хранилище данные живут только до перезапуска сервера, а
> назначить роль через SQL нельзя. Для проверки ролей используйте PostgreSQL
> либо юнит-тесты.

---

## Задание 6. CSRF-защита и сессии

### Сессии

**Файл:** `auth/session.go`

- `SessionStore` — потокобезопасная in-memory карта (`sync.RWMutex`).
- ID — 32 случайных байта из `crypto/rand`, hex-кодирование.
- Методы: `Create`, `Get`, `Delete`.
- Cookie `session_id`: `HttpOnly`, `Secure`, `SameSite=Lax`, срок 24 часа.

### CSRF-токен

**Файл:** `auth/csrf.go`, `handlers/csrf.go`

- `NewCSRFToken()` — 32 случайных байта в hex.
- Cookie `csrf_token` ставится при session-логине (`HttpOnly: false`, чтобы
  токен мог прочитать фронтенд).
- На небезопасных методах токен проверяется по заголовку `X-CSRF-Token`.
- Сравнение — `subtle.ConstantTimeCompare` (защита от timing-атак).
- `EnsureCSRFCookie` выставляет cookie, если её ещё нет (например, перед
  логином).

### Session-логин и логаут

```bash
# Логин: создаёт сессию и обе cookie
curl -c jar.txt -X POST https://localhost:8443/auth/session-login \
     --cacert certs/ca.crt --cert certs/client.crt --key certs/client.key \
     -H 'Content-Type: application/json' \
     -d '{"email":"a@x.com","password":"secret1"}'
# -> { "user": ..., "csrf_token": "..." }

# Извлечь CSRF-токен из cookie jar
CSRF=$(awk '$6 == "csrf_token" { print $7 }' jar.txt)

# POST с сессией и CSRF
curl -b jar.txt -X POST https://localhost:8443/users \
     --cacert certs/ca.crt --cert certs/client.crt --key certs/client.key \
     -H "X-CSRF-Token: $CSRF" \
     -H 'Content-Type: application/json' \
     -d '{"name":"Bob","email":"b@x.com","age":30}'

# Без CSRF-заголовка — 403
curl -b jar.txt -X POST https://localhost:8443/users \
     --cacert certs/ca.crt --cert certs/client.crt --key certs/client.key \
     -H 'Content-Type: application/json' \
     -d '{"name":"Bob","email":"b@x.com","age":30}'

# Логаут: удаляет сессию и сбрасывает cookie
curl -b jar.txt -X POST https://localhost:8443/auth/session-logout \
     --cacert certs/ca.crt --cert certs/client.crt --key certs/client.key
```

CSRF проверяется только для session-аутентификации. Запросы с `Bearer`-токеном
CSRF не требуют: браузер не добавляет заголовок `Authorization` автоматически.

---

## REST API

Все запросы идут по HTTPS и требуют клиентский сертификат (mTLS).

| Метод | Путь | Доступ | Успех | Описание |
|-------|------|--------|-------|----------|
| POST | `/auth/register` | mTLS | `201` | Регистрация, возвращает JWT |
| POST | `/auth/login` | mTLS | `200` | JWT-логин |
| POST | `/auth/session-login` | mTLS | `200` | Сессия + CSRF-cookie |
| POST | `/auth/session-logout` | mTLS | `204` | Удаление сессии |
| GET | `/users` | JWT или сессия | `200` | Список пользователей |
| GET | `/users/{id}` | JWT или сессия | `200` | Пользователь по ID |
| POST | `/users` | admin | `201` | Создать пользователя |
| PUT | `/users/{id}` | admin | `200` | Обновить пользователя |
| DELETE | `/users/{id}` | admin | `204` | Удалить пользователя |

Параметры списка: `name` (подстрока, без учёта регистра), `min_age`, `max_age`,
`limit` (по умолчанию 20, максимум 100), `offset`.

Валидация пользователя: имя 2–100 символов, корректный email, возраст 1–150.
Пароль при регистрации: 6–72 байта (ограничение bcrypt).

`POST /users` создаёт запись пользователя без пароля: такой пользователь не
сможет войти, пока пароль не будет задан отдельно. Это admin-CRUD из
лабораторных №8–9, а не регистрация аккаунта.

Формат ошибок:

```json
{ "error": "validation failed", "details": { "email": "email is invalid" } }
```

---

## Тестирование

```bash
go test ./...
```

Покрыто:

- регистрация: `201`, `400` (слабый пароль), `400` (дубликат email);
- JWT-логин: `200`, `401` (неверный пароль / неизвестный email);
- `RequireAuth`: `401` без токена, `401` с невалидным, `200` с валидным;
- сессии: логин, доступ с cookie, логаут, `401` после логаута;
- CSRF: `403` без заголовка, `403` с неверным токеном, `201` с корректным;
- роли: `403` для `user`, `200` для `admin`;
- CRUD: `201`, `200`, `204`, `400` (валидация), `401`, `404`;
- фильтры и пагинация: `name`, `min_age`, `max_age`, `limit`, `offset`;
- ошибки `min_age > max_age`, `limit > 100`, `offset < 0`;
- валидация моделей и in-memory хранилище.

`go test -race` требует cgo и C-компилятора. На Windows без gcc запускайте
обычный `go test ./...`.

### Ручные проверки безопасности

| Проверка | Ожидаемый результат |
|----------|---------------------|
| `GET /users` без токена и cookie | `401` |
| `GET /users` с невалидным JWT | `401` |
| `POST /users` с cookie, но без `X-CSRF-Token` | `403` |
| `POST /users` с cookie и неверным CSRF | `403` |
| `POST /users` с JWT роли `user` | `403` |
| `POST /users` с JWT роли `admin` | `201` |
| Просроченный JWT | `401` |
| Запрос без клиентского сертификата | TLS handshake failure |
| Повторный запрос со старой сессией после логаута | `401` |

---

## Известные ограничения

1. **`JWT_SECRET`.** При пустом значении используется `dev-secret-change-me`
   и выводится предупреждение. В продакшене секрет обязателен, и приложение
   должно падать без него.
2. **In-memory сессии.** `SessionStore` не переживает перезапуск сервера и не
   чистит истёкшие записи. Cookie живёт 24 часа, но серверная сессия — до
   рестарта. Для масштабирования нужно внешнее хранилище (Redis).
3. **In-memory пользователи.** Без `STORAGE=postgres` данные теряются при
   перезапуске.
4. **Слабый KDF в AES-утилите.** `SHA-256(password)` — не KDF; для реальных
   паролей используйте `scrypt`/`argon2id`.
5. **Первый администратор.** Регистрация всегда выдаёт роль `user`; admin
   назначается SQL-запросом (см. выше).
6. **CLI-клиент работает только с JWT.** Session/CSRF-маршруты предназначены
   для браузера или ручных запросов.
7. **Сертификаты самоподписанные.** `gen.sh` подходит для лабораторной, но не
   для продакшена.
8. **Windows + curl.** Системный `curl.exe` (Schannel) не принимает PEM-пару
   `--cert/--key`. Используйте Python или Linux/WSL.

---

## Типичные проблемы

| Симптом | Причина и решение |
|---------|-------------------|
| `open certs/client.crt: no such file` | Клиент запущен не из корня `laba10`. Перейдите в `laba10` и запустите `go run ./cmd/client` |
| `pq: database "lab10" does not exist` | Старый том PostgreSQL. Выполните `docker compose down -v && docker compose up -d` |
| `connection refused` на `:5434` | PostgreSQL не запущен: `docker compose up -d` |
| `tls: client didn't provide a certificate` | Запрос без клиентского сертификата — mTLS обязателен |
| `x509: certificate signed by unknown authority` | У клиента нет `certs/ca.crt` или сертификат сервера выпущен другим CA |
| `invalid or expired token` | JWT просрочен (TTL 1 час) или подписан другим `JWT_SECRET` |
| `csrf header missing` | POST/PUT/DELETE с session-cookie без заголовка `X-CSRF-Token` |
| `WARNING: JWT_SECRET is not set` | Задайте `JWT_SECRET` перед запуском сервера |
