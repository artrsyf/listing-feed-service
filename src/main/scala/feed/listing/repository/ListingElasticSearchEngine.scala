package feed.listing.repository

import java.time.Instant

import com.sksamuel.elastic4s.ElasticClient
import com.sksamuel.elastic4s.ElasticDsl._
import com.sksamuel.elastic4s.analysis.Analysis
import com.sksamuel.elastic4s.fields.ElasticField
import com.sksamuel.elastic4s.requests.mappings.dynamictemplate.DynamicMapping
import com.sksamuel.elastic4s.zio.instances._
import io.scalaland.chimney.dsl._
import zio._

import feed.listing.domain.dto.elastic.ElasticPayload
import feed.listing.domain.entity.Listing
import feed.listing.domain.entity.ListingError.PersistenceLayerError
import feed.listing.domain.model.elastic.ElasticListing
import feed.listing.domain.model.elastic.ElasticListingImage
import feed.listing.shared.infrastructure.elastic.ElasticConfig

final class ListingElasticSearchEngine(
  listingSearchConfig: ElasticConfig,
  elasticClient: ElasticClient)
    extends ListingSearchEngine {
  override def insertMany(listings: Chunk[Listing]): IO[PersistenceLayerError, Unit] =
    indexBulk(listings.map { listing =>
      val elasticListing =
        listing
          .into[ElasticListing]
          .withFieldRenamed(_.status, _.status)
          .withFieldComputed(
            _.images,
            _.images.map(img =>
              ElasticListingImage(img.id, listing.id, img.url, img.url, img.position, Instant.now())
            )
          )
          .transform

      ElasticPayload[ElasticListing]("temp_id", 1, elasticListing)
    }).mapError(e => PersistenceLayerError(e.getMessage))

  private def createIndexIfNotExists(
    fields: Seq[ElasticField],
    analysis: Option[Analysis],
    shards: Option[Int] = None
  ) =
    for {
      indexExistsResponse <- elasticClient.execute {
        indexExists(listingSearchConfig.listingIndexName)
      }
      _ <- {
        val createIndexQuery = {
          val req = createIndex(listingSearchConfig.listingIndexName)

          val reqWithAnalisis = analysis
            .fold(req)(req.analysis(_))
            .mapping(properties(fields).dynamic(DynamicMapping.Strict))

          shards.fold(reqWithAnalisis)(reqWithAnalisis.shards)
        }

        elasticClient.execute(createIndexQuery)
      }.unless(indexExistsResponse.result.exists)
    } yield ()

  private def indexBulk(esPayloads: Chunk[ElasticPayload[ElasticListing]]): Task[Unit] =
    elasticClient.execute {
      bulk(esPayloads.toList.map(_.toElasticRequest(listingSearchConfig.listingIndexName)))
    }.unit
}

object ListingElasticSearchEngine {
  val layer: RLayer[ElasticClient, ListingElasticSearchEngine] =
    ZLayer.derive[ListingElasticSearchEngine]
}
