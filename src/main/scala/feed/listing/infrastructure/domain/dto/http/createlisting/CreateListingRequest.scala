package feed.listing.infrastructure.domain.dto.http.createlisting

import sttp.tapir.Schema
import zio.json._

final case class CreateListingRequest(
    title: String,
    description: String,
    price: BigDecimal,
    currency: String,
    imageKeys: List[String]
)

object CreateListingRequest {
  implicit val codec: JsonCodec[CreateListingRequest] = DeriveJsonCodec
    .gen[CreateListingRequest]
  implicit val schema: Schema[CreateListingRequest] = Schema.derived[CreateListingRequest]
}
