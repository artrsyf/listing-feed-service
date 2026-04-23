package feed.listing.domain.model.postgres

import doobie.util.meta.Meta

sealed trait ListingStatus extends Product with Serializable

object ListingStatus {
  case object Active  extends ListingStatus
  case object Sold    extends ListingStatus
  case object Draft   extends ListingStatus
  case object Deleted extends ListingStatus

  def fromString(s: String): ListingStatus = s match {
    case "ACTIVE"  => Active
    case "SOLD"    => Sold
    case "DRAFT"   => Draft
    case "DELETED" => Deleted
  }

  def toString(s: ListingStatus): String = s match {
    case Active  => "ACTIVE"
    case Sold    => "SOLD"
    case Draft   => "DRAFT"
    case Deleted => "DELETED"
  }

  implicit val meta: Meta[ListingStatus] =
    Meta[String].timap(fromString)(toString)
}
