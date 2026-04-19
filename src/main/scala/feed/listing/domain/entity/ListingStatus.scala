package feed.listing.domain.entity

import doobie.util.meta.Meta

sealed trait ListingStatus

object ListingStatus {
  case object Active  extends ListingStatus
  case object Sold    extends ListingStatus
  case object Draft   extends ListingStatus
  case object Deleted extends ListingStatus
}
