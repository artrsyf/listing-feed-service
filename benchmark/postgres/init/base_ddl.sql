CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =========================================================
-- USERS
-- =========================================================

CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,

    email           TEXT NOT NULL,
    country         VARCHAR(2) NOT NULL,

    status          SMALLINT NOT NULL,

    created_at      TIMESTAMP NOT NULL,

    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- =========================================================
-- CATEGORIES
-- =========================================================

CREATE TABLE categories (
    id              BIGSERIAL PRIMARY KEY,

    parent_id       BIGINT NULL,

    name            TEXT NOT NULL,

    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_categories_parent
        FOREIGN KEY(parent_id)
        REFERENCES categories(id)
);

-- =========================================================
-- PRODUCTS
-- =========================================================

CREATE TABLE products (
    id              BIGSERIAL PRIMARY KEY,

    category_id     BIGINT NOT NULL,

    sku             TEXT NOT NULL,
    name            TEXT NOT NULL,

    price           NUMERIC(12,2) NOT NULL,

    rating          DOUBLE PRECISION NOT NULL,

    is_active       BOOLEAN NOT NULL DEFAULT TRUE,

    created_at      TIMESTAMP NOT NULL,

    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_products_category
        FOREIGN KEY(category_id)
        REFERENCES categories(id)
);

-- =========================================================
-- ORDERS
-- =========================================================

CREATE TABLE orders (
    id              BIGSERIAL PRIMARY KEY,

    user_id         BIGINT NOT NULL,

    status          SMALLINT NOT NULL,

    total_amount    NUMERIC(14,2) NOT NULL,

    created_at      TIMESTAMP NOT NULL,

    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_orders_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
);

-- =========================================================
-- ORDER ITEMS
-- =========================================================

CREATE TABLE order_items (
    id              BIGSERIAL PRIMARY KEY,

    order_id        BIGINT NOT NULL,

    product_id      BIGINT NOT NULL,

    quantity        INTEGER NOT NULL,

    price           NUMERIC(12,2) NOT NULL,

    created_at      TIMESTAMP NOT NULL,

    CONSTRAINT fk_order_items_order
        FOREIGN KEY(order_id)
        REFERENCES orders(id),

    CONSTRAINT fk_order_items_product
        FOREIGN KEY(product_id)
        REFERENCES products(id)
);

-- =========================================================
-- BASELINE INDEXES
-- =========================================================

-- USERS

CREATE INDEX idx_users_country
    ON users(country);

CREATE INDEX idx_users_created_at
    ON users(created_at);

CREATE INDEX idx_users_status
    ON users(status);

-- PRODUCTS

CREATE INDEX idx_products_category_id
    ON products(category_id);

CREATE INDEX idx_products_price
    ON products(price);

CREATE INDEX idx_products_rating
    ON products(rating);

CREATE UNIQUE INDEX idx_products_sku
    ON products(sku);

-- ORDERS

CREATE INDEX idx_orders_user_id
    ON orders(user_id);

CREATE INDEX idx_orders_created_at
    ON orders(created_at);

CREATE INDEX idx_orders_status
    ON orders(status);

CREATE INDEX idx_orders_user_created
    ON orders(user_id, created_at);

CREATE INDEX idx_orders_created_status
    ON orders(created_at, status);

-- ORDER ITEMS

CREATE INDEX idx_order_items_order_id
    ON order_items(order_id);

CREATE INDEX idx_order_items_product_id
    ON order_items(product_id);

CREATE INDEX idx_order_items_order_product
    ON order_items(order_id, product_id);

-- =========================================================
-- ANALYZE
-- =========================================================

ANALYZE;