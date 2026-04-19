package feed.listing.delivery

import zio._

import feed.listing.domain.dto
import feed.listing.shared.apierror.ApiError

trait ListingHandler {
  def getRecentListings: IO[ApiError, dto.GetAllListingsResponse]
  def createListing(req: dto.CreateListingRequest): IO[ApiError, dto.CreateListingResponse]
}
