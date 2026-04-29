package feed.listing.infrastructure.query

import zio._

import feed.listing.core.entity.ListingError

trait ObjectStorage {
  def getPublicUrl(key: String): IO[ListingError, String]

  def generateUploadUrl(key: String): IO[ListingError, String]

  def deleteObject(key: String): IO[ListingError, Unit]
}
