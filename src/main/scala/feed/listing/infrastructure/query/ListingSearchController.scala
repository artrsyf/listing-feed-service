package feed.listing.infrastructure.query

import io.scalaland.chimney.dsl._
import zio._

import feed.listing.core.entity.ListingError
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchCriteria
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchResult
import feed.listing.infrastructure.domain.dto.elastic.PageRequest
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.listing.infrastructure.domain.dto.http.SearchListingsResponse

final class ListingSearchController(listingSearchEngine: ListingSearchReadEngine) {
  def search(
      criteria: ListingSearchCriteria,
      page: PageRequest
  ): IO[ListingError, ListingSearchResult] =
    listingSearchEngine
      .searchListings(criteria, page)
}

object ListingSearchController {
  val layer: RLayer[ListingSearchReadEngine, ListingSearchController] =
    ZLayer.derive[ListingSearchController]
}
