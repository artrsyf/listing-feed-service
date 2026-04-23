package feed.listing.infrastructure.domain.dto.elastic

import com.sksamuel.elastic4s.ElasticApi.indexInto
import com.sksamuel.elastic4s.Indexable
import com.sksamuel.elastic4s.requests.common.VersionType.ExternalGte

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
