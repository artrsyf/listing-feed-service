import type { ListingResponse } from '../types/listing'
import { Link } from 'react-router-dom'
import './ListingCard.css'

interface ListingCardProps {
  listing: ListingResponse
}

function ListingCard({ listing }: ListingCardProps) {
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
    
    if (diffDays === 0) return 'Сегодня'
    if (diffDays === 1) return 'Вчера'
    if (diffDays < 7) return `${diffDays} дн. назад`
    
    return date.toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'long'
    })
  }

  const formatPrice = (price: number, currency: string) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: currency,
      maximumFractionDigits: 0
    }).format(price)
  }

  return (
    <Link to={`/listing/${listing.id}`} className="listing-card">
      <div className="listing-card__image-wrapper">
        {listing.images.length > 0 ? (
          <img
            src={listing.images[0]}
            alt={listing.title}
            className="listing-card__image"
            loading="lazy"
          />
        ) : (
          <div className="listing-card__image-placeholder">Нет фото</div>
        )}
      </div>
      <div className="listing-card__content">
        <h3 className="listing-card__title">{listing.title}</h3>
        <p className="listing-card__price">
          {formatPrice(listing.price, listing.currency)}
        </p>
        <p className="listing-card__description">{listing.description}</p>
        <span className="listing-card__date">{formatDate(listing.createdAt)}</span>
      </div>
    </Link>
  )
}

export default ListingCard
