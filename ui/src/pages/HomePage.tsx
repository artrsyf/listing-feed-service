import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/listingApi'
import type { ListingResponse } from '../types/listing'
import ListingCard from '../components/ListingCard'
import ListingGrid from '../components/ListingGrid'
import InfiniteScroll from '../components/InfiniteScroll'
import Loader from '../components/Loader'
import './HomePage.css'

const PAGE_SIZE = 20

function HomePage() {
  const [listings, setListings] = useState<ListingResponse[]>([])
  const [cursor, setCursor] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(true)
  const [isLoading, setIsLoading] = useState(false)
  const [isInitialLoad, setIsInitialLoad] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore) return

    setIsLoading(true)
    setError(null)

    try {
      const response = await api.getListings(PAGE_SIZE, cursor ?? undefined)
      
      setListings(prev => [...prev, ...response.listings])
      setCursor(response.cursor)
      setHasMore(response.cursor !== null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки')
    } finally {
      setIsLoading(false)
      setIsInitialLoad(false)
    }
  }, [cursor, isLoading, hasMore])

  useEffect(() => {
    if (listings.length === 0 && !isLoading) {
      loadMore()
    }
  }, [])

  if (isInitialLoad && isLoading) {
    return (
      <div className="home-page">
        <header className="home-page__header">
          <h1 className="home-page__title">Лента объявлений</h1>
          <Link to="/create" className="home-page__create-btn">
            Подать объявление
          </Link>
        </header>
        <Loader />
      </div>
    )
  }

  return (
    <div className="home-page">
      <header className="home-page__header">
        <h1 className="home-page__title">Лента объявлений</h1>
        <Link to="/create" className="home-page__create-btn">
          Подать объявление
        </Link>
      </header>

      {error && (
        <div className="home-page__error">
          <p>Ошибка: {error}</p>
          <button onClick={loadMore} className="home-page__retry-btn">
            Повторить
          </button>
        </div>
      )}

      <ListingGrid>
        <InfiniteScroll
          hasMore={hasMore}
          isLoading={isLoading}
          onLoadMore={loadMore}
        >
          {listings.map(listing => (
            <ListingCard key={listing.id} listing={listing} />
          ))}
        </InfiniteScroll>
      </ListingGrid>
    </div>
  )
}

export default HomePage
