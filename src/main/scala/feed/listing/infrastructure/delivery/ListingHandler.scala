package feed.listing.infrastructure.delivery

import java.time.Instant

import zio._

import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.listing.infrastructure.domain.dto.http.createlisting.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.createlisting.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.generateuploadurl.GenerateUploadUrlResponse
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsRequest
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsResponse
import feed.shared.apierror.ApiError

trait ListingHandler {
  def getRecentListings(
      cursor: Option[String],
      limit: Option[Int]
  ): IO[ApiError, GetAllListingsResponse]

  def searchListings(req: SearchListingsRequest): IO[ApiError, SearchListingsResponse]

  def getListing(listingId: ListingId): IO[ApiError, ListingResponse]

  def createListing(req: CreateListingRequest): IO[ApiError, CreateListingResponse]

  def generateUploadUrl: IO[ApiError, GenerateUploadUrlResponse]
}
