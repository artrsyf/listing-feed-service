package feed.listing.core

import java.time.Instant
import java.util.UUID

import zio._

import feed.listing.core.entity.ListingError
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsRequest
import feed.shared.collections.EventQueue

final class ListingService(
    listingRepo: ListingRepository,
    listingSearchEngine: ListingSearchCreateEngine,
    listingCreateIndexQueue: EventQueue[entity.Listing],
    listingConfig: ListingConfig
) extends ZLayer.Derive.Scoped[Any, Nothing] {
  override def scoped(implicit trace: Trace): ZIO[Any & Scope, Nothing, Any] =
    listingCreateIndexQueue.subscribe { listingsChunk =>
      listingSearchEngine.insertMany(listingsChunk).tapError(e => ZIO.logError(e.msg)).ignore
    }

  @deprecated("Чтение идет через отдельный контроллер", "25-04-2026")
  def getRecentListings(
      cursor: Option[Instant],
      limit: Int
  ): IO[ListingError, List[entity.Listing]] =
    listingRepo.getRecentListings(cursor, limit)

  @deprecated("Чтение идет через отдельный контроллер", "28-04-2026")
  def getListing(listingId: ListingId): IO[ListingError, entity.Listing] =
    listingRepo.getById(listingId).someOrFail(ListingError.NotFound)

  def createListing(listing: entity.Listing): IO[ListingError, UUID] =
    for {
      _ <- ZIO
        .fail(ListingError.ValidationError("Too many images"))
        .when(listing.images.size > listingConfig.imagesLimit)
      _ <- listingRepo.create(listing)
      _ <- listingCreateIndexQueue.push(listing)
    } yield listing.id
}

object ListingService {
  val layer: RLayer[ListingRepository & ListingSearchCreateEngine & EventQueue[
    entity.Listing
  ], ListingService] =
    ZLayer.derive[ListingService]
}
