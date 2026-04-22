package feed.listing.repository

import com.sksamuel.elastic4s.ElasticApi.indexInto
import com.sksamuel.elastic4s.ElasticClient
import com.sksamuel.elastic4s.ElasticDsl._
import com.sksamuel.elastic4s.Indexable
import com.sksamuel.elastic4s.analysis.Analysis
import com.sksamuel.elastic4s.fields.ElasticField
import com.sksamuel.elastic4s.requests.common.VersionType.ExternalGte
import com.sksamuel.elastic4s.requests.mappings.dynamictemplate.DynamicMapping
import com.sksamuel.elastic4s.zio.instances._
import zio._

import feed.listing.domain.model.ElasticListing

sealed trait ListingSearchRepository {
  def indexBulk(esPayloads: Chunk[ElasticPayload[ElasticListing]]): Task[Unit]
}

final class ListingElasticRepository(elasticClient: ElasticClient) extends ListingSearchRepository {
  private val indexName = "MyIndexName" // TODO: Вынести в конфиг

  private def createIndexIfNotExists(
    fields: Seq[ElasticField],
    analysis: Option[Analysis],
    shards: Option[Int] = None
  ) =
    for {
      indexExistsResponse <- elasticClient.execute {
        indexExists(indexName)
      }
      _ <- {
        val createIndexQuery = {
          val req = createIndex(indexName)

          val reqWithAnalisis = analysis
            .fold(req)(req.analysis(_))
            .mapping(properties(fields).dynamic(DynamicMapping.Strict))

          shards.fold(reqWithAnalisis)(reqWithAnalisis.shards)
        }

        elasticClient.execute(createIndexQuery)
      }.unless(indexExistsResponse.result.exists)
    } yield ()

  override def indexBulk(esPayloads: Chunk[ElasticPayload[ElasticListing]]): Task[Unit] =
    elasticClient.execute {
      bulk(esPayloads.toList.map(_.toElasticRequest(indexName)))
    }.unit
}

object ListingElasticRepository {
  // TODO: Посмотреть почему derive не вывозит
  val layer: URLayer[ElasticClient, ListingElasticRepository] =
    ZLayer.fromFunction(new ListingElasticRepository(_))
}

final case class ElasticPayload[T: Indexable](
  _id: String,
  version: Long = 1L,
  payload: T) {
  def toElasticRequest(indexName: String) =
    indexInto(indexName)
      .id(_id)
      .doc(payload)
      .versionType(ExternalGte)
      .version(version)
}
