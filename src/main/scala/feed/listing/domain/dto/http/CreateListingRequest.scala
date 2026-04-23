package feed.listing.domain.dto.http

import sttp.tapir.Schema
import zio.json._

final case class CreateListingRequest(
  title: String,
  description: String,
  price: BigDecimal,
  currency: String,
  imageUrls: List[String])

object CreateListingRequest {
  implicit val codec: JsonCodec[CreateListingRequest] = DeriveJsonCodec
    .gen[CreateListingRequest]
  implicit val schema: Schema[CreateListingRequest] = Schema.derived[CreateListingRequest]
}
