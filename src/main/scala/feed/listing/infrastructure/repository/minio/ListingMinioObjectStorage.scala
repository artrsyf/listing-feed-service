package feed.listing.infrastructure.repository.minio

import io.minio._
import io.minio.http.Method
import zio._

import feed.listing.core.entity.ListingError
import feed.listing.infrastructure.query.ObjectStorage

final class ListingMinioObjectStorage(client: MinioClient, minioConfig: MinioConfig)
    extends ObjectStorage {
  override def getPublicUrl(key: String): IO[ListingError, String] =
    ZIO
      .attempt {
        client
          .getPresignedObjectUrl(
            GetPresignedObjectUrlArgs
              .builder()
              .method(Method.GET)
              .bucket(minioConfig.bucket)
              .`object`(key)
              .expiry(60 * 60)
              .build()
          )
          .replace(minioConfig.privateEndpoint, minioConfig.publicEndpoint)
      }
      .mapError(e => ListingError.PersistenceLayerError(e.getMessage))

  override def generateUploadUrl(key: String): IO[ListingError, String] =
    ZIO
      .attempt {
        client
          .getPresignedObjectUrl(
            GetPresignedObjectUrlArgs
              .builder()
              .method(Method.PUT)
              .bucket(minioConfig.bucket)
              .`object`(key)
              .expiry(60 * 10)
              .build()
          )
          .replace(minioConfig.privateEndpoint, minioConfig.publicEndpoint)
      }
      .mapError(e => ListingError.PersistenceLayerError(e.getMessage))

  override def deleteObject(key: String): IO[ListingError, Unit] =
    ZIO
      .attempt {
        client.removeObject(
          RemoveObjectArgs
            .builder()
            .bucket(minioConfig.bucket)
            .`object`(key)
            .build()
        )
      }
      .mapError(e => ListingError.PersistenceLayerError(e.getMessage))
}

object ListingMinioObjectStorage {
  val layer: RLayer[MinioClient, ObjectStorage] =
    ZLayer.fromZIO {
      for {
        client <- ZIO.service[MinioClient]
        config <- ZIO.config[MinioConfig]

        _ <- ensureBucketExists(client, config.bucket).retry(
          Schedule.exponential(200.millis) && Schedule.recurs(5)
        )
      } yield new ListingMinioObjectStorage(client, config)
    }

  private def ensureBucketExists(client: MinioClient, bucket: String): Task[Unit] =
    for {
      exists <- ZIO.attempt {
        client.bucketExists(
          BucketExistsArgs
            .builder()
            .bucket(bucket)
            .build()
        )
      }

      _ <- ZIO.when(!exists) {
        ZIO.attempt {
          client.makeBucket(
            MakeBucketArgs
              .builder()
              .bucket(bucket)
              .build()
          )
        }
      }
    } yield ()
}
