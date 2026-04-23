package feed.listing.domain.model.postgres

import java.time.Instant
import java.util.UUID

import feed.listing.domain.types.ListingId

final case class ListingImage(
  id: UUID,
  listingId: ListingId,
  url: String,
  key: String,
  position: Int,
  createdAt: Instant)
