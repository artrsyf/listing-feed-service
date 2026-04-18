package feed.advertisement.domain.dto

import sttp.tapir.Schema
import zio.json._

final case class GetAllAdvertisementsResponse(advertisements: List[Advertisement])

object GetAllAdvertisementsResponse {
  implicit val codec: JsonCodec[GetAllAdvertisementsResponse] = DeriveJsonCodec
    .gen[GetAllAdvertisementsResponse]
  implicit val schema: Schema[GetAllAdvertisementsResponse] = Schema
    .derived[GetAllAdvertisementsResponse]
}

final case class Advertisement(id: Int)

object Advertisement {
  implicit val codec: JsonCodec[Advertisement] = DeriveJsonCodec.gen[Advertisement]
  implicit val schema: Schema[Advertisement]   = Schema.derived[Advertisement]
}
