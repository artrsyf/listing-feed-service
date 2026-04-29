package feed.listing.core

import zio.Chunk
import zio.IO

import feed.listing.core.entity.ListingError.PersistenceLayerError

trait ListingSearchCreateEngine {
  def insertMany(listings: Chunk[entity.Listing]): IO[PersistenceLayerError, Unit]
}
