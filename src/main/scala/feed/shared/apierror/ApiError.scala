package feed.shared.apierror

import sttp.model.StatusCode
import sttp.tapir._
import sttp.tapir.json.zio.jsonBody
import zio.json._

sealed trait ApiError extends Product with Serializable

object ApiError {
  final case class Internal(msg: String) extends ApiError

  object Internal {
    implicit val internalCodec: JsonCodec[Internal] = DeriveJsonCodec
      .gen[Internal]
    implicit val internalSchema: Schema[Internal] = Schema
      .derived[Internal]

    val default =
      Internal("Internal server error")
  }

  final case class NotFound(msg: String) extends ApiError

  object NotFound {
    implicit val notFoundCodec: JsonCodec[NotFound] = DeriveJsonCodec
      .gen[NotFound]
    implicit val notFoundSchema: Schema[NotFound] = Schema
      .derived[NotFound]

    val listing = NotFound("Listing not found")

    val default =
      NotFound("Object not found")
  }

  final case class BadRequest(msg: String) extends ApiError

  object BadRequest {
    implicit val notFoundCodec: JsonCodec[BadRequest] = DeriveJsonCodec
      .gen[BadRequest]
    implicit val notFoundSchema: Schema[BadRequest] = Schema
      .derived[BadRequest]

    val listingCreate = BadRequest("Incorrect request for listing create")

    val default =
      BadRequest("Bad request")
  }

  val errorMapper = oneOf[ApiError](
    oneOfVariant(StatusCode.NotFound, jsonBody[ApiError.NotFound]),
    oneOfVariant(StatusCode.InternalServerError, jsonBody[ApiError.Internal]),
    oneOfVariant(StatusCode.BadRequest, jsonBody[ApiError.BadRequest]),
  )
}
