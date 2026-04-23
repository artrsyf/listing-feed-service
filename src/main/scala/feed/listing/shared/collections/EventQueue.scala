package feed.listing.shared.collections

import zio._
import zio.stream.ZStream

import feed.listing.domain.entity

trait EventQueue[A] {
  def subscribe(f: Chunk[A] => UIO[Unit]): UIO[Unit]

  def push(record: A): UIO[Unit]
}

object EventQueue {
  implicit val chunkedListingEventQueueLayer
    : ZLayer[Any, Config.Error, EventQueue[entity.Listing]] =
    ZLayer.scoped {
      EventQueue.makeChunked[entity.Listing](100, 1.second, 10.seconds)
    }

  abstract class EventQueueWithSubscribersRef[A](
    val subscribersRef: Ref[Chunk[Chunk[A] => UIO[Unit]]])
      extends EventQueue[A] {
    override def subscribe(f: Chunk[A] => UIO[Unit]): UIO[Unit] =
      subscribersRef.update(_ :+ f)

    def push(record: A): UIO[Unit]
  }

  def makeChunked[A](
    chunkSize: Int,
    withinDuration: Duration,
    shutdownTimeout: Duration
  ) =
    for {
      queue               <- Queue.bounded[Option[A]](chunkSize)
      emptySubscribersRef <- Ref.make(Chunk.empty[Chunk[A] => UIO[Unit]])
      searchEngineQueue = new EventQueueWithSubscribersRef[A](emptySubscribersRef) {
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
