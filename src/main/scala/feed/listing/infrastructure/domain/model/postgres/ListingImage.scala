package feed.listing.infrastructure.domain.model.postgres

import java.time.Instant
import java.util.UUID

import feed.listing.core.entity.ListingId

final case class ListingImage(
    id: UUID,
    listingId: ListingId,
    key: String,
    position: Int,
    createdAt: Instant
)
