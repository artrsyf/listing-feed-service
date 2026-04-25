package feed.listing.infrastructure.delivery

import java.time.Instant

import io.scalaland.chimney.dsl._
import zio._

import feed.listing.core.ListingService
import feed.listing.core.entity
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.http.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.shared.apierror.ApiError

final class ListingHttpHandler(
  listingService: ListingService,
  listingConfig: ListingConfig)
    extends ListingHandler {
  override def getRecentListings(
    cursor: Option[Instant],
    limit: Option[Int]
  ): IO[ApiError, GetAllListingsResponse] =
    listingService
      .getRecentListings(cursor, limit.getOrElse(listingConfig.limit))
      .mapBoth(
        {
          case entity.ListingError
                .PersistenceLayerError(_) =>
            ApiError.Internal.default
          case _ => ApiError.Internal.default
        },
        listings =>
          feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse(
            listings
              .map(
                _.into[ListingResponse]
                  .withFieldComputed(_.images, _.images.map(_.url))
                  .transform
              )
          )
      )

  override def getListing(listingId: ListingId): IO[ApiError, ListingResponse] =
    listingService
      .getListing(listingId)
      .mapBoth(
        {
          case entity.ListingError
                .PersistenceLayerError(_) =>
            ApiError.Internal.default

          case entity.ListingError.Notfound =>
            ApiError.NotFound.listing
        },
        _.into[ListingResponse]
          .withFieldComputed(_.images, _.images.map(_.url))
          .transform
      )

  override def createListing(req: CreateListingRequest): IO[ApiError, CreateListingResponse] =
    for {
      id  <- ZIO.succeed(java.util.UUID.randomUUID())
      now <- Clock.instant
      listing = entity.Listing(
        id = id,
        title = req.title,
        description = req.description,
        price = req.price,
        currency = req.currency,
        status = entity.ListingStatus.Active,
        images = req.imageUrls.zipWithIndex.map { case (url, idx) =>
          entity.ListingImage(id = java.util.UUID.randomUUID(), url = url, position = idx)
        },
        createdAt = now,
        updatedAt = now
      )
      _ <- listingService
        .createListing(listing)
        .mapError {
          case entity.ListingError.PersistenceLayerError(_) =>
            ApiError.Internal.default
          case _ => ApiError.Internal.default
        }
    } yield feed.listing.infrastructure.domain.dto.http
      .CreateListingResponse(listing.id, listing.createdAt)
}

object ListingHttpHandler {
  val layer: RLayer[ListingService, ListingHandler] = ZLayer
    .derive[ListingHttpHandler]
}
