package feed.listing.infrastructure.repository.minio

import zio.Config
import zio.config.magnolia.deriveConfig

final case class MinioConfig(
    privateEndpoint: String,
    publicEndpoint: String,
    accessKey: String,
    secretKey: String,
    bucket: String
)

object MinioConfig {
  implicit val config: Config[MinioConfig] =
    deriveConfig[MinioConfig]
      .nested("minio")
      .nested("feed")
}
