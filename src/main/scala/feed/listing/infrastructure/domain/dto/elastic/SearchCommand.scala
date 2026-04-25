package feed.listing.infrastructure.domain.dto.elastic

final case class SearchCommand(criteria: ListingSearchCriteria, page: PageRequest)

final case class ListingSearchCriteria(
    query: Option[String],
    minPrice: Option[BigDecimal],
    maxPrice: Option[BigDecimal]
)

final case class PageRequest(cursor: Option[String], limit: Int)
