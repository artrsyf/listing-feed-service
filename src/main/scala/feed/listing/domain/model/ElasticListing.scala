package feed.listing.domain.model

import java.time.Instant

import com.sksamuel.elastic4s.Indexable
import zio.json.DeriveJsonCodec
import zio.json.EncoderOps
import zio.json.JsonCodec

import feed.listing.domain.types.ListingId

case class ElasticListing(
  id: ListingId,
  title: String,
  description: String,
  price: BigDecimal,
  currency: String,
  status: ElasticListingStatus,
  images: List[ListingImage],
  createdAt: Instant,
  updatedAt: Instant)

object ElasticListing {
  implicit val jsonCodec: JsonCodec[ElasticListing] = DeriveJsonCodec.gen[ElasticListing]
  implicit val indexable: Indexable[ElasticListing] =
    (el: ElasticListing) => el.toJson
}
