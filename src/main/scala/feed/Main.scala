package feed

import sttp.tapir.server.ziohttp._
import zio._
import zio.http.Server

import feed.advertisement.delivery.AdvertisementHttpHandler
import feed.advertisement.repository.AdvertisementPostgresRepository
import feed.advertisement.route.AdvertisementRouteImpl
import feed.advertisement.shared.infrastructure.PostgresDataSource
import feed.advertisement.usecase.AdvertisementService

object Main extends ZIOAppDefault {

  override def run = (for {
    aggregator <- ZIO.service[RouteAggregator]

    app = ZioHttpInterpreter().toHttp(aggregator.allRoutes.toList)

    _ <- Server.serve(app)
  } yield ()).provide(
    Server.defaultWithPort(8080),
    PostgresDataSource.layer,
    AdvertisementPostgresRepository.layer,
    AdvertisementService.layer,
    AdvertisementHttpHandler.layer,
    AdvertisementRouteImpl.layer,
    RouteAggregatorImpl.layer
  )
}
