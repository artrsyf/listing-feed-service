package feed.listing.infrastructure.domain.dto.http

import sttp.tapir.Schema
import zio.json.DeriveJsonCodec
import zio.json.JsonCodec

final case class SearchListingsRequest(
    query: Option[String],
    cursor: Option[String],
    limit: Int,
    minPrice: Option[BigDecimal],
    maxPrice: Option[BigDecimal]
)

object SearchListingsRequest {
  implicit val codec: JsonCodec[SearchListingsRequest] = DeriveJsonCodec
    .gen[SearchListingsRequest]
  implicit val schema: Schema[SearchListingsRequest] = Schema
    .derived[SearchListingsRequest]
}
