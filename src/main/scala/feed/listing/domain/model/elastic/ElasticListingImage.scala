package feed.listing.domain.model.elastic

import java.time.Instant
import java.util.UUID

import zio.json.DeriveJsonCodec
import zio.json.JsonCodec

import feed.listing.domain.types.ListingId

final case class ElasticListingImage(
  id: UUID,
  listingId: ListingId,
  url: String,
  key: String,
  position: Int,
  createdAt: Instant)

object ElasticListingImage {
  implicit val jsonCodec: JsonCodec[ElasticListingImage] = DeriveJsonCodec.gen[ElasticListingImage]
}
