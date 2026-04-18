package feed.advertisement.usecase

import zio._

import feed.advertisement.domain.entity
import feed.advertisement.domain.entity.AdvertisementError
import feed.advertisement.repository.AdvertisementRepository

final class AdvertisementService(advertisementRepo: AdvertisementRepository) {
  def getAll: IO[AdvertisementError, List[entity.Advertisement]] = advertisementRepo.getAll
}

object AdvertisementService {
  val layer: ZLayer[AdvertisementRepository, Nothing, AdvertisementService] = ZLayer
    .derive[AdvertisementService]
}
