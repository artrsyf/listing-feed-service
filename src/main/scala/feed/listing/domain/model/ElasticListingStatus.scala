package feed.listing.domain.model

import enumeratum.Enum
import enumeratum.EnumEntry
import enumeratum.EnumEntry.Lowercase
import zio.json.JsonCodec
import zio.json.JsonDecoder
import zio.json.JsonEncoder

// TODO: Вынести в shared часть
trait ZioJsonEnum[A <: EnumEntry] { self: Enum[A] =>
  implicit val jsonCodec: JsonCodec[A] = JsonCodec.fromEncoderDecoder(
    JsonEncoder.string.contramap(_.entryName),
    JsonDecoder.string.mapOrFail(self.withNameEither(_).left.map(_.getMessage()))
  )
}

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
