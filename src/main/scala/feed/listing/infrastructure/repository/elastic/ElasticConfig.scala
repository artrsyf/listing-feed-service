package feed.listing.infrastructure.repository.elastic

import zio.Config
import zio.config.magnolia.deriveConfig

final case class ElasticConfig(
  url: String,
  listingIndexName: String)

object ElasticConfig {
  implicit val config: Config[ElasticConfig] =
    deriveConfig[ElasticConfig]
      .nested("elastic")
      .nested("feed")
}
