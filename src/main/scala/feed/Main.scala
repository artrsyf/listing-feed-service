package feed

import sttp.tapir.server.ziohttp._
import zio._
import zio.config.typesafe._
import zio.http.Server

import feed.listing.delivery.ListingHttpHandler
import feed.listing.repository.ListingPostgresRepository
import feed.listing.route.ListingRouteImpl
import feed.listing.shared.infrastructure.DbConfig
import feed.listing.shared.infrastructure.PostgresDataSource
import feed.listing.usecase.ListingService

object Main extends ZIOAppDefault {
  override val bootstrap: ZLayer[ZIOAppArgs, Any, Any] =
    Runtime
      .setConfigProvider(
        ConfigProvider
          .fromResourcePath()
      )

  override def run =
    (for {
      aggregator <- ZIO
        .service[RouteAggregator]

      app =
        ZioHttpInterpreter()
          .toHttp(aggregator.allRoutes.toList)

      _ <- Server.serve(app)
    } yield ()).provide(
      Server
        .defaultWithPort(8080),
      Runtime.setConfigProvider(
        TypesafeConfigProvider
          .fromResourcePath()
      ),
      PostgresDataSource.layer,
      ListingPostgresRepository.layer,
      ListingService.layer,
      ListingHttpHandler.layer,
      ListingRouteImpl.layer,
      RouteAggregatorImpl.layer
    )
}
