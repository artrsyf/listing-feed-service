package feed.listing.domain.entity

sealed trait ListingStatus extends Product with Serializable

object ListingStatus {
  case object Active  extends ListingStatus
  case object Sold    extends ListingStatus
  case object Draft   extends ListingStatus
  case object Deleted extends ListingStatus
}
