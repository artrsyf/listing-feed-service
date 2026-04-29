package feed.listing.infrastructure.domain.model.elastic

import java.time.Instant
import java.util.UUID

import zio.json.DeriveJsonCodec
import zio.json.JsonCodec

import feed.listing.core.entity.ListingId

final case class ElasticListingImage(
    id: UUID,
    listingId: ListingId,
    key: String,
    position: Int,
    createdAt: Instant
)

object ElasticListingImage {
  implicit val jsonCodec: JsonCodec[ElasticListingImage] = DeriveJsonCodec.gen[ElasticListingImage]
}
