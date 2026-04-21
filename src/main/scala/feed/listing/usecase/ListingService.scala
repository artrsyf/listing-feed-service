package feed.listing.usecase

import java.util.UUID

import zio._

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError
import feed.listing.domain.types.ListingId
import feed.listing.repository.ListingRepository

final class ListingService(
  listingRepo: ListingRepository,
  listingConfig: ListingConfig) {
  def getRecentListings: IO[ListingError, List[entity.Listing]] =
    listingRepo.getRecentListings(listingConfig.limit)

  def getListing(listingId: ListingId): IO[ListingError, entity.Listing] =
    listingRepo.getById(listingId).someOrFail(ListingError.Notfound)

  def createListing(listing: entity.Listing): IO[ListingError, UUID] =
    listingRepo.create(listing).as(listing.id)
}

object ListingService {
  val layer: RLayer[ListingRepository, ListingService] = ZLayer
    .derive[ListingService]
}
