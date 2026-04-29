package feed.listing.infrastructure.repository.minio

import io.minio.{MinioClient => JavaMinioClient}
import zio._

object MinioClient {
  val layer: ZLayer[Any, Throwable, JavaMinioClient] =
    ZLayer.fromZIO {
      for {
        config <- ZIO.config[MinioConfig]
        client <- ZIO.attempt {
          JavaMinioClient
            .builder()
            .endpoint(config.privateEndpoint)
            .credentials(config.accessKey, config.secretKey)
            .build()
        }
      } yield client
    }
}
