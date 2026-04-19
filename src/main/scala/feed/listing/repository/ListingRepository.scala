package feed.listing.repository

import zio._

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError.PersistenceLayerError

trait ListingRepository {
  def getRecentListings(limit: Int): IO[PersistenceLayerError, List[entity.Listing]]
  def create(l: entity.Listing): IO[PersistenceLayerError, Unit]
}
