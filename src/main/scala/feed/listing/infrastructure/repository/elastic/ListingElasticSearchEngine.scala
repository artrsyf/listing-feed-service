package feed.listing.infrastructure.repository.elastic

import java.time.Instant
import java.util.Base64

import com.sksamuel.elastic4s.ElasticClient
import com.sksamuel.elastic4s.ElasticDsl._
import com.sksamuel.elastic4s.analysis.Analysis
import com.sksamuel.elastic4s.ext.OptionImplicits.RichOptionImplicits
import com.sksamuel.elastic4s.fields.ElasticField
import com.sksamuel.elastic4s.requests.mappings.dynamictemplate.DynamicMapping
import com.sksamuel.elastic4s.requests.searches.sort.FieldSort
import com.sksamuel.elastic4s.zio.instances._
import io.scalaland.chimney.dsl._
import zio._
import zio.json._

import feed.listing.core.ListingSearchCreateEngine
import feed.listing.core.entity.Listing
import feed.listing.core.entity.ListingError.PersistenceLayerError
import feed.listing.core.entity.ListingId
import feed.listing.infrastructure.domain.dto.elastic.ElasticPayload
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchCriteria
import feed.listing.infrastructure.domain.dto.elastic.ListingSearchResult
import feed.listing.infrastructure.domain.dto.elastic.PageRequest
import feed.listing.infrastructure.domain.dto.http.searchlistings.SearchListingsRequest
import feed.listing.infrastructure.domain.model.elastic.ElasticListing
import feed.listing.infrastructure.domain.model.elastic.ElasticListingImage
import feed.listing.infrastructure.query.ListingSearchReadEngine

final class ListingElasticSearchEngine(
    listingSearchConfig: ElasticConfig,
    elasticClient: ElasticClient
) extends ListingSearchReadEngine
    with ListingSearchCreateEngine {
  override def insertMany(listings: Chunk[Listing]): IO[PersistenceLayerError, Unit] =
    indexBulk(listings.map { listing =>
      val elasticListing =
        listing
          .into[ElasticListing]
          .withFieldRenamed(_.status, _.status)
          .withFieldComputed(
            _.images,
            _.images.map(img =>
              ElasticListingImage(img.id, listing.id, img.key, img.position, Instant.now())
            )
          )
          .transform

      ElasticPayload[ElasticListing](listing.id.toString, 1, elasticListing)
    }).mapError(e => PersistenceLayerError(e.getMessage))

  override def searchListings(
      criteria: ListingSearchCriteria,
      page: PageRequest
  ): IO[PersistenceLayerError, ListingSearchResult] = {
    val baseQuery =
      boolQuery()
        .must(criteria.query.map(q => matchQuery("title", q)).toList)
        .filter(
          List(
            criteria.minPrice.map(p => rangeQuery("price").gte(p.toFloat)),
            criteria.maxPrice.map(p => rangeQuery("price").lte(p.toFloat))
          ).flatten
        )

    val searchRequest =
      search(listingSearchConfig.listingIndexName)
        .sortBy(FieldSort("createdAt").desc(), FieldSort("id").desc())
        .query(baseQuery)
        .size(page.limit)

    val withCursor =
      page.cursor match {
        case Some(c) =>
          val decoded = decodeCursor(c)
          searchRequest.searchAfter(decoded)
        case None =>
          searchRequest
      }

    elasticClient.execute(withCursor).flatMap { resp =>
      val hits = resp.result.hits.hits

      for {
        listings <- ZIO.foreach(hits) { hit =>
          ZIO.logInfo(hits.toString) *>
            ZIO
              .fromEither(hit.sourceAsString.fromJson[ElasticListing])
              .mapError(err => new RuntimeException(s"Failed to decode ES document: $err"))
        }
      } yield {
        val nextCursor = hits.lastOption.flatMap(_.sort)
        val nextEncodedCursor = nextCursor.map(encodeCursor)

        ListingSearchResult(listings.toList.map(_.transformInto[Listing]), nextEncodedCursor)
      }
    }
  }.mapError(e => PersistenceLayerError(e.getMessage))

  case class Cursor(values: List[String])

  object Cursor {
    implicit val jsonCodec: JsonCodec[Cursor] = DeriveJsonCodec.gen[Cursor]
  }

  def encodeCursor(v: Seq[AnyRef]): String =
    Base64.getEncoder.encodeToString(Cursor(v.map(_.toString).toList).toJson.getBytes)

  def decodeCursor(c: String): List[String] =
    new String(Base64.getDecoder.decode(c)).fromJson[Cursor].toOption.get.values

  override def getById(id: ListingId): IO[PersistenceLayerError, Option[Listing]] =
    elasticClient
      .execute {
        get(listingSearchConfig.listingIndexName, id.toString())
      }
      .flatMap { resp =>
        if (!resp.result.found) ZIO.none
        else
          ZIO
            .fromEither(resp.result.sourceAsString.fromJson[ElasticListing])
            .map(_.transformInto[Listing])
            .map(Some(_))
            .mapError(err => new RuntimeException(s"Failed to decode ES document: $err"))
      }
      .mapError(e => PersistenceLayerError(e.getMessage))

  private def createIndexIfNotExists(
      fields: Seq[ElasticField],
      analysis: Option[Analysis],
      shards: Option[Int] = None
  ) =
    for {
      indexExistsResponse <- elasticClient.execute {
        indexExists(listingSearchConfig.listingIndexName)
      }
      res <- {
        val createIndexQuery = {
          val req = createIndex(listingSearchConfig.listingIndexName)

          val reqWithAnalisis = analysis
            .fold(req)(req.analysis(_))
            .mapping(properties(fields).dynamic(DynamicMapping.Strict))

          shards.fold(reqWithAnalisis)(reqWithAnalisis.shards)
        }

        elasticClient.execute(createIndexQuery)
      }.unless(indexExistsResponse.result.exists)
    } yield res

  private def indexBulk(esPayloads: Chunk[ElasticPayload[ElasticListing]]): Task[Unit] =
    elasticClient.execute {
      bulk(esPayloads.toList.map(_.toElasticRequest(listingSearchConfig.listingIndexName)))
    }.unit
}

object ListingElasticSearchEngine {
  val layer: RLayer[ElasticClient, ListingElasticSearchEngine] =
    ZLayer.fromZIO {
      for {
        client <- ZIO.service[ElasticClient]
        config <- ZIO.config[ElasticConfig]

        engine = new ListingElasticSearchEngine(config, client)

        resp <- engine
          .createIndexIfNotExists(fields = ElasticListing.listingFields, analysis = None)
          .retry(Schedule.exponential(200.millis) && Schedule.recurs(5))
        _ <- ZIO.foreach(resp) { esResp =>
          ZIO
            .fail(new RuntimeException(esResp.error.reason))
            .when(esResp.isError)
            .someOrElseZIO(ZIO.logInfo("Successfully created ES index"))
        }
      } yield engine
    }
}
