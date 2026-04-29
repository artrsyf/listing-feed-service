package feed.listing.infrastructure.domain.dto.http.generateuploadurl

import sttp.tapir.Schema
import zio.json._

final case class GenerateUploadUrlRequest(contentType: String)

object GenerateUploadUrlRequest {
  implicit val codec: JsonCodec[GenerateUploadUrlRequest] = DeriveJsonCodec.gen
  implicit val schema: Schema[GenerateUploadUrlRequest] = Schema.derived[GenerateUploadUrlRequest]
}
