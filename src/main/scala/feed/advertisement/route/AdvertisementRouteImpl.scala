package feed.advertisement.route

import sttp.tapir.json.zio._
import sttp.tapir.ztapir._
import zio._

import feed.advertisement.delivery.AdvertisementHandler
import feed.advertisement.domain.dto
import feed.advertisement.shared.apierror.ApiError.errorMapper

final class AdvertisementRouteImpl(advertisementHandler: AdvertisementHandler)
    extends AdvertisementRoute {

  val baseEndpoint = endpoint.tag("some tag").in("api" / "feed/advertisement")

  val getAllAdvertisements = baseEndpoint.get
    .out(jsonBody[dto.GetAllAdvertisementsResponse])
    .errorOut(errorMapper)
    .zServerLogic(_ => advertisementHandler.getAll)

  override val routes = Chunk(getAllAdvertisements)
}

object AdvertisementRouteImpl {
  val layer: ZLayer[AdvertisementHandler, Nothing, AdvertisementRoute] = ZLayer
    .derive[AdvertisementRouteImpl]
}
