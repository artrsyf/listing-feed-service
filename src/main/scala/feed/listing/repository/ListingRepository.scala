package feed.listing.repository

import java.time.Instant

import zio._

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError.PersistenceLayerError
import feed.listing.domain.types.ListingId

trait ListingRepository {
  def getRecentListings(
    cursor: Option[Instant],
    limit: Int
  ): IO[PersistenceLayerError, List[entity.Listing]]
  def getById(listingId: ListingId): IO[PersistenceLayerError, Option[entity.Listing]]
  def create(l: entity.Listing): IO[PersistenceLayerError, Unit]
}
