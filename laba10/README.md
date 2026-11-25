# Лабораторная работа №10 — Защита информации в Go

Практическое применение криптографических примитивов стандартной библиотеки Go:
хэширование, симметричное и асимметричное шифрование, цифровая подпись, TLS с
взаимной аутентификацией, JWT-авторизация, роли, CSRF-защита и сессии.

Работа является продолжением лабораторных №8–9 (REST API + CLI-клиент): к уже
существующему серверу пользователей добавлены все механизмы безопасности из
задания.

---

## Содержание

- [Требования](#требования)
- [Быстрый старт](#быстрый-старт)
- [Задание 1. Хэширование данных](#задание-1-хэширование-данных)
- [Задание 2. Симметричное шифрование (AES-256-GCM)](#задание-2-симметричное-шифрование-aes-256-gcm)
- [Задание 3. Цифровая подпись (RSA-2048 + SHA-256)](#задание-3-цифровая-подпись-rsa-2048--sha-256)
- [Задание 4. TLS с взаимной аутентификацией](#задание-4-tls-с-взаимной-аутентификацией)
- [Задание 5. Защита REST API: JWT и роли](#задание-5-защита-rest-api-jwt-и-роли)
- [Задание 6. CSRF-защита и сессии](#задание-6-csrf-защита-и-сессии)
- [Тестирование](#тестирование)
- [Известные ограничения](#известные-ограничения)

---

## Требования

- **Go** ≥ 1.22 (используются `http.ServeMux` с method+path паттернами и `r.PathValue`)
- **OpenSSL** (для `gen.sh`)
- **Docker** и **Docker Compose** (для PostgreSQL)
- Windows: Git Bash или MSYS2 для запуска `gen.sh`

---

## Быстрый старт

### 1. Поднять PostgreSQL

```bash
docker compose up -d
# контейнер lab10-postgres на порту 5434
```

### 2. Сгенерировать TLS-сертификаты

```bash
chmod +x gen.sh
./gen.sh

mkdir -p certs
mv ca.crt server.* client.* certs/
```

### 3. Запустить сервер

```bash
export JWT_SECRET="change-me-in-prod"
export DB_DSN="postgres://lab:lab@localhost:5434/lab8?sslmode=disable"
go run ./cmd/server
```

### 4. Запустить CLI-клиент

```bash
cd cmd/client
go build -o client
API_URL=https://localhost:8443 ./client
```

Меню клиента:

| Пункт | Действие |
|-------|----------|
| 1 | Список пользователей (фильтры, пагинация) |
| 2 | Пользователь по ID |
| 3 | Создать пользователя |
| 4 | Обновить пользователя |
| 5 | Удалить пользователя |
| 6 | Войти (JWT) |
| 7 | Зарегистрироваться |
| 8 | Выйти (logout) |
| 0 | Выход из программы |

JWT-токен сохраняется в файл `.lab9-token` в рабочей директории клиента.

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

# MD5
./hash -algo=md5 -text "hello"

# SHA-512
./hash -algo=sha512 -text "hello"

# Проверка целостности
./hash -algo=sha256 -text "hello" \
       -check=2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
# result:    OK — хэши совпадают
```

Сравнение выполняется без учёта регистра (`equalFoldHex`), что удобно при
копировании хэша из внешнего источника.

> ⚠️ MD5 и SHA-256 не годятся для хранения паролей — только для контроля
> целостности данных. Для паролей см. `bcrypt` в `internal/handlers/auth.go`.

---

## Задание 2. Симметричное шифрование (AES-256-GCM)

**Файл:** `cmd/aes/main.go`

- Алгоритм: **AES-256-GCM**.
- Ключ выводится из пароля через `sha256(password)` (32 байта).
- Nonce генерируется из `crypto/rand`.
- Формат результата: `Base64( nonce || ciphertext || tag )`.
- GCM даёт аутентификацию — при неверном ключе или порче шифротекста
  расшифровка вернёт ошибку.

```bash
go build -o aes ./cmd/aes
```

**Примеры:**

```bash
# Шифрование
./aes -mode=enc -text "секретное сообщение" -key "my-strong-pass"
# -> "c2FtcGxlLWJhc2U2NC1jaXBoZXJ0ZXh0..."

# Расшифровка
./aes -mode=dec -text "c2FtcGxlLWJhc2U2NC1jaXBoZXJ0ZXh0..." -key "my-strong-pass"
# -> "секретное сообщение"

# Неверный ключ — ошибка
./aes -mode=dec -text "<тот же шифротекст>" -key "другой пароль"
# error: decrypt: ... (неверный ключ или повреждённый шифротекст)
```

> ⚠️ В продакшене для вывода ключа из пароля используйте `scrypt` или
> `argon2id` (`golang.org/x/crypto/scrypt`), а не голый SHA-256.

---

## Задание 3. Цифровая подпись (RSA-2048 + SHA-256)

**Файл:** `cmd/rsa/main.go`

- Генерация ключей: `rsa.GenerateKey(rand.Reader, 2048)`
- Приватный ключ: PKCS#8 + PEM (`PRIVATE KEY`)
- Публичный ключ: PKIX + PEM (`PUBLIC KEY`)
- Подпись: `RSA-PKCS1v15` над SHA-256
- Проверка: `rsa.VerifyPKCS1v15`

```bash
go build -o rsa ./cmd/rsa
```

**Полный сценарий:**

```bash
# 1. Генерация пары ключей
./rsa -mode=gen -priv=priv.pem -pub=pub.pem

# 2. Подписать сообщение
SIG=$(./rsa -mode=sign -priv=priv.pem -text "привет, мир")
echo "$SIG"

# 3. Проверить подпись
./rsa -mode=verify -pub=pub.pem -text "привет, мир" -sig="$SIG"
# подпись верна

# 4. Проверить на изменённом сообщении — ошибка
./rsa -mode=verify -pub=pub.pem -text "привет, МИР" -sig="$SIG"
# подпись НЕВЕРНА: ...
```

---

## Задание 4. TLS с взаимной аутентификацией

### Генерация сертификатов

**Файл:** `gen.sh`

Скрипт создаёт собственный CA (`lab10-ca`) и подписывает им:

- **server.crt** — CN=localhost, SAN `DNS:localhost, IP:127.0.0.1`;
- **client.crt** — CN=lab10-client (для mutual TLS).

```bash
./gen.sh
# OK: ca.crt, server.{crt,key}, client.{crt,key}
```

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

- Клиент загружает `certs/client.crt`, `certs/client.key` и `certs/ca.crt`.
- **`InsecureSkipVerify` не используется.**
- Сервер обязан предъявить сертификат, подписанный `lab10-ca`, — иначе
  TLS-хэндшейк провалится.
- Клиент тоже обязан предъявить сертификат — это и есть **mTLS**.

### Ручная проверка

```bash
curl --cacert certs/ca.crt \
     --cert   certs/client.crt \
     --key    certs/client.key \
     https://localhost:8443/users
```

Без `--cert` сервер разорвёт соединение на этапе хэндшейка.

---

## Задание 5. Защита REST API: JWT и роли

### JWT-менеджер

**Файл:** `internal/auth/jwt.go`

```go
type Claims struct {
    UserID int    `json:"uid"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}
```

- Алгоритм подписи: **HS256**
- TTL: 1 час, issuer: `"laba10"`
- `Parse` проверяет, что метод подписи — HMAC (защита от подмены `alg`)
- Принимает `secret` из аргумента конструктора; в проде — из `JWT_SECRET`

### Middleware

**Файл:** `internal/handlers/auth_middleware.go`

```go
func RequireAuth(tm *auth.TokenManager, ss *auth.SessionStore, next HandlerFunc) HandlerFunc
func RequireRole(next HandlerFunc, roles ...string) HandlerFunc
```

Логика `RequireAuth`:

1. Если есть `Authorization: Bearer <jwt>` — парсит токен, кладёт `uid`/`role`
   в `context.Context`.
2. Иначе — берёт `session_id` cookie, ищет сессию в `SessionStore`.
3. Для небезопасных методов (не `GET`/`HEAD`/`OPTIONS`) — проверяет CSRF.
4. Иначе — `401 Unauthorized`.

`RequireRole` проверяет `role` из контекста по списку разрешённых.

**Пример использования:**

```go
mux.HandleFunc("GET /admin/users",
    Wrap(RequireAuth(tm, ss, RequireRole(h.AdminList, "admin"))))
```

### Поток аутентификации по JWT

```bash
# Регистрация (сразу выдаёт токен)
curl -X POST https://localhost:8443/auth/register \
     -H 'Content-Type: application/json' \
     -d '{"name":"Alice","email":"a@x.com","age":25,"password":"secret1"}'
# -> { "token": "...", "user": { ... } }

# Логин
curl -X POST https://localhost:8443/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"a@x.com","password":"secret1"}'

# Запрос с токеном
curl https://localhost:8443/users -H "Authorization: Bearer <token>"
```

---

## Задание 6. CSRF-защита и сессии

### Сессии

**Файл:** `internal/auth/session.go`

- `SessionStore` — потокобезопасная in-memory карта (`sync.RWMutex`).
- ID — 32 случайных байта из `crypto/rand`, закодированных в hex.
- Методы: `Create`, `Get`, `Delete`.
- Хранилище легко заменяется на Redis/БД — интерфейс минимален.

### CSRF-токен

**Файл:** `internal/auth/csrf.go`

`NewCSRFToken()` → 32 случайных байта в hex. Токен:

- выставляется в cookie `csrf_token` (`HttpOnly: false` — JS должен прочитать);
- проверяется на небезопасных методах через заголовок `X-CSRF-Token`;
- сравнивается через `subtle.ConstantTimeCompare` — защита от timing-атак.

### Session-логин

**Файл:** `internal/handlers/auth.go`

```bash
POST /auth/session/login
```

- проверяет email/password (`bcrypt`);
- создаёт сессию, ставит cookie `session_id` (`HttpOnly`, `Secure`, `SameSite=Lax`);
- ставит cookie `csrf_token`;
- возвращает `{ "user": ..., "csrf_token": "..." }`.

```bash
POST /auth/session/logout
```

- удаляет сессию из `SessionStore`;
- ставит `MaxAge: -1` для обеих cookie.

### Защита в middleware

```go
switch r.Method {
case http.MethodGet, http.MethodHead, http.MethodOptions:
    // безопасные методы — CSRF не нужен
default:
    if err := checkCSRF(r); err != nil {
        return err
    }
}
```

`checkCSRF` (`internal/handlers/csrf.go`):

1. Cookie `csrf_token` есть и не пуста?
2. Заголовок `X-CSRF-Token` есть?
3. Значения совпадают (`ConstantTimeCompare`)?

Иначе — `403 Forbidden`.

### `EnsureCSRFCookie`

Опциональный middleware: если у клиента ещё нет `csrf_token`, генерирует и
выставляет его — например, на странице логина, чтобы фронтенд мог прочитать
токен заранее.

### Пример работы с cookie через `curl`

```bash
# Логин по сессии
curl -c jar.txt -X POST https://localhost:8443/auth/session/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"a@x.com","password":"secret1"}'

# Извлечь CSRF-токен
CSRF=$(grep csrf_token jar.txt | awk '{print $7}')

# Запрос с сессией + CSRF (POST)
curl -b jar.txt -X POST https://localhost:8443/users \
     -H "X-CSRF-Token: $CSRF" \
     -H 'Content-Type: application/json' \
     -d '{"name":"Bob","email":"b@x.com","age":30}'

# Запрос без CSRF — 403
curl -b jar.txt -X POST https://localhost:8443/users \
     -H 'Content-Type: application/json' \
     -d '{"name":"Bob","email":"b@x.com","age":30}'
```

---

## Тестирование

### Unit-тесты

```bash
go test ./internal/handlers/... -v
go test ./... -race
```

Покрыты:

- `POST /auth/register`: `201`, `400` (слабый пароль), `400` (дубликат email)
- `POST /auth/login`: `200`, `401` (неверный пароль / неизвестный email)
- `RequireAuth`: `401` без токена, `401` с невалидным, `200` с валидным
- CRUD пользователей: `201`, `200`, `204`, `400` (валидация), `401`, `404`
- Фильтрация и пагинация: `name`, `min_age`, `max_age`, `limit`, `offset`
- Ошибки `min_age > max_age`, `limit > 100`, `offset < 0`

### Ручные проверки безопасности

| Проверка | Ожидаемый результат |
|----------|---------------------|
| `GET /users` без токена и cookie | `401` |
| `POST /users` с cookie, но без `X-CSRF-Token` | `403` |
| `POST /users` с cookie и неверным CSRF | `403` |
| `POST /users` с `Bearer` JWT (без CSRF) | `201` — JWT-путь не требует CSRF |
| `GET /admin/*` с ролью `user` | `403` |
| Просроченный JWT | `401` |
| Запрос без клиентского сертификата (mTLS) | TLS handshake failure |
| Порт `5432` вместо `5434` в DSN | ошибка подключения к БД |

---

## Известные ограничения

1. **Тесты используют устаревший API.** В `auth_test.go` и `user_test.go`
   встречаются вызовы старой сигнатуры:
   ```go
   tokens := storage.NewTokenStore()
   a := NewAuthHandler(store, tokens)
   token := tokens.Issue(u.ID)
   ```
   Текущий код работает с `*auth.TokenManager` и `*auth.SessionStore`:
   ```go
   tm := auth.NewTokenManager("secret")
   ss := auth.NewSessionStore()
   a := NewAuthHandler(store, tm, ss)
   token, _ := tm.Issue(u.ID, u.Role)
   ```
   Перед запуском `go test` эти файлы нужно привести к актуальному API.

2. **`JWT_SECRET`.** В `NewTokenManager` есть fallback `"dev-secret-change-me"`.
   Для локальной отладки — ок, для продакшена — недопустимо. Секрет должен
   приходить из env, а при его отсутствии приложение обязано падать.

3. **In-memory сессии.** `SessionStore` теряет данные при рестарте сервера.
   Для горизонтального масштабирования нужно внешнее хранилище (Redis).

4. **Слабый KDF в AES-утилите.** `sha256(password)` — не KDF. Для реальных
   паролей используйте `scrypt`/`argon2id`.

5. **Порт PostgreSQL.** `docker-compose.yml` публикует БД на **`5434`**, а не
   на `5432`. Указывайте это в `DB_DSN`.

6. **`certs/`.** `api.go` ожидает файлы по путям `certs/client.crt`,
   `certs/client.key`, `certs/ca.crt`. После `gen.sh` переместите их в
   `certs/` (или запускайте скрипт из `certs/`).
