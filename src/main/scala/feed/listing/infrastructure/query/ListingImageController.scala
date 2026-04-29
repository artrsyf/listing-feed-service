package feed.listing.infrastructure.query

import zio.Random
import zio._

import feed.listing.infrastructure.domain.dto.http.generateuploadurl.GenerateUploadUrlResponse
import feed.shared.apierror.ApiError

final class ListingImageController(objectStorage: ObjectStorage) {
  def generateUploadUrl(userId: String): IO[ApiError, GenerateUploadUrlResponse] =
    for {
      randomSuffix <- Random.nextUUID
      key <- ZIO.succeed(s"$userId/${randomSuffix}")
      url <- objectStorage
        .generateUploadUrl(key)
        .orElseFail(ApiError.Internal.default)
    } yield GenerateUploadUrlResponse(key, url)
}

object ListingImageController {
  val layer: RLayer[ObjectStorage, ListingImageController] =
    ZLayer.derive[ListingImageController]
}
