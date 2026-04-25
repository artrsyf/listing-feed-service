package feed.listing.infrastructure.route

import java.time.Instant

import sttp.tapir.json.zio._
import sttp.tapir.ztapir._
import zio._

import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.delivery.ListingHandler
import feed.listing.infrastructure.domain.dto.http.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.shared.apierror.ApiError.errorMapper

final class ListingRouteImpl(listingHandler: ListingHandler) extends ListingRoute {

  private val baseEndpoint =
    endpoint
      .tag("some tag")
      .in("api" / "feed")

  private val getRecentListings =
    baseEndpoint.get
      .in("listings")
      .in(query[Option[Instant]]("cursor"))
      .in(query[Option[Int]]("limit"))
      .out(jsonBody[GetAllListingsResponse])
      .errorOut(errorMapper)
      .zServerLogic { case (cursor, limit) => listingHandler.getRecentListings(cursor, limit) }

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

  override val routes =
    Chunk(getRecentListings, getListing, createListing)
}

object ListingRouteImpl {
  val layer: ZLayer[ListingHandler, Nothing, ListingRoute] = ZLayer
    .derive[ListingRouteImpl]
}
