package feed.listing.delivery

import io.scalaland.chimney.dsl._
import zio._

import feed.listing.domain.dto
import feed.listing.domain.entity
import feed.listing.domain.types.ListingId
import feed.listing.shared.apierror.ApiError
import feed.listing.usecase.ListingService

final class ListingHttpHandler(listingService: ListingService) extends ListingHandler {
  override def getRecentListings: IO[ApiError, dto.GetAllListingsResponse] =
    listingService.getRecentListings
      .mapBoth(
        {
          case entity.ListingError
                .PersistenceLayerError(_) =>
            ApiError.Internal.default
          case _ => ApiError.Internal.default
        },
        listings =>
          dto
            .GetAllListingsResponse(
              listings
                .map(
                  _.into[dto.ListingResponse]
                    .withFieldComputed(_.images, _.images.map(_.url))
                    .transform
                )
            )
      )

  override def getListing(listingId: ListingId): IO[ApiError, dto.ListingResponse] =
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
        _.into[dto.ListingResponse]
          .withFieldComputed(_.images, _.images.map(_.url))
          .transform
      )

  override def createListing(req: dto.CreateListingRequest)
    : IO[ApiError, dto.CreateListingResponse] =
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
    } yield dto.CreateListingResponse(listing.id, listing.createdAt)
}

object ListingHttpHandler {
  val layer: ZLayer[ListingService, Nothing, ListingHandler] = ZLayer
    .derive[ListingHttpHandler]
}
