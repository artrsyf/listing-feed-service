package feed.listing.infrastructure.repository.postgres

import java.time.Instant

import doobie._
import doobie.hikari.HikariTransactor
import doobie.implicits._
import doobie.postgres.implicits._
import io.scalaland.chimney.dsl._
import zio._
import zio.interop.catz._

import feed.listing.domain.entity
import feed.listing.domain.entity.ListingError.PersistenceLayerError
import feed.listing.domain.entity.ListingId
import feed.listing.infrastructure.domain.model.postgres
import feed.listing.infrastructure.domain.model.postgres.Listing
import feed.listing.infrastructure.domain.model.postgres.ListingImage
import feed.listing.repository.ListingRepository

class ListingPostgresRepository(xa: Transactor[Task]) extends ListingRepository {
  override def getRecentListings(
    cursor: Option[Instant],
    limit: Int
  ): IO[PersistenceLayerError, List[entity.Listing]] = {
    val req = cursor match {
      case Some(c) =>
        sql"""
        SELECT id, title, description, price, currency, status, created_at, updated_at
        FROM listings
        WHERE status = 'ACTIVE'::listing_status
          AND created_at < $c
        ORDER BY created_at DESC
        LIMIT $limit
      """
      case None =>
        sql"""
        SELECT id, title, description, price, currency, status, created_at, updated_at
        FROM listings
        WHERE status = 'ACTIVE'::listing_status
        ORDER BY created_at DESC
        LIMIT $limit
      """
    }

    (for {
      listings <- req
        .query[Listing]
        .to[List]
        .transact(xa)

      result <- ZIO.foreach(listings) { l =>
        for {
          images <- getListingImages(l.id)
        } yield entity.Listing(
          id = l.id,
          title = l.title,
          description = l.description,
          price = l.price,
          currency = l.currency,
          status = l.status.transformInto[entity.ListingStatus],
          images = images.map(_.transformInto[entity.ListingImage]),
          createdAt = l.createdAt,
          updatedAt = l.updatedAt
        )
      }
    } yield result).mapError(e => PersistenceLayerError(e.getMessage))
  }

  override def getById(listingId: ListingId): IO[PersistenceLayerError, Option[entity.Listing]] =
    (for {
      listing <- sql"""
        SELECT id, title, description, price, currency, status, created_at, updated_at
        FROM listings
        WHERE id = $listingId
        LIMIT 1
      """
        .query[Listing]
        .option
        .transact(xa)
      images <- getListingImages(listingId)
      result = listing.map(l =>
        entity.Listing(
          id = l.id,
          title = l.title,
          description = l.description,
          price = l.price,
          currency = l.currency,
          status = l.status.transformInto[entity.ListingStatus],
          images = images.map(_.transformInto[entity.ListingImage]),
          createdAt = l.createdAt,
          updatedAt = l.updatedAt
        )
      )
    } yield result).mapError(e => PersistenceLayerError(e.getMessage))

  override def create(listing: entity.Listing): IO[PersistenceLayerError, Unit] = {
    val action =
      for {
        _ <- insertListing(listing)
        _ <- insertImages(listing.id, listing.createdAt, listing.images)
      } yield ()

    action
      .transact(xa)
      .mapError(e => PersistenceLayerError(e.getMessage))
  }

  private def getListingImages(listingId: ListingId) = sql"""
        SELECT id, listing_id, url, key, position, created_at
        FROM listing_images
        WHERE listing_id = $listingId
        ORDER BY position ASC
      """
    .query[ListingImage]
    .to[List]
    .transact(xa)

  private def insertListing(listing: entity.Listing): ConnectionIO[Int] = {
    val listingModel = listing.transformInto[Listing]

    sql"""
      INSERT INTO listings (
        id, title, description, price, currency, status, created_at, updated_at
      )
      VALUES (
        ${listingModel.id}, ${listingModel.title}, ${listingModel.description}, ${listingModel.price},
        ${listingModel.currency}, ${listingModel.status}::listing_status, ${listingModel.createdAt}, ${listingModel.updatedAt}
      )
    """.update.run
  }

  private def insertImages(
    listingId: ListingId,
    createdAt: Instant,
    images: List[entity.ListingImage]
  ): ConnectionIO[Int] = {
    val modelImages = images.map { img =>
      postgres.ListingImage(
        id = img.id,
        listingId = listingId,
        url = img.url,
        key = img.url,
        position = img.position,
        createdAt = createdAt
      )
    }

    Update[ListingImage]("""
      INSERT INTO listing_images (
        id, listing_id, url, key, position, created_at
      ) VALUES (?, ?, ?, ?, ?, ?)
    """).updateMany(modelImages)
  }
}

object ListingPostgresRepository {
  val layer: RLayer[HikariTransactor[Task], ListingRepository] =
    ZLayer.derive[ListingPostgresRepository]
}
