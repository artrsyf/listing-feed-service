package feed.listing.infrastructure.domain.dto.http

import java.time.Instant
import java.util.UUID

import sttp.tapir.Schema
import zio.json._

final case class GetAllListingsResponse(listings: List[ListingResponse], cursor: Option[String])

object GetAllListingsResponse {
  implicit val codec: JsonCodec[GetAllListingsResponse] = DeriveJsonCodec
    .gen[GetAllListingsResponse]
  implicit val schema: Schema[GetAllListingsResponse] = Schema
    .derived[GetAllListingsResponse]
}

final case class ListingResponse(
    id: UUID,
    title: String,
    description: String,
    price: BigDecimal,
    currency: String,
    images: List[String],
    createdAt: Instant
)

object ListingResponse {
  implicit val codec: JsonCodec[ListingResponse] = DeriveJsonCodec
    .gen[ListingResponse]
  implicit val schema: Schema[ListingResponse] = Schema.derived[ListingResponse]
}
