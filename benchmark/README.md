# Benchmark: PostgreSQL joins vs Elasticsearch denormalized reads

Проект сравнивает чтение нормализованных данных в PostgreSQL через `JOIN` с чтением агрегированных/денормализованных документов в Elasticsearch.

Основной сценарий запуска полностью контейнерный: на хосте нужны только Docker, Docker Compose и `make`.

## Быстрый старт

```bash
make smoke
```

`make smoke` поднимает PostgreSQL и Elasticsearch, генерирует маленький датасет, пересоздает схемы/индексы, загружает данные и запускает короткий бенчмарк.

Обычный локальный прогон:

```bash
make all
```

Параметры можно переопределять:

```bash
CONFIG=config.large.yaml WORKERS=8 ITERATIONS=100 make all
```

## Команды

```bash
make up            # запустить PostgreSQL и Elasticsearch
make generate      # сгенерировать CSV для PostgreSQL и NDJSON для Elasticsearch
make init-db       # пересоздать PostgreSQL schema + benchmark indexes
make init-elastic  # пересоздать индекс orders в Elasticsearch
make load          # загрузить сгенерированные данные
make benchmark     # запустить сценарии чтения
make listing       # запустить marketplace/listing read-model benchmark
make listing-large # большой marketplace/listing benchmark для отчета
make photos        # PostgreSQL bytea vs MinIO benchmark для фотографий
make photos-large  # большой benchmark фотографий, около 5GB payload
make index-joins-large # изолированный benchmark PostgreSQL index profiles + join modes
make smoke         # полный короткий проверочный прогон
make reset         # удалить контейнеры и volume с данными БД
make down          # остановить контейнеры без удаления volume
```

Все Go-команды выполняются через сервис `benchmark-runner` в `docker-compose.yml`.

## Конфиги данных

- `config.smoke.yaml` - быстрый проверочный датасет.
- `config.yaml` - локальный датасет по умолчанию.
- `config.large.yaml` - пример крупного датасета.

Формат:

```yaml
users: 10000
orders: 50000
order_items: 150000
products: 10000
categories: 1000
batch_size: 10000
workers: 4
time_range_days: 365
seed: 42
```

`order_items` должен быть не меньше `orders`, чтобы у каждого заказа была хотя бы одна позиция.

## Что измеряется

PostgreSQL:

- сложный `JOIN` по `users`, `orders`, `order_items`, `products`, `categories`;
- варианты join-планов: hash join, merge join, nested loop, default optimizer;
- point lookup по `orders.id`;
- range scan по `orders.created_at`;
- агрегация по дням.

Elasticsearch:

- terms aggregation по стране пользователя;
- range query по `created_at`;
- term query по стране пользователя.

Marketplace/listing read model:

- PostgreSQL search page через `listings`, `listing_sellers`, `listing_attribute_values`;
- PostgreSQL facets по городу/бренду с фильтрами;
- Elasticsearch search page по денормализованному документу `listings`;
- Elasticsearch facets/aggregations по тем же фильтрам.

Photo/object storage:

- PostgreSQL хранит бинарные payload-ы в `photo_blobs.data BYTEA`;
- MinIO хранит те же payload-ы как S3 objects;
- отчет показывает размер PostgreSQL relation/индексов/raw payload, сумму размеров MinIO objects и random-read latency/throughput.

Метрики: total requests, duration, throughput, avg/min/max latency, p50/p95/p99.

## Структура

```text
benchmark/                  # runner и сценарии
cmd/                        # CLI: init-elastic, load, benchmark
executors/                  # PostgreSQL и Elasticsearch query executors
generator/                  # генерация нормализованных CSV и denormalized NDJSON
postgres/init/              # DDL и индексы
elasticsearch/              # settings/mapping индекса orders
output/                     # generated data, игнорируется git
```

## Расширение под PostgreSQL vs MinIO

Текущая структура разделяет генерацию, загрузку и исполнители запросов. Для второго этапа стоит добавлять новый backend без переписывания существующих сценариев:

1. Добавить storage-specific генератор/экспортер в `generator` или отдельный writer, например parquet/csv/object layout.
2. Добавить сервис MinIO в `docker-compose.yml` и отдельную команду инициализации bucket.
3. Добавить executor в `executors/`, например `FileExecutor` или `ObjectStorageExecutor`.
4. Расширить runner сценариями чтения файлов, не смешивая их с SQL/Elasticsearch query DSL.
5. Добавить `make init-minio`, `make load-files`, `make benchmark-files`.

Так PostgreSQL, Elasticsearch и будущий MinIO останутся независимыми backend-ами с общим runner/reporting слоем.

## Troubleshooting

```bash
make ps
make logs
make reset
```

Если меняли DDL или mapping, используйте `make reset` или полный `make all`, чтобы избежать старых volume-состояний.
