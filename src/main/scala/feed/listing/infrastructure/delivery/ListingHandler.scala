package feed.listing.infrastructure.delivery

import java.time.Instant

import zio._

import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.http.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.shared.apierror.ApiError

trait ListingHandler {
  def getRecentListings(
    cursor: Option[Instant],
    limit: Option[Int]
  ): IO[ApiError, GetAllListingsResponse]
  def getListing(listingId: ListingId): IO[ApiError, ListingResponse]
  def createListing(req: CreateListingRequest): IO[ApiError, CreateListingResponse]
}
