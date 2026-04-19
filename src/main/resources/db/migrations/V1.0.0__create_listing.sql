CREATE TYPE listing_status AS ENUM (
  'ACTIVE',
  'SOLD',
  'DRAFT',
  'DELETED'
);

CREATE TABLE listings (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  price NUMERIC(12,2) NOT NULL,
  currency TEXT NOT NULL,
  status listing_status NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_listings_created_at ON listings(created_at DESC);

CREATE TABLE listing_images (
  id UUID PRIMARY KEY,
  listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  key TEXT NOT NULL,
  position INT NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_listing_images_listing_id ON listing_images(listing_id);