package feed.listing.shared.instances

import enumeratum.Enum
import enumeratum.EnumEntry
import zio.json.JsonCodec
import zio.json.JsonDecoder
import zio.json.JsonEncoder

trait ZioJsonEnum[A <: EnumEntry] { self: Enum[A] =>
  implicit val jsonCodec: JsonCodec[A] = JsonCodec.fromEncoderDecoder(
    JsonEncoder.string.contramap(_.entryName),
    JsonDecoder.string.mapOrFail(self.withNameEither(_).left.map(_.getMessage()))
  )
}
