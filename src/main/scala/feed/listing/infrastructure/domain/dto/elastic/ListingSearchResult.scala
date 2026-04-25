package feed.listing.infrastructure.domain.dto.elastic

import feed.listing.core.entity.Listing

final case class ListingSearchResult(listings: List[Listing], cursor: Option[String])
