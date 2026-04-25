package feed.listing.core

import zio.Chunk
import zio.IO

import feed.listing.core.entity.ListingError.PersistenceLayerError
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchCriteria
import feed.listing.infrastructure.domain.dto.elastic.PageRequest
import feed.listing.infrastructure.domain.dto.http.SearchListingsRequest

trait ListingSearchIndexEngine {
  def insertMany(listings: Chunk[entity.Listing]): IO[PersistenceLayerError, Unit]
}
