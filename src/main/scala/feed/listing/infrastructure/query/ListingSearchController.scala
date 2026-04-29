package feed.listing.infrastructure.query

import io.scalaland.chimney.dsl._
import io.scalaland.chimney.syntax
import zio._

import feed.listing.core.entity.Listing
import feed.listing.core.entity.ListingError
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.elastic.EvaluatedListing
import feed.listing.infrastructure.domain.dto.elastic.EvaluatedListingImage
import feed.listing.infrastructure.domain.dto.elastic.EvaluatedListingSearchResult
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchCriteria
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchResult
import feed.listing.infrastructure.domain.dto.elastic.PageRequest
import feed.listing.infrastructure.domain.dto.http.ListingResponse
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsResponse
import feed.shared.apierror.ApiError.NotFound.listing

final class ListingSearchController(
    listingSearchEngine: ListingSearchReadEngine,
    listingObjectStorage: ObjectStorage
) {
  def search(
      criteria: ListingSearchCriteria,
      page: PageRequest
  ): IO[ListingError, EvaluatedListingSearchResult] =
    for {
      listingsSearchResult <- listingSearchEngine
        .searchListings(criteria, page)
      evaluatedListings <- ZIO.foreach(listingsSearchResult.listings)(listing =>
        evaluateListing(listing)
      )
    } yield EvaluatedListingSearchResult(
      listings = evaluatedListings,
      cursor = listingsSearchResult.cursor
    )

  def getById(id: ListingId): IO[ListingError, EvaluatedListing] =
    for {
      listing <- listingSearchEngine.getById(id).someOrFail(ListingError.NotFound)
      evaluated <- evaluateListing(listing)
    } yield evaluated

  private def evaluateListing(listing: Listing) =
    for {
      evaluatedImages <- ZIO.foreach(listing.images)(image =>
        listingObjectStorage
          .getPublicUrl(image.key)
          .map(url => EvaluatedListingImage(url = url, position = image.position))
      )
    } yield listing
      .into[EvaluatedListing]
      .withFieldConst(_.images, evaluatedImages)
      .transform

}

object ListingSearchController {
  val layer: RLayer[ListingSearchReadEngine & ObjectStorage, ListingSearchController] =
    ZLayer.derive[ListingSearchController]
}
