package feed.listing.infrastructure.domain.dto.http.generateuploadurl

import sttp.tapir.Schema
import zio.json._

final case class GenerateUploadUrlResponse(key: String, uploadUrl: String)

object GenerateUploadUrlResponse {
  implicit val codec: JsonCodec[GenerateUploadUrlResponse] = DeriveJsonCodec.gen
  implicit val schema: Schema[GenerateUploadUrlResponse] = Schema.derived[GenerateUploadUrlResponse]

}
