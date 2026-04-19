package feed.listing.route

import sttp.tapir.ztapir.ZServerEndpoint
import zio.Chunk

trait ListingRoute {
  def routes: Chunk[ZServerEndpoint[Any, Any]]
}
