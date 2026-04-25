package feed

import sttp.tapir.ztapir.ZServerEndpoint
import zio.Chunk
import zio.ZLayer

import feed.listing.infrastructure.route.ListingRoute

trait RouteAggregator {
  def allRoutes: Chunk[ZServerEndpoint[Any, Any]]
}

final class RouteAggregatorImpl(listingRoute: ListingRoute) extends RouteAggregator {
  override def allRoutes: Chunk[ZServerEndpoint[Any, Any]] =
    listingRoute.routes
}

object RouteAggregatorImpl {
  val layer: ZLayer[ListingRoute, Nothing, RouteAggregator] = ZLayer
    .derive[RouteAggregatorImpl]
}
