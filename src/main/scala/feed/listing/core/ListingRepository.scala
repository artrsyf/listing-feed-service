package feed.listing.core

import java.time.Instant

import zio._

import feed.listing.core.entity.ListingError.PersistenceLayerError
import feed.listing.core.entity.ListingId

trait ListingRepository {
  @deprecated("Модель чтения идет через ES", "25-04-2026")
  def getRecentListings(
      cursor: Option[Instant],
      limit: Int
  ): IO[PersistenceLayerError, List[entity.Listing]]
  def getById(listingId: ListingId): IO[PersistenceLayerError, Option[entity.Listing]]
  def create(l: entity.Listing): IO[PersistenceLayerError, Unit]
}
