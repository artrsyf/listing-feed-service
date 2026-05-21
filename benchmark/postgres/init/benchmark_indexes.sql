-- =========================================================
-- BENCHMARK INDEXES FOR POSTGRESQL
-- =========================================================
-- This script creates various index types for benchmark comparison
-- Run after base_ddl.sql and data loading

-- =========================================================
-- HASH INDEXES (for equality comparisons)
-- =========================================================

-- Hash index on orders.id for point lookups
CREATE INDEX IF NOT EXISTS idx_orders_id_hash
    ON orders USING HASH (id);

-- Hash index on users.id
CREATE INDEX IF NOT EXISTS idx_users_id_hash
    ON users USING HASH (id);

-- =========================================================
-- COMPOSITE INDEXES FOR JOIN OPTIMIZATION
-- =========================================================

-- Covering index for orders with commonly accessed columns
CREATE INDEX IF NOT EXISTS idx_orders_created_user_covering
    ON orders (created_at, user_id) INCLUDE (total_amount, status);

-- Composite index for order_items join optimization
CREATE INDEX IF NOT EXISTS idx_order_items_order_includes
    ON order_items (order_id) INCLUDE (product_id, quantity, price);

-- =========================================================
-- BRIN INDEXES (for time-series range queries)
-- =========================================================

-- BRIN index for orders.created_at (efficient for time range queries)
CREATE INDEX IF NOT EXISTS idx_orders_created_at_brin
    ON orders USING BRIN (created_at);

-- BRIN index for order_items.created_at
CREATE INDEX IF NOT EXISTS idx_order_items_created_at_brin
    ON order_items USING BRIN (created_at);

-- =========================================================
-- GIN INDEXES (for array/JSON queries)
-- =========================================================

-- Note: GIN indexes are more relevant for JSONB columns
-- Example if you add a JSONB column:
-- ALTER TABLE orders ADD COLUMN metadata JSONB;
-- CREATE INDEX idx_orders_metadata_gin ON orders USING GIN (metadata);

-- =========================================================
-- GiST INDEXES (for geospatial/full-text search)
-- =========================================================

-- Example for future geospatial queries:
-- ALTER TABLE users ADD COLUMN location GEOGRAPHY(POINT);
-- CREATE INDEX idx_users_location_gist ON users USING GIST (location);

-- Full-text search example:
-- ALTER TABLE products ADD COLUMN search_vector tsvector;
-- UPDATE products SET search_vector = to_tsvector('english', name);
-- CREATE INDEX idx_products_search_gist ON products USING GIST (search_vector);

-- =========================================================
-- PARTIAL INDEXES (for filtered queries)
-- =========================================================

-- Partial index for recent orders (last 90 days)
CREATE INDEX IF NOT EXISTS idx_orders_recent
    ON orders (created_at) 
    WHERE created_at > NOW() - interval '90 days';

-- Partial index for active products
CREATE INDEX IF NOT EXISTS idx_products_active
    ON products (id, category_id, price)
    WHERE is_active = true;

-- =========================================================
-- EXPRESSION INDEXES
-- =========================================================

-- Index on date truncation for aggregation queries
CREATE INDEX IF NOT EXISTS idx_orders_date_trunc_day
    ON orders (DATE_TRUNC('day', created_at));

-- Index on country for grouping (already exists as B-Tree)
-- This is for demonstration of expression indexes
CREATE INDEX IF NOT EXISTS idx_users_country_upper
    ON users (UPPER(country));

-- =========================================================
-- UPDATE STATISTICS
-- =========================================================

ANALYZE;

-- =========================================================
-- INDEX SIZE REPORT
-- =========================================================

SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;
