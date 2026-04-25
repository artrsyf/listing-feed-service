package feed.listing.infrastructure.domain.model.postgres

import java.time.Instant

import feed.listing.core.entity.ListingId

final case class Listing(
  id: ListingId,
  title: String,
  description: String,
  price: BigDecimal,
  currency: String,
  status: ListingStatus,
  createdAt: Instant,
  updatedAt: Instant)
