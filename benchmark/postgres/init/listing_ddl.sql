DROP TABLE IF EXISTS listing_attribute_values CASCADE;
DROP TABLE IF EXISTS listings CASCADE;
DROP TABLE IF EXISTS listing_sellers CASCADE;
DROP TABLE IF EXISTS listing_categories CASCADE;

CREATE TABLE listing_categories (
    id          BIGINT PRIMARY KEY,
    parent_id   BIGINT NULL REFERENCES listing_categories(id),
    name        TEXT NOT NULL
);

CREATE TABLE listing_sellers (
    id          BIGINT PRIMARY KEY,
    city        TEXT NOT NULL,
    rating      NUMERIC(3,2) NOT NULL,
    created_at  TIMESTAMP NOT NULL
);

CREATE TABLE listings (
    id           BIGINT PRIMARY KEY,
    seller_id    BIGINT NOT NULL REFERENCES listing_sellers(id),
    category_id  BIGINT NOT NULL REFERENCES listing_categories(id),
    city         TEXT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL,
    price        NUMERIC(14,2) NOT NULL,
    condition    SMALLINT NOT NULL,
    delivery     BOOLEAN NOT NULL,
    promoted     BOOLEAN NOT NULL,
    status       SMALLINT NOT NULL,
    created_at   TIMESTAMP NOT NULL,
    updated_at   TIMESTAMP NOT NULL
);

CREATE TABLE listing_attribute_values (
    listing_id  BIGINT NOT NULL REFERENCES listings(id),
    attr_key    TEXT NOT NULL,
    attr_value  TEXT NOT NULL,
    PRIMARY KEY (listing_id, attr_key)
);

CREATE INDEX idx_listings_search_core
    ON listings (category_id, city, status, price, created_at DESC);

CREATE INDEX idx_listings_city_category_price
    ON listings (city, category_id, price);

CREATE INDEX idx_listings_created_at
    ON listings (created_at DESC);

CREATE INDEX idx_listings_seller
    ON listings (seller_id);

CREATE INDEX idx_listing_attrs_key_value_listing
    ON listing_attribute_values (attr_key, attr_value, listing_id);

CREATE INDEX idx_listing_sellers_city_rating
    ON listing_sellers (city, rating);

ANALYZE;
