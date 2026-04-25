package feed.listing.infrastructure.repository.postgres

import zio.Config
import zio.config.magnolia._

final case class PostgresConfig(url: String, user: String, password: String)

object PostgresConfig {
  implicit val config: Config[PostgresConfig] =
    deriveConfig[PostgresConfig]
      .nested("postgres")
      .nested("feed")
}
