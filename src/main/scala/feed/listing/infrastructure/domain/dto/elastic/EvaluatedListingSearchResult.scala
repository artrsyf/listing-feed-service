package feed.listing.infrastructure.domain.dto.elastic

import java.time.Instant
import java.util.UUID

import feed.listing.core.entity.Listing
import feed.listing.core.entity.ListingId
import feed.listing.core.entity.ListingStatus

final case class EvaluatedListingSearchResult(
    listings: List[EvaluatedListing],
    cursor: Option[String]
)

final case class EvaluatedListing(
    id: ListingId,
    title: String,
    description: String,
    price: BigDecimal,
    currency: String,
    status: ListingStatus,
    images: List[EvaluatedListingImage],
    createdAt: Instant,
    updatedAt: Instant
)

final case class EvaluatedListingImage(url: String, position: Int)
