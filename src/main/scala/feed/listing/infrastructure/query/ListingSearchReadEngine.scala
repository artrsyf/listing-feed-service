package feed.listing.infrastructure.query

import zio.IO

import feed.listing.core.entity.Listing
import feed.listing.core.entity.ListingError.PersistenceLayerError
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchCriteria
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchResult
import feed.listing.infrastructure.domain.dto.elastic.PageRequest

trait ListingSearchReadEngine {
  def searchListings(
      criteria: ListingSearchCriteria,
      page: PageRequest
  ): IO[PersistenceLayerError, ListingSearchResult]

  def getById(id: ListingId): IO[PersistenceLayerError, Option[Listing]]
}
