package feed.listing.infrastructure.domain.model.postgres

import enumeratum.EnumEntry.Uppercase
import enumeratum._

sealed trait ListingStatus extends Product with Serializable with EnumEntry with Uppercase

object ListingStatus extends Enum[ListingStatus] with DoobieEnum[ListingStatus] {
  case object Active extends ListingStatus
  case object Sold extends ListingStatus
  case object Draft extends ListingStatus
  case object Deleted extends ListingStatus

  val values: IndexedSeq[ListingStatus] = findValues
}
