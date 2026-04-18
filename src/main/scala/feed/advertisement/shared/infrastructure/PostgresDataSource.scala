package feed.advertisement.shared.infrastructure

import doobie.hikari.HikariTransactor
import zio._
import zio.interop.catz._

object PostgresDataSource {

  val layer: ZLayer[Any, Throwable, HikariTransactor[Task]] = ZLayer.scoped {
    for {
      cfg <- ZIO.config[DbConfig]

      connectEC <- ZIO.executor.map(_.asExecutionContext)

      xa <- HikariTransactor
        .newHikariTransactor[Task](
          driverClassName = "org.postgresql.Driver",
          url = cfg.url,
          user = cfg.user,
          pass = cfg.password,
          connectEC = connectEC,
          logHandler = None
        )
        .toScopedZIO
    } yield xa
  }
}
