# Benchmark: PostgreSQL vs Elasticsearch

Ультимативный бенчмарк для сравнения различных подходов чтения данных в PostgreSQL и Elasticsearch.

## 🎯 Возможности

### PostgreSQL
- **Join алгоритмы**: Hash Join, Merge Join, Nested Loop
- **Типы индексов**: B-Tree, Hash, BRIN, частичные индексы
- **Сценарии**:
  - Сложные JOIN запросы (5 таблиц)
  - Point lookup (поиск по первичному ключу)
  - Range scan (диапазонные запросы по времени)
  - Агрегации (GROUP BY, SUM, COUNT)

### Elasticsearch
- **Агрегации**: terms, metrics, date_histogram
- **Поиск**: full-text, filtered queries
- **Range queries**: временные диапазоны

## 📊 Генерация данных

Генератор создаёт реалистичные данные с распределением Zipf:
- **Users**: 1M записей (10 стран, 3 статуса)
- **Categories**: 10K записей (иерархическая структура)
- **Products**: 1M записей (с привязкой к категориям)
- **Orders**: 10M записей (за последний год)
- **Order Items**: 50M записей (1-12 позиций на заказ)

## 🚀 Быстрый старт

### Полный запуск "из коробки" (рекомендуется)

```bash
# 1. Клонировать репозиторий и перейти в директорию
git clone <repository-url>
cd benchmark

# 2. Запустить всё одной командой (контейнеры + генерация + загрузка + бенчмарк)
make all
```

### Пошаговый запуск

#### Шаг 1: Запуск инфраструктуры

```bash
# Запуск PostgreSQL, Elasticsearch и Kibana
make docker-up

# Проверка статуса
docker-compose ps
```

#### Шаг 2: Генерация тестовых данных

```bash
# Полная генерация (PostgreSQL + Elasticsearch)
make generate

# Альтернативно через Go:
go run generator/cmd/main.go -config config.yaml -out ./output
```

#### Шаг 3: Инициализация баз данных

```bash
# PostgreSQL - схема создаётся автоматически через docker-entrypoint-initdb.d
# Дополнительные индексы для бенчмарка:
docker-compose exec -T postgres psql -U benchmark -d benchmark -f /docker-entrypoint-initdb.d/benchmark_indexes.sql

# Elasticsearch - создание индекса с настройками
./scripts/init-elastic.sh
```

#### Шаг 4: Загрузка данных

```bash
# Загрузка CSV в PostgreSQL и NDJSON в Elasticsearch
make load
```

#### Шаг 5: Запуск бенчмарков

```bash
# Запуск всех сценариев
make benchmark
```

## 📁 Структура проекта

```
benchmark/
├── cmd/
│   └── main.go              # Точка входа бенчмарка
├── benchmark/
│   ├── runner.go            # Движок бенчмарка
│   └── metrics.go           # Сбор и расчёт метрик
├── executors/
│   ├── postgres.go          # PostgreSQL executor
│   └── elastic.go           # Elasticsearch executor
├── generator/
│   ├── cmd/main.go          # Генератор данных
│   └── internal/
│       ├── config/          # Конфигурация
│       ├── distribution/    # Распределения (Zipf, Pareto)
│       ├── generator/       # Генераторы сущностей
│       ├── model/           # Модели данных
│       └── writer/          # Writers (CSV, NDJSON)
├── postgres/
│   └── init/
│       ├── base_ddl.sql     # Основная схема
│       └── benchmark_indexes.sql # Индексы для тестов
├── elasticsearch/
│   └── index-settings.json  # Настройки индекса
├── scripts/
│   ├── init-elastic.sh      # Инициализация Elasticsearch
│   └── init-postgres.sh     # Инициализация PostgreSQL
├── docker-compose.yml       # Инфраструктура
├── config.yaml              # Конфиг генератора
├── Makefile                 # Команды
└── README.md                # Этот файл
```

## 🔧 Конфигурация

### config.yaml

```yaml
# Размеры датасетов
users: 1000000        # 1M пользователей
orders: 10000000      # 10M заказов
order_items: 50000000 # 50M позиций
products: 1000000     # 1M товаров
categories: 10000     # 10K категорий

# Настройки пакетной обработки
batch_size: 50000
workers: 8

# Временной диапазон (дней)
time_range_days: 365

# Seed для воспроизводимости
seed: 42
```

### Изменение размера датасета

Для быстрого тестирования можно уменьшить данные в `config.yaml`:

```yaml
users: 10000        # 10K пользователей
orders: 100000      # 100K заказов
order_items: 500000 # 500K позиций
products: 10000     # 10K товаров
categories: 100     # 100 категорий
```

## 📈 Сценарии бенчмарка

### PostgreSQL Join сценарии

| Сценарий | Описание |
|----------|----------|
| Hash Join | Принудительное использование Hash Join |
| Merge Join | Принудительное использование Merge Join |
| Nested Loop | Принудительное использование Nested Loop |
| Default | Выбор оптимизатора PostgreSQL |

### Тестовые запросы

1. **Join Scenario** - сложный 5-табличный JOIN с агрегацией
2. **Point Lookup** - поиск заказа по ID
3. **Range Scan** - заказы за период (7/30/90 дней)
4. **Aggregation** - дневная агрегация заказов

### Elasticsearch сценарии

1. **Aggregation** - группировка по странам
2. **Range Query** - заказы за период
3. **Term Query** - поиск по стране

## 📊 Метрики

Бенчмарк собирает следующие метрики:

- **Throughput**: запросов в секунду (RPS)
- **Latency**: средняя, мин, макс
- **Percentiles**: p50, p95, p99
- **Duration**: общее время выполнения сценария

## 🔬 Типы индексов PostgreSQL

### B-Tree (по умолчанию)
```sql
CREATE INDEX idx_orders_user_id ON orders(user_id);
```

### Hash (для equality)
```sql
CREATE INDEX idx_orders_id_hash ON orders USING HASH (id);
```

### BRIN (для time-series)
```sql
CREATE INDEX idx_orders_created_at_brin ON orders USING BRIN (created_at);
```

### Частичные индексы
```sql
CREATE INDEX idx_orders_recent ON orders (created_at) 
WHERE created_at > NOW() - interval '90 days';
```

### Покрывающие индексы
```sql
CREATE INDEX idx_orders_covering 
ON orders (created_at, user_id) INCLUDE (total_amount, status);
```

## 🧹 Очистка

```bash
# Удалить сгенерированные данные
make clean

# Остановить контейнеры
make docker-down

# Полная очистка
make clean && make docker-down
```

## 📝 Требования

- Docker & Docker Compose
- Go 1.23+
- 16GB+ RAM (рекомендуется для больших датасетов)
- 50GB+ свободного места на диске

## ⚠️ Замечания

1. **Память**: Для генерации полного датасета (10M заказов) требуется ~8GB RAM
2. **Время**: Генерация данных занимает 10-30 минут в зависимости от hardware
3. **Elasticsearch**: Индексация 10M документов может занять 5-15 минут

## 🔍 Troubleshooting

### PostgreSQL не запускается
```bash
docker-compose logs postgres
```

### Elasticsearch недоступен
```bash
curl http://localhost:9200/_cluster/health?pretty
```

### Ошибка при загрузке данных
Проверьте, что файлы сгенерированы:
```bash
ls -lh ./output/
```

### Ошибка компиляции Go
```bash
# Установить зависимости
go mod download
go mod tidy
```

## 📚 Дополнительные ресурсы

- [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/using-explain.html)
- [Elasticsearch Query DSL](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl.html)
- [Use The Index, Luke](https://use-the-index-luke.com/)
