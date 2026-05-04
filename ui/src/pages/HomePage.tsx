import { useState, useEffect, useCallback, useRef } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/listingApi'
import type { ListingResponse } from '../types/listing'
import ListingCard from '../components/ListingCard'
import ListingGrid from '../components/ListingGrid'
import InfiniteScroll from '../components/InfiniteScroll'
import Loader from '../components/Loader'
import SearchBar from '../components/SearchBar'
import './HomePage.css'

const PAGE_SIZE = 20

function HomePage() {
  const navigate = useNavigate()
  const [listings, setListings] = useState<ListingResponse[]>([])
  const [cursor, setCursor] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(true)
  const [isLoading, setIsLoading] = useState(false)
  const [isInitialLoad, setIsInitialLoad] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState<string | null>(null)
  const hasLoadedRef = useRef(false)

  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore) return

    setIsLoading(true)
    setError(null)

    try {
      let response
      if (searchQuery) {
        response = await api.searchListings({
          query: searchQuery,
          cursor: cursor ?? undefined,
          limit: PAGE_SIZE
        })
      } else {
        response = await api.getListings(PAGE_SIZE, cursor ?? undefined)
      }

      // Если получили пустой список - прекращаем загрузку
      if (response.listings.length === 0) {
        setHasMore(false)
        return
      }

      setListings(prev => [...prev, ...response.listings])
      setCursor('nextCursor' in response ? response.nextCursor : response.cursor)
      // Если курсор null или список меньше запрошенного размера - больше данных нет
      const resultCursor = 'nextCursor' in response ? response.nextCursor : response.cursor
      setHasMore(resultCursor !== null && response.listings.length === PAGE_SIZE)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка загрузки')
    } finally {
      setIsLoading(false)
      setIsInitialLoad(false)
    }
  }, [cursor, isLoading, hasMore, searchQuery])

  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query)
    setListings([])
    setCursor(null)
    setHasMore(true)
    hasLoadedRef.current = false
    setIsInitialLoad(true)
    // Не устанавливаем isLoading(true) - это сделает loadMore()
  }, [])

  const handleNavigateToListing = useCallback((id: string) => {
    navigate(`/listing/${id}`)
  }, [navigate])

  const handleClearSearch = useCallback(() => {
    setSearchQuery(null)
    setListings([])
    setCursor(null)
    setHasMore(true)
    hasLoadedRef.current = false
    setIsInitialLoad(true)
  }, [])

  useEffect(() => {
    if (listings.length === 0 && !isLoading && !hasLoadedRef.current) {
      hasLoadedRef.current = true
      loadMore()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchQuery])

  if (isInitialLoad && isLoading) {
    return (
      <div className="home-page">
        <header className="home-page__header">
          <h1 className="home-page__title">Лента объявлений</h1>
          <div className="home-page__actions">
            <SearchBar onSearch={handleSearch} />
            <Link to="/create" className="home-page__create-btn">
              Подать объявление
            </Link>
          </div>
        </header>
        <Loader />
      </div>
    )
  }

  return (
    <div className="home-page">
      <header className="home-page__header">
        <h1 className="home-page__title">Лента объявлений</h1>
        <div className="home-page__actions">
          <SearchBar onSearch={handleSearch} onNavigateToListing={handleNavigateToListing} />
          <Link to="/create" className="home-page__create-btn">
            Подать объявление
          </Link>
        </div>
      </header>

      {searchQuery && (
        <div className="home-page__search-info">
          <span>Результаты поиска по запросу: <strong>"{searchQuery}"</strong></span>
          <button onClick={handleClearSearch} className="home-page__clear-search">
            Показать все объявления
          </button>
        </div>
      )}

      {error && (
        <div className="home-page__error">
          <p>Ошибка: {error}</p>
          <button onClick={loadMore} className="home-page__retry-btn">
            Повторить
          </button>
        </div>
      )}

      {!error && listings.length === 0 && !isLoading && (
        <div className="home-page__empty">
          <p>Объявлений не найдено</p>
          {searchQuery && (
            <button onClick={handleClearSearch} className="home-page__clear-search-btn">
              Сбросить поиск
            </button>
          )}
        </div>
      )}

      <ListingGrid>
        <InfiniteScroll
          hasMore={hasMore}
          isLoading={isLoading}
          onLoadMore={loadMore}
          listingsCount={listings.length}
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
