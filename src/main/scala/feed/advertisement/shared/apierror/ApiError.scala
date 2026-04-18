package feed.advertisement.shared.apierror

import sttp.model.StatusCode
import sttp.tapir._
import sttp.tapir.json.zio.jsonBody
import zio.json._

sealed trait ApiError
object ApiError {
  final case class Internal(msg: String) extends ApiError

  object Internal {
    implicit val internalCodec: JsonCodec[Internal] = DeriveJsonCodec.gen[Internal]
    implicit val internalSchema: Schema[Internal]   = Schema.derived[Internal]

    val default = Internal("Internal server error")
  }

  final case class NotFound(msg: String) extends ApiError

  object NotFound {
    implicit val notFoundCodec: JsonCodec[NotFound] = DeriveJsonCodec.gen[NotFound]
    implicit val notFoundSchema: Schema[NotFound]   = Schema.derived[NotFound]
  }

  val errorMapper = oneOf[ApiError](
    oneOfVariant(StatusCode.NotFound, jsonBody[ApiError.NotFound]),
    oneOfVariant(StatusCode.InternalServerError, jsonBody[ApiError.Internal])
  )
}
