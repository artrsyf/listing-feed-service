import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { api } from '../api/listingApi'
import type { ListingResponse } from '../types/listing'
import Loader from '../components/Loader'
import './ListingPage.css'

function ListingPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [listing, setListing] = useState<ListingResponse | null>(null)
  const [selectedImageIndex, setSelectedImageIndex] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) {
      navigate('/')
      return
    }

    const fetchListing = async () => {
      setIsLoading(true)
      setError(null)

      try {
        const data = await api.getListing(id)
        setListing(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка загрузки')
      } finally {
        setIsLoading(false)
      }
    }

    fetchListing()
  }, [id, navigate])

  const formatPrice = (price: number, currency: string) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: currency,
      maximumFractionDigits: 0
    }).format(price)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  if (isLoading) {
    return (
      <div className="listing-page">
        <Loader />
      </div>
    )
  }

  if (error || !listing) {
    return (
      <div className="listing-page">
        <div className="listing-page__error">
          <h2>Ошибка</h2>
          <p>{error || 'Объявление не найдено'}</p>
          <Link to="/" className="listing-page__back-btn">
            На главную
          </Link>
        </div>
      </div>
    )
  }

  const hasMultipleImages = listing.images.length > 1

  return (
    <div className="listing-page">
      <nav className="listing-page__nav">
        <Link to="/" className="listing-page__back-link">
          ← Назад к ленте
        </Link>
      </nav>

      <main className="listing-page__content">
        <div className="listing-page__gallery">
          {listing.images.length > 0 ? (
            <>
              <img
                src={listing.images[selectedImageIndex]?.url}
                alt={`${listing.title} - фото ${selectedImageIndex + 1}`}
                className="listing-page__main-image"
              />
              
              {hasMultipleImages && (
                <div className="listing-page__thumbnails">
                  {listing.images.map((image, index) => (
                    <button
                      key={image.position}
                      onClick={() => setSelectedImageIndex(index)}
                      className={`listing-page__thumbnail ${
                        index === selectedImageIndex ? 'listing-page__thumbnail--active' : ''
                      }`}
                    >
                      <img src={image.url} alt={`Фото ${index + 1}`} />
                    </button>
                  ))}
                </div>
              )}
            </>
          ) : (
            <div className="listing-page__no-image">Нет фото</div>
          )}
        </div>

        <div className="listing-page__info">
          <h1 className="listing-page__title">{listing.title}</h1>
          
          <p className="listing-page__price">
            {formatPrice(listing.price, listing.currency)}
          </p>

          <div className="listing-page__meta">
            <span className="listing-page__date">
              Опубликовано: {formatDate(listing.createdAt)}
            </span>
            <span className="listing-page__id">ID: {listing.id.slice(0, 8)}</span>
          </div>

          <div className="listing-page__description">
            <h2>Описание</h2>
            <p>{listing.description}</p>
          </div>

          <div className="listing-page__actions">
            <button className="listing-page__contact-btn">
              Показать контакты
            </button>
          </div>
        </div>
      </main>
    </div>
  )
}

export default ListingPage
