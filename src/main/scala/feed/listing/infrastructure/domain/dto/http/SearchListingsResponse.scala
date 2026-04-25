package feed.listing.infrastructure.domain.dto.http

import sttp.tapir.Schema
import zio.json.DeriveJsonCodec
import zio.json.JsonCodec

final case class SearchListingsResponse(listings: List[ListingResponse], nextCursor: Option[String])

object SearchListingsResponse {
  implicit val codec: JsonCodec[SearchListingsResponse] = DeriveJsonCodec
    .gen[SearchListingsResponse]
  implicit val schema: Schema[SearchListingsResponse] = Schema
    .derived[SearchListingsResponse]
}
