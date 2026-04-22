package feed.listing.domain.entity

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

sealed trait ListingStatus extends EnumEntry with Lowercase with Product with Serializable

object ListingStatus extends Enum[ListingStatus] with ZioJsonEnum[ListingStatus] {
  override val values: IndexedSeq[ListingStatus] = findValues

  case object Active  extends ListingStatus
  case object Sold    extends ListingStatus
  case object Draft   extends ListingStatus
  case object Deleted extends ListingStatus
}
