package feed.listing.infrastructure.domain.dto.http

import java.time.Instant
import java.util.UUID

import sttp.tapir.Schema
import zio.json._

final case class CreateListingResponse(
  id: UUID,
  createdAt: Instant)

object CreateListingResponse {
  implicit val codec: JsonCodec[CreateListingResponse] = DeriveJsonCodec
    .gen[CreateListingResponse]
  implicit val schema: Schema[CreateListingResponse] = Schema.derived[CreateListingResponse]
}
