package feed.listing.domain.model.elastic

import enumeratum.Enum
import enumeratum.EnumEntry
import enumeratum.EnumEntry.Lowercase

import feed.listing.shared.instances.ZioJsonEnum

sealed trait ElasticListingStatus extends EnumEntry with Lowercase with Product with Serializable

object ElasticListingStatus
    extends Enum[ElasticListingStatus]
    with ZioJsonEnum[ElasticListingStatus] {
  override val values: IndexedSeq[ElasticListingStatus] = findValues

  case object Active  extends ElasticListingStatus
  case object Sold    extends ElasticListingStatus
  case object Draft   extends ElasticListingStatus
  case object Deleted extends ElasticListingStatus
}
