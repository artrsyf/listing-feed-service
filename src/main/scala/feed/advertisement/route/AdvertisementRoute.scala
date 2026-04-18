package feed.advertisement.route

import sttp.tapir.ztapir.ZServerEndpoint
import zio.Chunk

trait AdvertisementRoute {
  def routes: Chunk[ZServerEndpoint[Any, Any]]
}
