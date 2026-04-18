package feed.advertisement.repository

import doobie._
import doobie.hikari.HikariTransactor
import doobie.implicits._
import io.scalaland.chimney.dsl._
import zio._
import zio.interop.catz._

import feed.advertisement.domain.entity
import feed.advertisement.domain.entity.AdvertisementError.PersistenceLayerError
import feed.advertisement.domain.model.Advertisement

class AdvertisementPostgresRepository(xa: Transactor[Task]) extends AdvertisementRepository {
  def getAll: IO[PersistenceLayerError, List[entity.Advertisement]] =
    sql"SELECT id FROM advertisement"
      .query[Advertisement]
      .to[List]
      .transact(xa)
      .mapBoth(
        e => PersistenceLayerError(e.getMessage),
        advertisements => advertisements.map(_.transformInto[entity.Advertisement])
      )
}

object AdvertisementPostgresRepository {
  val layer: ZLayer[HikariTransactor[Task], Nothing, AdvertisementRepository] = ZLayer
    .fromFunction(new AdvertisementPostgresRepository(_))
}
