package feed.listing.usecase

import zio.Config
import zio.config.magnolia._

final case class ListingConfig(limit: Int)

object ListingConfig {
  implicit val config: Config[ListingConfig] =
    deriveConfig[ListingConfig]
      .nested("listing")
      .nested("feed")
}
