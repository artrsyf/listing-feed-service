package feed.advertisement.domain.entity

sealed trait AdvertisementError

object AdvertisementError {
  final case class PersistenceLayerError(msg: String) extends AdvertisementError
}
