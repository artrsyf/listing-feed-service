package feed.listing.core.entity

sealed trait ListingError extends Product with Serializable

object ListingError {
  final case class PersistenceLayerError(msg: String) extends ListingError

  final case class ValidationError(msg: String) extends ListingError

  case object NotFound extends ListingError
}
