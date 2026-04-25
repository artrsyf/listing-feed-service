package feed.listing.core

import zio.Chunk
import zio.IO

import feed.listing.core.entity.ListingError.PersistenceLayerError

trait ListingSearchEngine {
  def insertMany(listings: Chunk[entity.Listing]): IO[PersistenceLayerError, Unit]
}
