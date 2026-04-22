package feed.listing.domain.model

import java.time.Instant
import java.util.UUID

import zio.json.DeriveJsonCodec
import zio.json.JsonCodec

import feed.listing.domain.types.ListingId

final case class ListingImage(
  id: UUID,
  listingId: ListingId,
  url: String,
  key: String,
  position: Int,
  createdAt: Instant)

object ListingImage {
  implicit val jsonCodec: JsonCodec[ListingImage] = DeriveJsonCodec.gen[ListingImage]
}
