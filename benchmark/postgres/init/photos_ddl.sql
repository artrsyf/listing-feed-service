DROP TABLE IF EXISTS photo_blobs CASCADE;

CREATE TABLE photo_blobs (
    id          BIGINT PRIMARY KEY,
    listing_id  BIGINT NOT NULL,
    content_type TEXT NOT NULL,
    data        BYTEA NOT NULL,
    size_bytes  BIGINT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_photo_blobs_listing_id
    ON photo_blobs(listing_id);

ANALYZE;
