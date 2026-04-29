package feed.listing.infrastructure.route

import java.time.Instant

import sttp.tapir.json.zio._
import sttp.tapir.ztapir._
import zio._

import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.delivery.ListingHandler
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.listing.infrastructure.domain.dto.http.createlisting.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.createlisting.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.generateuploadurl.GenerateUploadUrlRequest
import feed.listing.infrastructure.domain.dto.http.generateuploadurl.GenerateUploadUrlResponse
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsRequest
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsResponse
import feed.shared.apierror.ApiError.errorMapper

final class ListingRouteImpl(listingHandler: ListingHandler) extends ListingRoute {

  private val baseEndpoint =
    endpoint
      .tag("some tag")
      .in("api" / "feed")

  private val getRecentListings =
    baseEndpoint.get
      .in("listings")
      .in(query[Option[String]]("cursor"))
      .in(query[Option[Int]]("limit"))
      .out(jsonBody[GetAllListingsResponse])
      .errorOut(errorMapper)
      .zServerLogic { case (cursor, limit) => listingHandler.getRecentListings(cursor, limit) }

  private val searchListings =
    baseEndpoint.get
      .in("search")
      .in(query[Option[String]]("q"))
      .in(query[Option[String]]("cursor"))
      .in(query[Int]("limit").default(20))
      .in(query[Option[BigDecimal]]("minPrice"))
      .in(query[Option[BigDecimal]]("maxPrice"))
      .out(jsonBody[SearchListingsResponse])
      .errorOut(errorMapper)
      .zServerLogic { case (q, cursor, limit, minPrice, maxPrice) =>
        listingHandler.searchListings(SearchListingsRequest(q, cursor, limit, minPrice, maxPrice))
      }

  private val getListing =
    baseEndpoint.get
      .in("listing" / path[ListingId]("listing_id"))
      .out(jsonBody[ListingResponse])
      .errorOut(errorMapper)
      .zServerLogic(listingId => listingHandler.getListing(listingId))

  private val createListing =
    baseEndpoint.post
      .in("listings")
      .in(jsonBody[CreateListingRequest])
      .out(jsonBody[CreateListingResponse])
      .errorOut(errorMapper)
      .zServerLogic(req => listingHandler.createListing(req))

  private val generateUploadUrl =
    baseEndpoint.post
      .in("images" / "upload-url")
      .in(jsonBody[GenerateUploadUrlRequest])
      .out(jsonBody[GenerateUploadUrlResponse])
      .errorOut(errorMapper)
      .zServerLogic(_ => listingHandler.generateUploadUrl)

  override val routes =
    Chunk(getRecentListings, getListing, searchListings, createListing, generateUploadUrl)
}

object ListingRouteImpl {
  val layer: ZLayer[ListingHandler, Nothing, ListingRoute] = ZLayer
    .derive[ListingRouteImpl]
}
