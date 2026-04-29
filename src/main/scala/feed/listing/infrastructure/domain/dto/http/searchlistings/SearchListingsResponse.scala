package feed.listing.infrastructure.domain.dto.http.searchlistings

import sttp.tapir.Schema
import zio.json.DeriveJsonCodec
import zio.json.JsonCodec

import feed.listing.infrastructure.domain.dto.http.ListingResponse

final case class SearchListingsResponse(listings: List[ListingResponse], nextCursor: Option[String])

object SearchListingsResponse {
  implicit val codec: JsonCodec[SearchListingsResponse] = DeriveJsonCodec
    .gen[SearchListingsResponse]
  implicit val schema: Schema[SearchListingsResponse] = Schema
    .derived[SearchListingsResponse]
}
