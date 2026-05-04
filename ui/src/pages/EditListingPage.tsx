import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { api } from '../api/listingApi'
import type { ListingResponse, UpdateListingRequest } from '../types/listing'
import Loader from '../components/Loader'
import './EditListingPage.css'

function EditListingPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [listing, setListing] = useState<ListingResponse | null>(null)
  const [formData, setFormData] = useState<UpdateListingRequest>({})
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
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
        setFormData({
          title: data.title,
          description: data.description,
          price: data.price,
          currency: data.currency
        })
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Ошибка загрузки')
      } finally {
        setIsLoading(false)
      }
    }

    fetchListing()
  }, [id, navigate])

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: name === 'price' ? parseFloat(value) || 0 : value
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSaving(true)
    setError(null)

    try {
      if (!id) return
      
      await api.updateListing(id, formData)
      navigate(`/listing/${id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
      setIsSaving(false)
    }
  }

  if (isLoading) {
    return (
      <div className="edit-listing-page">
        <Loader />
      </div>
    )
  }

  if (error || !listing) {
    return (
      <div className="edit-listing-page">
        <div className="edit-listing-page__error">
          <h2>Ошибка</h2>
          <p>{error || 'Объявление не найдено'}</p>
          <Link to="/" className="edit-listing-page__back-btn">
            На главную
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="edit-listing-page">
      <nav className="edit-listing-page__nav">
        <Link to={`/listing/${id}`} className="edit-listing-page__back-link">
          ← Назад к объявлению
        </Link>
      </nav>

      <main className="edit-listing-page__content">
        <h1 className="edit-listing-page__title">Редактировать объявление</h1>

        {error && (
          <div className="edit-listing-page__error-inline">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="edit-listing-page__form">
          <div className="edit-listing-page__field">
            <label htmlFor="title" className="edit-listing-page__label">
              Название *
            </label>
            <input
              type="text"
              id="title"
              name="title"
              value={formData.title || ''}
              onChange={handleChange}
              required
              className="edit-listing-page__input"
            />
          </div>

          <div className="edit-listing-page__field">
            <label htmlFor="price" className="edit-listing-page__label">
              Цена *
            </label>
            <input
              type="number"
              id="price"
              name="price"
              value={formData.price || 0}
              onChange={handleChange}
              required
              min="0"
              step="0.01"
              className="edit-listing-page__input"
            />
          </div>

          <div className="edit-listing-page__field">
            <label htmlFor="currency" className="edit-listing-page__label">
              Валюта *
            </label>
            <select
              id="currency"
              name="currency"
              value={formData.currency || 'RUB'}
              onChange={handleChange}
              required
              className="edit-listing-page__input"
            >
              <option value="RUB">RUB - Российский рубль</option>
              <option value="USD">USD - Доллар США</option>
              <option value="EUR">EUR - Евро</option>
            </select>
          </div>

          <div className="edit-listing-page__field">
            <label htmlFor="description" className="edit-listing-page__label">
              Описание *
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description || ''}
              onChange={handleChange}
              required
              rows={5}
              className="edit-listing-page__textarea"
            />
          </div>

          <div className="edit-listing-page__actions">
            <button
              type="submit"
              disabled={isSaving}
              className="edit-listing-page__submit-btn"
            >
              {isSaving ? 'Сохранение...' : 'Сохранить изменения'}
            </button>
          </div>
        </form>
      </main>
    </div>
  )
}

export default EditListingPage
