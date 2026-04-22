package feed.listing.usecase

import java.time.Instant
import java.util.UUID

import io.scalaland.chimney.dsl._
import zio._
import zio.stream.ZStream

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError
import feed.listing.domain.model.ElasticListing
import feed.listing.domain.model.ListingImage
import feed.listing.domain.types.ListingId
import feed.listing.repository.ElasticPayload
import feed.listing.repository.ListingElasticRepository
import feed.listing.repository.ListingRepository
import feed.listing.repository.ListingSearchRepository

final class ListingService(
  listingRepo: ListingRepository,
  listingSearchRepostiory: ListingSearchRepository,
  listingCreateIndexQueue: AbstractDaemonQueue[entity.Listing])
    extends ZLayer.Derive.Scoped[Any, Nothing] {
  override def scoped(implicit trace: Trace): ZIO[Any & Scope, Nothing, Any] =
    listingCreateIndexQueue.subscribe { listingsChunk =>
      listingSearchRepostiory
        .indexBulk(listingsChunk.map { listing =>
          val elasticListing =
            listing
              .into[ElasticListing]
              .withFieldRenamed(_.status, _.status)
              .withFieldComputed(
                _.images,
                _.images.map(img =>
                  ListingImage(img.id, listing.id, img.url, img.url, img.position, Instant.now())
                )
              )
              .transform

          ElasticPayload[ElasticListing]("temp_id", 1, elasticListing)
        })
        .ignore
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
  implicit val defaultListingCreateIndexQueueLayer
    : ZLayer.Derive.Default.WithContext[Any, Config.Error, AbstractDaemonQueue[entity.Listing]] =
    ZLayer.Derive.Default.fromLayer {
      ZLayer.scoped {
        AbstractDaemonQueue.makeChunked(100, 1.second, 10.seconds)
      }
    }

  val layer: RLayer[ListingRepository & ListingSearchRepository, ListingService] =
    ZLayer.derive[ListingService]
}

// TODO: Порефачить работу с очередью, вытащить параметр в конфиг
sealed trait DaemonQueue[A] {
  def subscribe(f: Chunk[A] => UIO[Unit]): UIO[Unit]

  def push(record: A): UIO[Unit]
}

abstract class AbstractDaemonQueue[A](val subscribersRef: Ref[Chunk[Chunk[A] => UIO[Unit]]])
    extends DaemonQueue[A] {
  override def subscribe(f: Chunk[A] => UIO[Unit]): UIO[Unit] =
    subscribersRef.update(_ :+ f)
}

object AbstractDaemonQueue {
  def makeChunked[A](
    chunkSize: Int,
    withinDuration: Duration,
    shutdownTimeout: Duration
  ) =
    for {
      queue               <- Queue.bounded[Option[A]](chunkSize)
      emptySubscribersRef <- Ref.make(Chunk.empty[Chunk[A] => UIO[Unit]])
      searchEngineQueue = new AbstractDaemonQueue[A](emptySubscribersRef) {
        override def push(record: A): UIO[Unit] =
          queue.offer(Some(record)).unit
      }
      fiber <- ZStream
        .fromQueueWithShutdown(queue)
        .collectWhileSome
        .groupedWithin(chunkSize, withinDuration)
        .runForeach((events: Chunk[A]) =>
          emptySubscribersRef.get.flatMap { subscribers =>
            ZIO.foreachDiscard(subscribers)(f => f(events))
          }
        )
        .forkScoped

      _ <- ZIO.addFinalizer(
        queue.offer(None).exit *> fiber.await.interruptible.timeout(shutdownTimeout)
      )
    } yield searchEngineQueue
}
