package feed.listing.usecase

import java.time.Instant
import java.util.UUID

import zio._

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError
import feed.listing.domain.entity.ListingId
import feed.listing.repository.ListingRepository
import feed.listing.repository.ListingSearchEngine
import feed.listing.shared.collections.EventQueue

final class ListingService(
  listingRepo: ListingRepository,
  listingSearchEngine: ListingSearchEngine,
  listingCreateIndexQueue: EventQueue[entity.Listing])
    extends ZLayer.Derive.Scoped[Any, Nothing] {
  override def scoped(implicit trace: Trace): ZIO[Any & Scope, Nothing, Any] =
    listingCreateIndexQueue.subscribe { listingsChunk =>
      listingSearchEngine.insertMany(listingsChunk).tapError(e => ZIO.logError(e.msg)).ignore
    }

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
  val layer
    : RLayer[ListingRepository & ListingSearchEngine & EventQueue[entity.Listing], ListingService] =
    ZLayer.derive[ListingService]
}
