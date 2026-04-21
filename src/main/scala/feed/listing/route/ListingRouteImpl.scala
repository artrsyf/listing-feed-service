package feed.listing.route

import sttp.tapir.json.zio._
import sttp.tapir.ztapir._
import zio._

import feed.listing.delivery.ListingHandler
import feed.listing.domain.dto
import feed.listing.domain.types.ListingId
import feed.listing.shared.apierror.ApiError.errorMapper

final class ListingRouteImpl(listingHandler: ListingHandler) extends ListingRoute {

  private val baseEndpoint =
    endpoint
      .tag("some tag")
      .in("api" / "feed")

  private val getRecentListings =
    baseEndpoint.get
      .in("listings")
      .out(jsonBody[dto.GetAllListingsResponse])
      .errorOut(errorMapper)
      .zServerLogic(_ => listingHandler.getRecentListings)

  private val getListing =
    baseEndpoint.get
      .in("listing" / path[ListingId]("listing_id"))
      .out(jsonBody[dto.ListingResponse])
      .errorOut(errorMapper)
      .zServerLogic(listingId => listingHandler.getListing(listingId))

  private val createListing =
    baseEndpoint.post
      .in("listings")
      .in(jsonBody[dto.CreateListingRequest])
      .out(jsonBody[dto.CreateListingResponse])
      .errorOut(errorMapper)
      .zServerLogic(req => listingHandler.createListing(req))

  override val routes =
    Chunk(getRecentListings, getListing, createListing)
}

object ListingRouteImpl {
  val layer: ZLayer[ListingHandler, Nothing, ListingRoute] = ZLayer
    .derive[ListingRouteImpl]
}
