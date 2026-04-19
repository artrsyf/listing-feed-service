package feed.listing.domain.entity

sealed trait ListingError

object ListingError {
  final case class PersistenceLayerError(msg: String) extends ListingError
}
