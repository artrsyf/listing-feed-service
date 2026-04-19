package feed.listing.usecase

import java.util.UUID

import zio._

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError
import feed.listing.repository.ListingRepository

final class ListingService(
  listingRepo: ListingRepository,
  listingConfig: ListingConfig) {
  def getRecentListings: IO[ListingError, List[entity.Listing]] =
    listingRepo.getRecentListings(listingConfig.limit)

  def createListing(listing: entity.Listing): IO[ListingError, UUID] =
    for {
      id  <- ZIO.succeed(java.util.UUID.randomUUID())
      now <- Clock.instant
      _   <- listingRepo.create(listing)
    } yield id
}

object ListingService {
  val layer: RLayer[ListingRepository, ListingService] = ZLayer
    .derive[ListingService]
}
