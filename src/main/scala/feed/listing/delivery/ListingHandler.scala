package feed.listing.delivery

import java.time.Instant

import zio._

import feed.listing.domain.dto
import feed.listing.domain.types.ListingId
import feed.listing.shared.apierror.ApiError

trait ListingHandler {
  def getRecentListings(
    cursor: Option[Instant],
    limit: Option[Int]
  ): IO[ApiError, dto.GetAllListingsResponse]
  def getListing(listingId: ListingId): IO[ApiError, dto.ListingResponse]
  def createListing(req: dto.CreateListingRequest): IO[ApiError, dto.CreateListingResponse]
}
