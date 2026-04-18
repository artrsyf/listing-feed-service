package feed.advertisement.delivery

import feed.advertisement.domain.dto
import feed.advertisement.shared.apierror.ApiError

import zio._

trait AdvertisementHandler {
  def getAll: IO[ApiError, dto.GetAllAdvertisementsResponse]
}
