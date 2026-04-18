package feed.advertisement.delivery

import io.scalaland.chimney.dsl._
import zio._

import feed.advertisement.domain.dto
import feed.advertisement.domain.entity
import feed.advertisement.shared.apierror.ApiError
import feed.advertisement.usecase.AdvertisementService

final class AdvertisementHttpHandler(advertisementService: AdvertisementService)
    extends AdvertisementHandler {
  override def getAll: IO[ApiError, dto.GetAllAdvertisementsResponse] = advertisementService.getAll
    .mapBoth(
      { case entity.AdvertisementError.PersistenceLayerError(_) => ApiError.Internal.default },
      advertisements =>
        dto
          .GetAllAdvertisementsResponse(advertisements.map(_.transformInto[dto.Advertisement]))
    )
}

object AdvertisementHttpHandler {
  val layer: ZLayer[AdvertisementService, Nothing, AdvertisementHandler] = ZLayer
    .derive[AdvertisementHttpHandler]
}
