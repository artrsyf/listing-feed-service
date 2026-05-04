import { useState, useEffect, useRef, useCallback } from 'react'
import { api } from '../api/listingApi'
import type { ListingResponse } from '../types/listing'
import './SearchBar.css'

const DEBOUNCE_DELAY = 300
const SUGGESTIONS_LIMIT = 5

interface SearchBarProps {
  onSearch: (query: string) => void
  onNavigateToListing?: (id: string) => void
}

function SearchBar({ onSearch, onNavigateToListing }: SearchBarProps) {
  const [query, setQuery] = useState('')
  const [suggestions, setSuggestions] = useState<ListingResponse[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  const wrapperRef = useRef<HTMLDivElement>(null)
  const debounceRef = useRef<number | null>(null)

  // Закрытие при клике вне
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  // Debounce поиск
  useEffect(() => {
    if (debounceRef.current) {
      window.clearTimeout(debounceRef.current)
    }

    if (!query.trim()) {
      setSuggestions([])
      setIsOpen(false)
      return
    }

    debounceRef.current = window.setTimeout(async () => {
      setIsLoading(true)
      try {
        const response = await api.searchListings({
          query: query.trim(),
          limit: SUGGESTIONS_LIMIT
        })
        setSuggestions(response.listings)
        setIsOpen(true)
      } catch (error) {
        console.error('Search error:', error)
        setSuggestions([])
      } finally {
        setIsLoading(false)
      }
    }, DEBOUNCE_DELAY)

    return () => {
      if (debounceRef.current) {
        window.clearTimeout(debounceRef.current)
      }
    }
  }, [query])

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      onSearch(query.trim())
      setIsOpen(false)
    }
  }, [query, onSearch])

  const handleSuggestionClick = (suggestion: ListingResponse) => {
    if (onNavigateToListing) {
      onNavigateToListing(suggestion.id)
    } else {
      // Fallback: поиск по названию
      setQuery(suggestion.title)
      onSearch(suggestion.title)
    }
    setIsOpen(false)
  }

  const handleClear = () => {
    setQuery('')
    setSuggestions([])
    setIsOpen(false)
  }

  return (
    <div className="search-bar-wrapper" ref={wrapperRef}>
      <form onSubmit={handleSubmit} className="search-bar">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => query.trim() && setIsOpen(true)}
          placeholder="Поиск объявлений..."
          className="search-bar__input"
          autoComplete="off"
        />
        {query && (
          <button
            type="button"
            onClick={handleClear}
            className="search-bar__clear"
            aria-label="Очистить"
          >
            ✕
          </button>
        )}
        <button type="submit" className="search-bar__submit" aria-label="Найти">
          🔍
        </button>
      </form>

      {(isOpen || isLoading) && (
        <div className="search-bar__suggestions">
          {isLoading ? (
            <div className="search-bar__loading">
              <div className="search-bar__spinner"></div>
              <span>Поиск...</span>
            </div>
          ) : suggestions.length > 0 ? (
            <ul className="search-bar__list">
              {suggestions.map((suggestion) => (
                <li
                  key={suggestion.id}
                  onClick={() => handleSuggestionClick(suggestion)}
                  className="search-bar__item"
                >
                  <div className="search-bar__item-image">
                    {suggestion.images[0]?.url ? (
                      <img src={suggestion.images[0].url} alt="" />
                    ) : (
                      <div className="search-bar__item-no-image">Нет фото</div>
                    )}
                  </div>
                  <div className="search-bar__item-content">
                    <div className="search-bar__item-title">{suggestion.title}</div>
                    <div className="search-bar__item-price">
                      {new Intl.NumberFormat('ru-RU', {
                        style: 'currency',
                        currency: suggestion.currency,
                        maximumFractionDigits: 0
                      }).format(suggestion.price)}
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          ) : query.trim() ? (
            <div className="search-bar__no-results">
              Ничего не найдено по запросу "{query}"
            </div>
          ) : null}
        </div>
      )}
    </div>
  )
}

export default SearchBar
