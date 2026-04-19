package feed.listing.shared.infrastructure

import zio.Config
import zio.config.magnolia._

final case class DbConfig(
  url: String,
  user: String,
  password: String)

object DbConfig {
  implicit val config: Config[DbConfig] =
    deriveConfig[DbConfig]
      .nested("db")
      .nested("feed")
}
