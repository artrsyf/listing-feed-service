package feed.listing.core

import zio.Config
import zio.config.magnolia._

final case class ListingConfig(limit: Int, imagesLimit: Int)

object ListingConfig {
  implicit val config: Config[ListingConfig] =
    deriveConfig[ListingConfig]
      .nested("listing")
      .nested("feed")
}
