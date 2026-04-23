package feed.listing.repository

import zio.Chunk
import zio.IO

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError.PersistenceLayerError

trait ListingSearchEngine {
  def insertMany(listings: Chunk[entity.Listing]): IO[PersistenceLayerError, Unit]
}
