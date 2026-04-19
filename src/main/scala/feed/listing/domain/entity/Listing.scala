package feed.listing.domain.entity

import java.time.Instant

import feed.listing.domain.types.ListingId

final case class Listing(
  id: ListingId,
  title: String,
  description: String,
  price: BigDecimal,
  currency: String,
  status: ListingStatus,
  images: List[ListingImage],
  createdAt: Instant,
  updatedAt: Instant)
