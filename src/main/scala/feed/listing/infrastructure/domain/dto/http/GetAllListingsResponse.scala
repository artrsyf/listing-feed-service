package feed.listing.infrastructure.domain.dto.http

import java.time.Instant

import sttp.tapir.Schema
import zio.json._

import feed.listing.core.entity.ListingId

final case class GetAllListingsResponse(listings: List[ListingResponse], cursor: Option[String])

object GetAllListingsResponse {
  implicit val codec: JsonCodec[GetAllListingsResponse] = DeriveJsonCodec
    .gen[GetAllListingsResponse]
  implicit val schema: Schema[GetAllListingsResponse] = Schema
    .derived[GetAllListingsResponse]
}

final case class ListingResponse(
    id: ListingId,
    title: String,
    description: String,
    price: BigDecimal,
    currency: String,
    images: List[ListingImageResponse],
    createdAt: Instant
)

object ListingResponse {
  implicit val codec: JsonCodec[ListingResponse] = DeriveJsonCodec
    .gen[ListingResponse]
  implicit val schema: Schema[ListingResponse] = Schema.derived[ListingResponse]
}

final case class ListingImageResponse(url: String, position: Int)

object ListingImageResponse {
  implicit val codec: JsonCodec[ListingImageResponse] = DeriveJsonCodec.gen
  implicit val schema: Schema[ListingImageResponse] = Schema.derived
}
