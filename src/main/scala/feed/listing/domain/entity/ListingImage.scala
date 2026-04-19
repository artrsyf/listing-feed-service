package feed.listing.domain.entity

import java.util.UUID

final case class ListingImage(
  id: UUID,
  url: String,
  position: Int)
