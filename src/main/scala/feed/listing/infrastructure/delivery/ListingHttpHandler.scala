package feed.listing.infrastructure.delivery

import java.time.Instant

import io.scalaland.chimney.dsl._
import zio._

import feed.listing.core.ListingService
import feed.listing.core.entity
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchCriteria
import feed.listing.infrastructure.domain.dto.elastic.PageRequest
import feed.listing.infrastructure.domain.dto.http.CreateListingRequest
import feed.listing.infrastructure.domain.dto.http.CreateListingResponse
import feed.listing.infrastructure.domain.dto.http.GetAllListingsResponse
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.listing.infrastructure.domain.dto.http.SearchListingsRequest
import feed.listing.infrastructure.domain.dto.http.SearchListingsResponse
import feed.listing.infrastructure.query.ListingSearchController
import feed.shared.apierror.ApiError

final class ListingHttpHandler(
    listingService: ListingService,
    listingConfig: ListingConfig,
    listingSearchController: ListingSearchController
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
                .withFieldComputed(_.images, _.images.map(_.url))
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
            listings = result.listings.map(
              _.into[ListingResponse]
                .withFieldComputed(_.images, _.images.map(_.url))
                .transform
            ),
            nextCursor = result.cursor
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
      id <- ZIO.succeed(java.util.UUID.randomUUID())
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
  val layer: RLayer[ListingService & ListingSearchController, ListingHandler] = ZLayer
    .derive[ListingHttpHandler]
}
