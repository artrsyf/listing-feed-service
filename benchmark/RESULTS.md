# Benchmark Results

Этот файл фиксирует итоговые результаты больших прогонов benchmark-а:

- PostgreSQL normalized reads vs Elasticsearch denormalized reads для marketplace/listing read model.
- PostgreSQL `BYTEA` vs MinIO для хранения и чтения фотографий.

Все прогоны выполнялись через Docker Compose из `Makefile`.

## 0. Isolated PostgreSQL Index/Join Benchmark

### Цель

Изолированно сравнить влияние разных наборов PostgreSQL-индексов и разных join-алгоритмов на чтение нормализованных order-данных.

В отличие от общего `make benchmark`, этот прогон для каждого index profile:

1. Дропает все вторичные индексы на order-схеме.
2. Создает только индексы текущего профиля.
3. Выполняет `ANALYZE`.
4. Прогоняет одинаковый набор сценариев.

Так результаты меньше смешиваются с "остаточными" индексами от других экспериментов.

### Команда запуска

```bash
make index-joins-large
```

Эквивалент:

```bash
make INDEX_CONFIG=config.index.large.yaml ITERATIONS=20 WORKERS=4 index-joins-large
```

Для сверхтяжелого прогона можно указать исходный `config.large.yaml`:

```bash
make INDEX_CONFIG=config.large.yaml ITERATIONS=20 WORKERS=4 index-joins-large
```

Но `config.large.yaml` содержит `10M orders` и `50M order_items`, поэтому перебор всех index profiles на нем может быть многочасовым и требовать существенно больше диска.

### Входные данные

Фактически прогнанный large-профиль:

| Entity | Count |
|---|---:|
| Users | 200,000 |
| Orders | 1,000,000 |
| Order items | 3,000,000 |
| Products | 200,000 |
| Categories | 10,000 |

### Index Profiles

| Profile | Суть |
|---|---|
| `pk_only` | Только primary key/foreign key constraints, без вторичных индексов benchmark-а |
| `btree_fk_time` | B-tree индексы по FK и `orders.created_at` |
| `brin_time_btree_fk` | B-tree по FK + BRIN по `orders.created_at` |
| `hash_lookup_btree_fk` | Hash index для lookup по id + B-tree FK/time |
| `covering_composite` | Composite/covering индексы для join/range/aggregation путей |
| `mixed_full` | Смешанный полный набор: B-tree, hash, BRIN, covering, partial, expression |

### Сценарии

| Scenario | Описание |
|---|---|
| `join_hash` | 5-table join с принудительным Hash Join |
| `join_merge` | 5-table join с принудительным Merge Join |
| `join_nested_loop` | 5-table join с принудительным Nested Loop |
| `join_default` | 5-table join, выбор оптимизатора PostgreSQL |
| `point_lookup` | `SELECT * FROM orders WHERE id = 100000` |
| `range_scan_30d` | Заказы за последние 30 дней |
| `aggregation_30d` | Daily aggregation за последние 30 дней |

### Results

| Profile | Scenario | Requests | RPS | Avg | p50 | p95 | p99 | Max |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| `pk_only` | `join_hash` | 20 | 4.29 | 887.84 ms | 725.22 ms | 1.94 s | 2.24 s | 2.24 s |
| `pk_only` | `join_merge` | 20 | 1.98 | 1.91 s | 2.05 s | 2.77 s | 2.84 s | 2.84 s |
| `pk_only` | `join_nested_loop` | 20 | 1.38 | 2.80 s | 1.81 s | 7.12 s | 8.08 s | 8.08 s |
| `pk_only` | `join_default` | 20 | 4.35 | 785.48 ms | 526.60 ms | 1.82 s | 1.86 s | 1.86 s |
| `pk_only` | `point_lookup` | 20 | 3539.38 | 1.02 ms | 0.47 ms | 4.56 ms | 5.50 ms | 5.50 ms |
| `pk_only` | `range_scan_30d` | 20 | 15.66 | 229.92 ms | 215.43 ms | 387.86 ms | 389.74 ms | 389.74 ms |
| `pk_only` | `aggregation_30d` | 20 | 22.51 | 160.50 ms | 129.20 ms | 322.28 ms | 326.83 ms | 326.83 ms |
| `btree_fk_time` | `join_hash` | 20 | 3.84 | 889.93 ms | 566.17 ms | 2.13 s | 2.36 s | 2.36 s |
| `btree_fk_time` | `join_merge` | 20 | 2.75 | 1.44 s | 1.35 s | 1.94 s | 1.99 s | 1.99 s |
| `btree_fk_time` | `join_nested_loop` | 20 | 3.04 | 1.25 s | 983.68 ms | 2.80 s | 3.17 s | 3.17 s |
| `btree_fk_time` | `join_default` | 20 | 6.33 | 611.59 ms | 455.33 ms | 1.29 s | 1.54 s | 1.54 s |
| `btree_fk_time` | `point_lookup` | 20 | 2133.39 | 1.40 ms | 0.47 ms | 8.40 ms | 9.26 ms | 9.26 ms |
| `btree_fk_time` | `range_scan_30d` | 20 | 32.13 | 123.69 ms | 121.33 ms | 148.36 ms | 148.44 ms | 148.44 ms |
| `btree_fk_time` | `aggregation_30d` | 20 | 62.72 | 60.08 ms | 57.64 ms | 72.22 ms | 74.61 ms | 74.61 ms |
| `brin_time_btree_fk` | `join_hash` | 20 | 3.89 | 893.04 ms | 596.12 ms | 2.02 s | 2.15 s | 2.15 s |
| `brin_time_btree_fk` | `join_merge` | 20 | 1.59 | 2.36 s | 2.37 s | 2.93 s | 2.98 s | 2.98 s |
| `brin_time_btree_fk` | `join_nested_loop` | 20 | 3.03 | 1.29 s | 1.05 s | 2.70 s | 3.40 s | 3.40 s |
| `brin_time_btree_fk` | `join_default` | 20 | 4.21 | 929.51 ms | 695.41 ms | 2.15 s | 2.51 s | 2.51 s |
| `brin_time_btree_fk` | `point_lookup` | 20 | 2575.67 | 1.34 ms | 0.62 ms | 7.60 ms | 7.66 ms | 7.66 ms |
| `brin_time_btree_fk` | `range_scan_30d` | 20 | 12.78 | 299.80 ms | 262.82 ms | 434.31 ms | 561.02 ms | 561.02 ms |
| `brin_time_btree_fk` | `aggregation_30d` | 20 | 19.41 | 191.35 ms | 153.64 ms | 369.69 ms | 393.12 ms | 393.12 ms |
| `hash_lookup_btree_fk` | `join_hash` | 20 | 5.44 | 666.66 ms | 478.72 ms | 1.48 s | 1.48 s | 1.48 s |
| `hash_lookup_btree_fk` | `join_merge` | 20 | 3.40 | 1.12 s | 991.04 ms | 1.50 s | 1.57 s | 1.57 s |
| `hash_lookup_btree_fk` | `join_nested_loop` | 20 | 3.56 | 1.07 s | 795.36 ms | 2.13 s | 2.92 s | 2.92 s |
| `hash_lookup_btree_fk` | `join_default` | 20 | 6.17 | 620.75 ms | 468.94 ms | 1.38 s | 1.45 s | 1.45 s |
| `hash_lookup_btree_fk` | `point_lookup` | 20 | 2537.80 | 1.37 ms | 0.71 ms | 6.39 ms | 7.76 ms | 7.76 ms |
| `hash_lookup_btree_fk` | `range_scan_30d` | 20 | 29.44 | 135.37 ms | 133.67 ms | 144.51 ms | 144.81 ms | 144.81 ms |
| `hash_lookup_btree_fk` | `aggregation_30d` | 20 | 59.44 | 66.38 ms | 64.40 ms | 75.92 ms | 76.10 ms | 76.10 ms |
| `covering_composite` | `join_hash` | 20 | 5.11 | 714.67 ms | 477.94 ms | 1.48 s | 1.48 s | 1.48 s |
| `covering_composite` | `join_merge` | 20 | 3.72 | 1.03 s | 976.03 ms | 1.21 s | 1.31 s | 1.31 s |
| `covering_composite` | `join_nested_loop` | 20 | 4.37 | 879.79 ms | 653.78 ms | 1.91 s | 2.04 s | 2.04 s |
| `covering_composite` | `join_default` | 20 | 7.40 | 520.51 ms | 406.30 ms | 1.15 s | 1.18 s | 1.18 s |
| `covering_composite` | `point_lookup` | 20 | 2957.40 | 1.11 ms | 0.49 ms | 6.54 ms | 6.65 ms | 6.65 ms |
| `covering_composite` | `range_scan_30d` | 20 | 33.51 | 118.90 ms | 118.77 ms | 133.12 ms | 133.75 ms | 133.75 ms |
| `covering_composite` | `aggregation_30d` | 20 | 120.81 | 31.38 ms | 32.20 ms | 37.27 ms | 37.52 ms | 37.52 ms |
| `mixed_full` | `join_hash` | 20 | 4.73 | 723.45 ms | 496.70 ms | 1.62 s | 1.64 s | 1.64 s |
| `mixed_full` | `join_merge` | 20 | 3.09 | 1.22 s | 1.21 s | 1.52 s | 1.54 s | 1.54 s |
| `mixed_full` | `join_nested_loop` | 20 | 3.52 | 1.08 s | 1.03 s | 1.34 s | 1.34 s | 1.34 s |
| `mixed_full` | `join_default` | 20 | 4.72 | 723.48 ms | 474.09 ms | 1.75 s | 1.76 s | 1.76 s |
| `mixed_full` | `point_lookup` | 20 | 2601.20 | 1.19 ms | 0.43 ms | 6.90 ms | 7.60 ms | 7.60 ms |
| `mixed_full` | `range_scan_30d` | 20 | 29.91 | 131.55 ms | 130.99 ms | 142.23 ms | 142.45 ms | 142.45 ms |
| `mixed_full` | `aggregation_30d` | 20 | 149.10 | 26.19 ms | 24.24 ms | 33.58 ms | 34.34 ms | 34.34 ms |

### Top Results By Scenario

| Scenario | Best Profile | RPS | Avg | p95 |
|---|---|---:|---:|---:|
| `point_lookup` | `pk_only` | 3539.38 | 1.02 ms | 4.56 ms |
| `range_scan_30d` | `covering_composite` | 33.51 | 118.90 ms | 133.12 ms |
| `aggregation_30d` | `mixed_full` | 149.10 | 26.19 ms | 33.58 ms |
| `join_hash` | `hash_lookup_btree_fk` | 5.44 | 666.66 ms | 1.48 s |
| `join_merge` | `covering_composite` | 3.72 | 1.03 s | 1.21 s |
| `join_nested_loop` | `covering_composite` | 4.37 | 879.79 ms | 1.91 s |
| `join_default` | `covering_composite` | 7.40 | 520.51 ms | 1.15 s |

### Выводы

`covering_composite` оказался лучшим профилем для общего join workload-а: он выиграл `join_default`, `join_merge`, `join_nested_loop` и `range_scan_30d`.

`mixed_full` выиграл `aggregation_30d`, но не стал лучшим для join-ов. Это нормальный результат: "много индексов" не означает "лучший план". Дополнительные индексы расширяют пространство выбора оптимизатора, но могут приводить к менее удачному плану для конкретного запроса.

`pk_only` выиграл `point_lookup`, потому что primary key index уже существует, а дополнительных индексов нет. Это снижает общий шум и не мешает простому lookup-у. Hash index по `orders.id` не дал преимущества над primary key b-tree для этого point lookup.

BRIN по `created_at` в этом synthetic dataset-е проиграл B-tree/covering вариантам на `range_scan_30d` и `aggregation_30d`. Причина: сгенерированные timestamps распределены случайно, а BRIN эффективен, когда физический порядок таблицы коррелирует с indexed column. Для append-only time-series таблицы BRIN обычно выглядит лучше.

Для join-алгоритмов:

- `join_default` с `covering_composite` дал лучший общий join result: `7.40 RPS`.
- Принудительный `hash_join` лучше остальных принудительных вариантов в нескольких профилях, но имеет более тяжелые latency tails.
- `merge_join` и `nested_loop` сильно зависят от доступных индексов и формы данных; без подходящих индексов они быстро деградируют.

## 0A. Orders Read Benchmark: PostgreSQL Joins vs Elasticsearch Denormalized Reads

### Цель

Изначальный benchmark сравнивал чтение нормализованных order-данных в PostgreSQL через `JOIN` с чтением денормализованных order-документов в Elasticsearch.

PostgreSQL-сценарии проверяли:

- разные join-режимы: hash join, merge join, nested loop, default optimizer;
- point lookup по `orders.id`;
- range scan по `orders.created_at`;
- aggregation по дням.

Elasticsearch-сценарии проверяли:

- aggregation по стране пользователя;
- range query по `created_at`;
- term search по стране пользователя.

Важно: это не изолированный benchmark каждого PostgreSQL index type. Скрипт `postgres/init/benchmark_indexes.sql` создает набор b-tree/hash/brin/partial/expression indexes, а benchmark измеряет поведение read-сценариев поверх этой схемы. Сравнение "разных индексов один к одному" можно добавить отдельным этапом, где схема пересоздается с разными index profiles.

### Команда запуска

Финальный стабильный прогон выполнялся на medium-профиле:

```bash
make all CONFIG=config.medium.yaml ITERATIONS=50 WORKERS=4
make benchmark ITERATIONS=50 WORKERS=4
```

Второй `make benchmark` был выполнен после увеличения Docker shared memory для PostgreSQL:

```yaml
shm_size: "1gb"
```

Для report-grade large-прогона профиль уже есть:

```bash
make all CONFIG=config.large.yaml ITERATIONS=50 WORKERS=4
```

Но `config.large.yaml` содержит значительно более тяжелый датасет и в этой сессии полностью не прогонялся.

### Входные данные Финального Прогона

| Entity | Count |
|---|---:|
| Users | 100,000 |
| Orders | 500,000 |
| Order items | 1,500,000 |
| Products | 100,000 |
| Categories | 5,000 |

### Результаты

| # | Backend | Scenario | Requests | RPS | Avg | p50 | p95 | p99 | Max |
|---:|---|---|---:|---:|---:|---:|---:|---:|---:|
| 1 | PostgreSQL | `point_lookup` | 300 | 11691.04 | 0.34 ms | 0.24 ms | 0.57 ms | 1.91 ms | 8.23 ms |
| 2 | Elasticsearch | `elastic_agg` | 50 | 1265.82 | 3.07 ms | 2.21 ms | 11.60 ms | 11.73 ms | 11.73 ms |
| 3 | Elasticsearch | `elastic_search` | 50 | 982.95 | 4.05 ms | 3.80 ms | 5.84 ms | 6.35 ms | 6.35 ms |
| 4 | Elasticsearch | `elastic_range` | 50 | 580.54 | 6.83 ms | 6.71 ms | 8.58 ms | 10.31 ms | 10.31 ms |
| 5 | PostgreSQL | `aggregation` | 50 | 278.34 | 14.03 ms | 13.09 ms | 22.65 ms | 25.67 ms | 25.67 ms |
| 6 | PostgreSQL | `range_scan` | 150 | 49.74 | 78.94 ms | 57.72 ms | 174.62 ms | 181.49 ms | 190.28 ms |
| 7 | PostgreSQL | `hash_join` | 50 | 14.24 | 274.77 ms | 197.98 ms | 640.10 ms | 691.92 ms | 691.92 ms |
| 8 | PostgreSQL | `default_join` | 50 | 12.43 | 318.91 ms | 304.86 ms | 392.79 ms | 455.01 ms | 455.01 ms |
| 9 | PostgreSQL | `nested_loop` | 50 | 8.80 | 443.45 ms | 426.75 ms | 557.26 ms | 564.73 ms | 564.73 ms |
| 10 | PostgreSQL | `merge_join` | 50 | 8.75 | 446.86 ms | 420.56 ms | 572.18 ms | 595.41 ms | 595.41 ms |

### Top By Speed

| Место | Решение | Итог |
|---:|---|---|
| 1 | PostgreSQL point lookup | `11691.04 RPS`, avg `0.34 ms`, p95 `0.57 ms` |
| 2 | Elasticsearch aggregation | `1265.82 RPS`, avg `3.07 ms`, p95 `11.60 ms` |
| 3 | Elasticsearch term search | `982.95 RPS`, avg `4.05 ms`, p95 `5.84 ms` |
| 4 | Elasticsearch range query | `580.54 RPS`, avg `6.83 ms`, p95 `8.58 ms` |
| 5 | PostgreSQL aggregation | `278.34 RPS`, avg `14.03 ms`, p95 `22.65 ms` |

### Выводы

PostgreSQL ожидаемо является лучшим вариантом для точечного чтения по первичному ключу или хорошо селективному b-tree index lookup.

Elasticsearch выигрывает в денормализованных read-сценариях, где нужно быстро фильтровать/агрегировать заранее собранный документ.

Для сложных join-сценариев PostgreSQL заметно медленнее Elasticsearch-сценариев из-за необходимости собирать результат из нескольких нормализованных таблиц. Среди join-режимов в этом прогоне:

- `hash_join` дал лучший RPS среди принудительных join modes, но с тяжелым хвостом p95/p99;
- `default_join` был стабильнее по p95, но медленнее по RPS;
- `merge_join` и `nested_loop` в этом workload-е оказались самыми медленными.

## 1. Marketplace Read Model: PostgreSQL vs Elasticsearch

### Цель

Проверить сценарий, близкий к read model сервиса объявлений/маркетплейса: выдача страницы поиска и facets/aggregations по большому количеству фильтров.

Сравнивались два подхода:

- PostgreSQL: нормализованная модель с таблицами `listings`, `listing_sellers`, `listing_attribute_values`.
- Elasticsearch: денормализованный документ `listings`, где основные поля, seller и attributes лежат в одном search-документе.

### Команда запуска

```bash
make listing-large
```

Эквивалент:

```bash
make LISTING_CONFIG=config.listing.large.yaml ITERATIONS=50 WORKERS=4 listing
```

### Входные данные

| Entity | Count |
|---|---:|
| Listings | 1,000,000 |
| Sellers | 200,000 |
| Categories | 10,000 |
| Listing attributes | 5,000,000 |

### Сценарии

`listing_pg_search_page`:

- PostgreSQL-запрос страницы поиска.
- Join по `listings`, `listing_sellers` и нескольким `listing_attribute_values`.
- Фильтры: category range, city, price range, delivery, condition, seller rating, brand, color, material, текстовый фильтр.
- Сортировка: promoted, created_at.
- `LIMIT 50`.

`listing_pg_facets`:

- PostgreSQL facets/aggregation.
- Фильтры по категории, цене, delivery, seller rating, color.
- Group by city and brand.

`listing_es_search_page`:

- Elasticsearch search page по денормализованному документу.
- Те же типы фильтров: category, city, price, delivery, condition, seller rating, brand/color/material, text query.
- Сортировка: promoted, created_at.

`listing_es_facets`:

- Elasticsearch aggregation/facets.
- Фильтры по status, category, price, delivery, seller rating, color.
- Aggregations по city, brand, price percentiles.

### Результаты

| # | Backend | Scenario | Requests | RPS | Avg | p50 | p95 | p99 | Max |
|---:|---|---|---:|---:|---:|---:|---:|---:|---:|
| 1 | Elasticsearch | `listing_es_facets` | 50 | 693.06 | 5.62 ms | 3.48 ms | 28.85 ms | 28.99 ms | 28.99 ms |
| 2 | Elasticsearch | `listing_es_search_page` | 50 | 172.21 | 23.01 ms | 10.90 ms | 140.34 ms | 142.22 ms | 142.22 ms |
| 3 | PostgreSQL | `listing_pg_facets` | 50 | 14.54 | 266.19 ms | 245.28 ms | 394.61 ms | 410.18 ms | 410.18 ms |
| 4 | PostgreSQL | `listing_pg_search_page` | 50 | 7.93 | 490.24 ms | 441.77 ms | 645.16 ms | 653.31 ms | 653.31 ms |

### Top By Speed

| Место | Решение | Итог |
|---:|---|---|
| 1 | Elasticsearch facets | `693.06 RPS`, avg `5.62 ms`, p95 `28.85 ms` |
| 2 | Elasticsearch search page | `172.21 RPS`, avg `23.01 ms`, p95 `140.34 ms` |
| 3 | PostgreSQL facets | `14.54 RPS`, avg `266.19 ms`, p95 `394.61 ms` |
| 4 | PostgreSQL search page | `7.93 RPS`, avg `490.24 ms`, p95 `645.16 ms` |

### Выводы

Для поисково-фильтровой read model Elasticsearch существенно быстрее PostgreSQL в текущей постановке.

Основная причина не в том, что PostgreSQL "плохой для чтения", а в разной модели данных:

- PostgreSQL-сценарий читает нормализованную модель и вынужден делать join по продавцу и нескольким строкам атрибутов.
- Elasticsearch читает заранее собранный денормализованный документ.
- Фильтры, full-text query, сортировки и facets являются естественным workload-ом для Elasticsearch/Lucene.

Практический вывод для read model объявлений:

- PostgreSQL хорошо подходит для transactional source of truth, связей, консистентности, админских выборок.
- Elasticsearch/OpenSearch лучше подходит для публичного поиска, фильтров, facets, сортировки и выдачи search result pages.
- Для честной production-архитектуры обычно используют PostgreSQL как primary database, а Elasticsearch как asynchronous read/search projection.

## 2. Photo Storage: PostgreSQL BYTEA vs MinIO

### Цель

Понять, что оптимальнее для хранения и чтения фотографий: PostgreSQL `BYTEA` или object storage MinIO.

Сравнивались два подхода:

- PostgreSQL: бинарный payload хранится в таблице `photo_blobs.data BYTEA`.
- MinIO: тот же бинарный payload хранится как S3 object.

Метаданные в этом benchmark-е минимальные; цель была измерить именно хранение и random read бинарных данных.

### Команда запуска

```bash
make photos-large
```

Эквивалент:

```bash
make PHOTOS_CONFIG=config.photos.large.yaml ITERATIONS=200 WORKERS=8 photos
```

### Входные данные

| Параметр | Значение |
|---|---:|
| Photos | 20,000 |
| Avg photo size | 256 KB |
| Size jitter | 25% |
| Raw payload | ~5 GB |
| Random reads | 200 |
| Workers | 8 |

Payload генерировался детерминированно из `photo_id` и `seed`, поэтому PostgreSQL и MinIO получали одинаковые байты.

### Storage Size

| Storage | Size |
|---|---:|
| PostgreSQL total relation size | 5201.23 MB |
| PostgreSQL raw payload | 5000.11 MB |
| PostgreSQL table main fork | 1.78 MB |
| PostgreSQL indexes | 0.69 MB |
| MinIO objects total | 5000.11 MB |
| MinIO `/data` via `du -sb` | 5167.85 MB |

### Read Results

| # | Backend | Scenario | Requests | RPS | MB/s | Avg | p50 | p95 | p99 | Max |
|---:|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| 1 | MinIO | `photo_minio_random_read` | 200 | 1490.63 | 373.86 | 5.29 ms | 4.45 ms | 11.71 ms | 22.18 ms | 24.41 ms |
| 2 | PostgreSQL | `photo_pg_random_read` | 200 | 1028.15 | 257.87 | 7.70 ms | 6.07 ms | 19.36 ms | 49.36 ms | 61.58 ms |

### Top By Throughput

| Место | Решение | Итог |
|---:|---|---|
| 1 | MinIO random object read | `373.86 MB/s`, `1490.63 RPS`, avg `5.29 ms`, p95 `11.71 ms` |
| 2 | PostgreSQL BYTEA random read | `257.87 MB/s`, `1028.15 RPS`, avg `7.70 ms`, p95 `19.36 ms` |

### Выводы

На large-прогоне MinIO оказался быстрее PostgreSQL по random read бинарных payload-ов:

- примерно на `45%` выше throughput по MB/s;
- примерно на `45%` выше RPS;
- ниже avg latency и p95 latency.

По занимаемому месту в single-node контейнерном прогоне разница небольшая:

- PostgreSQL total relation size: `5201.23 MB`;
- MinIO фактический `/data`: `5167.85 MB`.

Но для production это не главный аргумент. Более важны эксплуатационные свойства.

Почему MinIO/object storage обычно предпочтительнее для фотографий:

- бинарные файлы не раздувают PostgreSQL heap/TOAST;
- меньше давление на WAL, backup, replication и vacuum;
- проще масштабировать storage независимо от базы;
- проще подключать CDN, lifecycle policies, object versioning, tiering;
- S3 API является стандартной интеграционной моделью для медиа;
- PostgreSQL остается сфокусирован на транзакционных данных и метаданных.

Где PostgreSQL `BYTEA` может быть оправдан:

- маленький объем файлов;
- нужно строго транзакционно записывать payload вместе с сущностью;
- нет отдельной object-storage инфраструктуры;
- файлы маленькие и редко читаются;
- важнее простота, чем масштабирование storage.

Практический вывод:

- Для production-сервиса объявлений/маркетплейса фотографии лучше хранить в MinIO/S3-compatible storage.
- В PostgreSQL лучше хранить metadata: photo id, listing id, object key, size, checksum, content type, sort order, moderation status.
- PostgreSQL `BYTEA` стоит рассматривать только для малых объемов или специальных транзакционных случаев.

## Overall Summary

| Area | Лучший вариант в benchmark-е | Причина |
|---|---|---|
| Search/listing read model | Elasticsearch | Денормализованные документы, inverted index, native filters/facets/search |
| Photo binary storage/read | MinIO | Object storage лучше подходит для больших бинарных payload-ов |
| Transactional source of truth | PostgreSQL | Консистентность, SQL, связи, транзакции |

Итоговая архитектурная рекомендация:

- PostgreSQL: source of truth, транзакционные сущности, metadata, связи.
- Elasticsearch/OpenSearch: search/read projection для объявлений.
- MinIO/S3: фотографии и другие большие бинарные объекты.
