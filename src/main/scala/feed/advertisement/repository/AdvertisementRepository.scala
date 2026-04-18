package feed.advertisement.repository

import zio._

import feed.advertisement.domain.entity
import feed.advertisement.domain.entity.AdvertisementError.PersistenceLayerError

trait AdvertisementRepository {
  def getAll: IO[PersistenceLayerError, List[entity.Advertisement]]
}
