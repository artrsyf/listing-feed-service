package feed.listing.core.entity

import java.time.Instant

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
