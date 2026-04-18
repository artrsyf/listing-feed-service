package feed

import sttp.tapir.ztapir.ZServerEndpoint
import zio.Chunk
import zio.ZLayer

import feed.advertisement.route.AdvertisementRoute

trait RouteAggregator {
  def allRoutes: Chunk[ZServerEndpoint[Any, Any]]
}

final class RouteAggregatorImpl(advertisementRoute: AdvertisementRoute) extends RouteAggregator {
  override def allRoutes: Chunk[ZServerEndpoint[Any, Any]] = advertisementRoute.routes
}

object RouteAggregatorImpl {
  val layer: ZLayer[AdvertisementRoute, Nothing, RouteAggregator] = ZLayer
    .derive[RouteAggregatorImpl]
}
