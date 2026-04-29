package feed.listing.infrastructure.delivery

import java.time.Instant

import io.scalaland.chimney.dsl._
import zio._

import feed.listing.core.ListingConfig
import feed.listing.core.ListingService
import feed.listing.core.entity
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchCriteria
import feed.listing.infrastructure.domain.dto.elastic.PageRequest
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingImageResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.listing.infrastructure.domain.dto.http.createlisting.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.createlisting.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.generateuploadurl.GenerateUploadUrlResponse
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsRequest
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsResponse
import feed.listing.infrastructure.query.ListingImageController
import feed.listing.infrastructure.query.ListingSearchController
import feed.shared.apierror.ApiError

final class ListingHttpHandler(
    listingService: ListingService,
    listingConfig: ListingConfig,
    listingSearchController: ListingSearchController,
    listingImageController: ListingImageController
) extends ListingHandler {
  override def getRecentListings(
      cursor: Option[String],
      limit: Option[Int]
  ): IO[ApiError, GetAllListingsResponse] =
    listingSearchController
      .search(
        ListingSearchCriteria(None, None, None),
        PageRequest(cursor, limit.getOrElse(listingConfig.limit))
      )
      .mapBoth(
        {
          case entity.ListingError
                .PersistenceLayerError(_) =>
            ApiError.Internal.default
          case _ => ApiError.Internal.default
        },
        listingSearchResult =>
          GetAllListingsResponse(
            listingSearchResult.listings.map(
              _.into[ListingResponse]
                .withFieldComputed(
                  _.images,
                  _.images.map(i => ListingImageResponse(i.url, i.position))
                )
                .transform
            ),
            listingSearchResult.cursor
          )
      )

  override def searchListings(req: SearchListingsRequest): IO[ApiError, SearchListingsResponse] =
    listingSearchController
      .search(req.into[ListingSearchCriteria].transform, req.into[PageRequest].transform)
      .mapBoth(
        _ => ApiError.Internal.default,
        result =>
          SearchListingsResponse(
            listings = result.listings.map(_.transformInto[ListingResponse]),
            nextCursor = result.cursor
          )
      )

  override def getListing(listingId: ListingId): IO[ApiError, ListingResponse] =
    listingSearchController
      .getById(listingId)
      .mapBoth(
        {
          case entity.ListingError.PersistenceLayerError(_) => ApiError.Internal.default
          case entity.ListingError.NotFound                 => ApiError.NotFound.listing
          case _                                            => ApiError.Internal.default
        },
        _.transformInto[ListingResponse]
      )

  override def createListing(req: CreateListingRequest): IO[ApiError, CreateListingResponse] =
    for {
      id <- ZIO.succeed(java.util.UUID.randomUUID())
      now <- Clock.instant
      listing = entity.Listing(
        id = id,
        title = req.title,
        description = req.description,
        price = req.price,
        currency = req.currency,
        status = entity.ListingStatus.Active,
        images = req.imageKeys.zipWithIndex.map { case (key, idx) =>
          entity.ListingImage(id = java.util.UUID.randomUUID(), key = key, position = idx)
        },
        createdAt = now,
        updatedAt = now
      )
      _ <- listingService
        .createListing(listing)
        .mapError {
          case entity.ListingError.ValidationError(_)       => ApiError.BadRequest.listingCreate
          case entity.ListingError.PersistenceLayerError(_) => ApiError.Internal.default
          case _                                            => ApiError.Internal.default
        }
    } yield feed.listing.infrastructure.domain.dto.http.createlisting
      .CreateListingResponse(listing.id, listing.createdAt)

  override def generateUploadUrl: IO[ApiError, GenerateUploadUrlResponse] =
    listingImageController.generateUploadUrl(ListingHttpHandler.NotAuthenticatedUserId)
}

object ListingHttpHandler {
  private val NotAuthenticatedUserId = "not-authenticated-user"

  val layer
      : RLayer[ListingService & ListingSearchController & ListingImageController, ListingHandler] =
    ZLayer
      .derive[ListingHttpHandler]
}
