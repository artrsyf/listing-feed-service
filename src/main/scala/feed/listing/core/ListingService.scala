package feed.listing.core

import java.time.Instant
import java.util.UUID

import zio._

import feed.listing.core.entity.ListingError
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.http.SearchListingsRequest
import feed.shared.collections.EventQueue

final class ListingService(
    listingRepo: ListingRepository,
    listingSearchEngine: ListingSearchIndexEngine,
    listingCreateIndexQueue: EventQueue[entity.Listing]
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

  def getListing(listingId: ListingId): IO[ListingError, entity.Listing] =
    listingRepo.getById(listingId).someOrFail(ListingError.Notfound)

  def createListing(listing: entity.Listing): IO[ListingError, UUID] =
    for {
      _ <- listingRepo.create(listing)
      _ <- listingCreateIndexQueue.push(listing)
    } yield listing.id
}

object ListingService {
  val layer: RLayer[ListingRepository & ListingSearchIndexEngine & EventQueue[
    entity.Listing
  ], ListingService] =
    ZLayer.derive[ListingService]
}
