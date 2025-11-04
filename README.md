# Search Service

Микросервис полнотекстового поиска обращений с поддержкой фильтрации, курсорной пагинации и адаптивных стратегий выполнения запросов.

## Возможности

- **Модели разрешений:** `assigned_to_me`, `assigned_to_my_team`, `assigned_to_teams`, `all_except`, `all`
- **Фильтрация:** команды, менеджеры, теги, статусы, клиенты, компании, удалённые сообщения
- **Полнотекстовый поиск:** PostgreSQL FTS с поддержкой русского языка (минимум 3 символа)
- **Курсорная пагинация:** стабильное упорядочивание по релевантности и времени
- **Адаптивные стратегии:** автоматический выбор между filter-first и search-first
- **Индексация событий:** обработка сообщений и заметок через RabbitMQ

## Стек технологий

- **Go 1.24** — основной язык
- **PostgreSQL** — хранение и индексация (`tsvector`, GIN индексы)
- **RabbitMQ** — очередь событий
- **gorilla/mux** — HTTP роутинг
- **oapi-codegen v2** — кодогенерация из OpenAPI
- **Prometheus** — метрики
- **servicelib/swlog** — структурированное логирование

## Архитектура

```
search-service/
├── main.go, app.go          # Инициализация, запуск, graceful shutdown
├── config/                  # Конфигурация через scconfig
├── api/                     # HTTP сервер
│   └── handlers/            # Обработчики эндпоинтов
├── service/                 # Бизнес-логика
│   └── appeal/              # Сервис поиска обращений
│       ├── mapper/          # Преобразование моделей
│       ├── querybuilder/    # Построение поисковых запросов
│       └── textutils/       # Утилиты текста (HTML→текст)
├── database/                # Репозиторий (интерфейсы)
│   └── appeal/              # PostgreSQL реализация
│       └── sql/             # SQL-шаблоны с go:embed
├── events/appeal/           # Обработка событий RabbitMQ
│   ├── handlers/            # Обработчики типов событий
│   └── models/              # Модели событий
├── models/appeal/           # Доменные модели
├── openapi/                 # OpenAPI спецификация
│   ├── components/          # Компоненты схем
│   └── paths/               # Описания эндпоинтов
└── gen/                     # Сгенерированный код (api.gen.go)
```

## Конфигурация

Все параметры задаются через переменные окружения (используется `scconfig`):

**PostgreSQL:**
- `DB_HOST` (по умолчанию: `localhost`)
- `DB_PORT` (по умолчанию: `5432`)
- `DB_NAME` (по умолчанию: `search_service`)
- `DB_SESSION_USERNAME` (по умолчанию: `postgres`)
- `DB_SESSION_PASSWORD` (обязательный)

**RabbitMQ:**
- `RABBIT_HOST` (обязательный)
- `RABBIT_PORT` (обязательный)
- `RABBIT_USER` (обязательный)
- `RABBIT_PASSWORD` (обязательный)

**HTTP сервер:** порт `80` (хардкод в `config/models.go`)

Пример запуска (PowerShell):
```powershell
$env:DB_SESSION_PASSWORD = "password"
$env:RABBIT_HOST = "localhost"
$env:RABBIT_PORT = "5672"
$env:RABBIT_USER = "guest"
$env:RABBIT_PASSWORD = "guest"
.\search-service.exe
```

## API

**Базовый путь:** `/v1`

**Служебные эндпоинты:**
- `GET /health` — проверка живости
- `GET /metrics` — метрики Prometheus
- `GET /debug/pprof` — профилирование
- `GET /v1/swaggerui` — документация Swagger UI

### GET /v1/manager/appeals

Основной метод поиска обращений.

**Обязательные параметры:**
- `managerId` (UUID) — ID менеджера
- `appealPermission` (enum) — модель доступа (`assigned_to_me`, `assigned_to_my_team`, `assigned_to_teams`, `all_except`, `all`)

**Фильтры:**
- `appealRestrictions` (UUID[]) — ограничения по командам (для `assigned_to_teams`/`all_except`)
- `teamIds`, `managerIds`, `tagIds` (UUID[])
- `status` (enum) — `active`, `closed`, `snoozed`
- `clientId`, `companyId` (UUID)
- `hasDeletedMessages` (bool)

**Поиск и пагинация:**
- `search` (string) — текст запроса (мин. 3 символа)
- `size` (int) — размер страницы (по умолчанию 20, макс. 200)
- `nextCursor` (string) — курсор следующей страницы (base64 JSON)

**Формат ответа:**
```json
{
  "results": [
    {
      "id": 123,
      "searchRank": 0.85,
      "messageId": "uuid...",
      "status": "active",
      "client": { "id": "...", "fullName": "..." },
      "company": { "id": "...", "name": "...", "isVip": false },
      "isImportant": true,
      "isMentioned": false,
      "tags": [],
      "messagesWithUnreadMention": []
    }
  ],
  "meta": {
    "total": 42,
    "hasMore": true,
    "nextCursor": "eyJyYW5rIjowLjUsImFwcGVhbElkIjoxMjN9"
  }
}
```

## Архитектура поиска

### Модель данных

Сервис работает с обращениями на основе следующей модели:
- **Единственный клиент:** Каждое обращение связано с одним клиентом через поле `clientId` в таблице `appeals`
- **VIP статус:** `isImportant = TRUE` если клиент или его компания имеют VIP статус
- **Фильтрация:** Поиск по `clientId` или `companyId` находит обращения соответствующего клиента/компании

### Индексация

Используется единый `tsvector content` с весами для разных типов данных:
- **Таблицы:** `search_service_messages`, `search_service_notes`
- **Язык:** русский (`to_tsvector('russian', ...)`)
- **Индекс:** GIN
- **Ранжирование:** `ts_rank_cd(content, to_tsquery('russian', :query))`
- **Упрощение:** отказ от раздельных векторов (content/meta/idTokens)

### Стратегии выполнения

Репозиторий автоматически выбирает стратегию:
- **filter-first** — когда есть селективные фильтры (clientId, companyId) или отсутствует текст поиска
- **search-first** — когда поиск по тексту является основным критерием

Шаблоны: `database/appeal/sql/{filter_first,search_first}.sql`

### Курсорная пагинация

Курсор содержит: `{rank, appealId, createdAt}` (base64 JSON)

Обеспечивает стабильный порядок при повторных запросах даже при изменении данных.

## Обработка событий

Сервис подписан на очередь RabbitMQ для индексации:
- **Сообщения клиентов** (`client_message`) — индексация в `search_service_messages`
- **Сообщения менеджеров** (`manager_message`) — индексация в `search_service_messages`
- **Удаление сообщений** (`message_deleted`) — пометка удалённых

Обработчики в `events/appeal/handlers/`

## Метрики и мониторинг

**Prometheus метрики:**
- `search_requests_total{status,permission}` — количество запросов
- `search_latency_seconds{status,permission}` — латентность
- `search_results_found_total{permission}` — найденные результаты
- `messages_saved_total{type}` — сохранённые сообщения
- `message_processing_errors_total{type}` — ошибки обработки

**Порог медленного поиска:** 5 секунд (логируется как ERROR)

## Запуск

### Локально

```bash
go mod download
go build
$env:DB_SESSION_PASSWORD="password"
$env:RABBIT_PASSWORD="password"
.\search-service.exe
```

### Docker

```bash
docker build -t search-service .
docker run -p 80:80 \
  -e DB_HOST=postgres \
  -e DB_SESSION_PASSWORD=password \
  -e RABBIT_HOST=rabbitmq \
  -e RABBIT_PASSWORD=password \
  search-service
```

## Разработка

### Генерация кода из OpenAPI

```bash
make generate
```

Генерирует `gen/api.gen.go` на основе `openapi/openapi.yaml`

### Тестирование

```bash
go test ./...                    # Все тесты
go test ./database/appeal/...    # Репозиторий
go test ./service/appeal/...     # Сервисный слой
```

### Процесс изменений

1. Обновить OpenAPI спецификацию (`openapi/`)
2. Сгенерировать код (`make generate`)
3. Реализовать обработчики (`api/handlers/`)
4. Добавить бизнес-логику (`service/appeal/`)
5. Написать тесты
6. Проверить линтером (`.golangci.yaml`)

---

**Go:** 1.24  
**Архитектурные принципы:** `AGENTS.md`
