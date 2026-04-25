package feed.listing.infrastructure.repository.elastic

import com.sksamuel.elastic4s.ElasticProperties
import com.sksamuel.elastic4s.http.JavaClient
import com.sksamuel.elastic4s.{ElasticClient => ElasticClient4s}
import zio.ZIO
import zio.ZLayer

object ElasticClient {
  val live: ZLayer[Any, Throwable, ElasticClient4s] =
    ZLayer.scoped {
      ZIO.acquireRelease {
        for {
          config <- ZIO.config[ElasticConfig]
          elasticClient <- ZIO.attempt {
            val props = ElasticProperties(config.url)
            ElasticClient4s(JavaClient(props))
          }
        } yield elasticClient
      } { client =>
        ZIO.attempt(client.close()).orDie
      }
    }
}
