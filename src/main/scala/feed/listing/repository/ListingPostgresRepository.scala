package feed.listing.repository

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
import feed.listing.domain.model
import feed.listing.domain.model.ListingStatusDoobieImplicits._
import feed.listing.domain.types.ListingId

class ListingPostgresRepository(xa: Transactor[Task]) extends ListingRepository {
  override def getRecentListings(limit: Int): IO[PersistenceLayerError, List[entity.Listing]] =
    (for {
      listings <- sql"""
        SELECT id, title, description, price, currency, status, created_at, updated_at
        FROM listings
        WHERE status = 'ACTIVE'::listing_status
        ORDER BY created_at DESC
        LIMIT $limit
      """
        .query[model.Listing]
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
          status = l.status,
          images = images.map(_.transformInto[entity.ListingImage]),
          createdAt = l.createdAt,
          updatedAt = l.updatedAt
        )
      }
    } yield result).mapError(e => PersistenceLayerError(e.getMessage))

  override def getById(listingId: ListingId): IO[PersistenceLayerError, Option[entity.Listing]] =
    (for {
      listing <- sql"""
        SELECT id, title, description, price, currency, status, created_at, updated_at
        FROM listings
        WHERE id = $listingId
        LIMIT 1
      """
        .query[model.Listing]
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
          status = l.status,
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
    .query[model.ListingImage]
    .to[List]
    .transact(xa)

  // TODO: Починить костыль с енамом
  private def insertListing(listing: entity.Listing): ConnectionIO[Int] =
    sql"""
      INSERT INTO listings (
        id, title, description, price, currency, status, created_at, updated_at
      )
      VALUES (
        ${listing.id}, ${listing.title}, ${listing.description}, ${listing.price},
        ${listing.currency}, ${listing.status}::listing_status, ${listing.createdAt}, ${listing.updatedAt}
      )
    """.update.run

  private def insertImages(
    listingId: ListingId,
    createdAt: Instant,
    images: List[entity.ListingImage]
  ): ConnectionIO[Int] = {
    val modelImages = images.map { img =>
      model.ListingImage(
        id = img.id,
        listingId = listingId,
        url = img.url,
        key = img.url,
        position = img.position,
        createdAt = createdAt
      )
    }

    Update[model.ListingImage]("""
      INSERT INTO listing_images (
        id, listing_id, url, key, position, created_at
      ) VALUES (?, ?, ?, ?, ?, ?)
    """).updateMany(modelImages)
  }
}

object ListingPostgresRepository {
  val layer: ZLayer[HikariTransactor[Task], Nothing, ListingRepository] = ZLayer
    .fromFunction(new ListingPostgresRepository(_))
}
