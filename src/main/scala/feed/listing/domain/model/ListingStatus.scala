package feed.listing.domain.model

import doobie.util.meta.Meta

import feed.listing.domain.entity.ListingStatus
import feed.listing.domain.entity.ListingStatus._

object ListingStatusDoobieImplicits {
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
