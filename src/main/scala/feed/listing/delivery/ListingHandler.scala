package feed.listing.delivery

import java.time.Instant

import zio._

import feed.listing.domain.dto.http.CreateListingRequest
import feed.listing.domain.dto.http.CreateListingResponse
import feed.listing.domain.dto.http.GetAllListingsResponse
import feed.listing.domain.dto.http.ListingResponse
import feed.listing.domain.entity.ListingId
import feed.listing.shared.apierror.ApiError

trait ListingHandler {
  def getRecentListings(
    cursor: Option[Instant],
    limit: Option[Int]
  ): IO[ApiError, GetAllListingsResponse]
  def getListing(listingId: ListingId): IO[ApiError, ListingResponse]
  def createListing(req: CreateListingRequest): IO[ApiError, CreateListingResponse]
}
