package feed.listing.infrastructure.delivery

import java.time.Instant

import zio._

import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.http.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.listing.infrastructure.domain.dto.http.SearchListingsRequest
import feed.listing.infrastructure.domain.dto.http.SearchListingsResponse
import feed.shared.apierror.ApiError

trait ListingHandler {
  def getRecentListings(
      cursor: Option[String],
      limit: Option[Int]
  ): IO[ApiError, GetAllListingsResponse]
  def searchListings(req: SearchListingsRequest): IO[ApiError, SearchListingsResponse]
  def getListing(listingId: ListingId): IO[ApiError, ListingResponse]
  def createListing(req: CreateListingRequest): IO[ApiError, CreateListingResponse]
}
