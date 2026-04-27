import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/listingApi'
import type { CreateListingRequest } from '../types/listing'
import './CreateListingPage.css'

function CreateListingPage() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState<CreateListingRequest>({
    title: '',
    description: '',
    price: 0,
    currency: 'RUB',
    imageUrls: ['']
  })
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: name === 'price' ? parseFloat(value) || 0 : value
    }))
  }

  const handleImageChange = (index: number, value: string) => {
    const newImages = [...formData.imageUrls]
    newImages[index] = value
    setFormData(prev => ({ ...prev, imageUrls: newImages }))
  }

  const addImageField = () => {
    setFormData(prev => ({ ...prev, imageUrls: [...prev.imageUrls, ''] }))
  }

  const removeImageField = (index: number) => {
    if (formData.imageUrls.length === 1) return
    setFormData(prev => ({
      ...prev,
      imageUrls: prev.imageUrls.filter((_, i) => i !== index)
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    try {
      // Фильтруем пустые URL
      const dataToSubmit = {
        ...formData,
        imageUrls: formData.imageUrls.filter(url => url.trim() !== '')
      }

      if (dataToSubmit.imageUrls.length === 0) {
        setError('Добавьте хотя бы одно изображение')
        setIsLoading(false)
        return
      }

      const response = await api.createListing(dataToSubmit)
      
      // Перенаправляем на страницу созданного объявления
      navigate(`/listing/${response.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка создания объявления')
      setIsLoading(false)
    }
  }

  return (
    <div className="create-listing-page">
      <nav className="create-listing-page__nav">
        <Link to="/" className="create-listing-page__back-link">
          ← На главную
        </Link>
      </nav>

      <main className="create-listing-page__content">
        <h1 className="create-listing-page__title">Подать объявление</h1>

        {error && (
          <div className="create-listing-page__error">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="create-listing-page__form">
          <div className="create-listing-page__field">
            <label htmlFor="title" className="create-listing-page__label">
              Название *
            </label>
            <input
              type="text"
              id="title"
              name="title"
              value={formData.title}
              onChange={handleChange}
              required
              placeholder="Например: iPhone 15 Pro Max"
              className="create-listing-page__input"
            />
          </div>

          <div className="create-listing-page__field">
            <label htmlFor="price" className="create-listing-page__label">
              Цена *
            </label>
            <input
              type="number"
              id="price"
              name="price"
              value={formData.price}
              onChange={handleChange}
              required
              min="0"
              step="0.01"
              placeholder="0"
              className="create-listing-page__input"
            />
          </div>

          <div className="create-listing-page__field">
            <label htmlFor="currency" className="create-listing-page__label">
              Валюта *
            </label>
            <select
              id="currency"
              name="currency"
              value={formData.currency}
              onChange={handleChange}
              required
              className="create-listing-page__input"
            >
              <option value="RUB">RUB - Российский рубль</option>
              <option value="USD">USD - Доллар США</option>
              <option value="EUR">EUR - Евро</option>
            </select>
          </div>

          <div className="create-listing-page__field">
            <label htmlFor="description" className="create-listing-page__label">
              Описание *
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              required
              rows={5}
              placeholder="Опишите ваш товар..."
              className="create-listing-page__textarea"
            />
          </div>

          <div className="create-listing-page__field">
            <label className="create-listing-page__label">
              Изображения (URL) *
            </label>
            {formData.imageUrls.map((url, index) => (
              <div key={index} className="create-listing-page__image-field">
                <input
                  type="url"
                  value={url}
                  onChange={(e) => handleImageChange(index, e.target.value)}
                  placeholder={`https://example.com/image${index + 1}.jpg`}
                  className="create-listing-page__input"
                />
                {formData.imageUrls.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeImageField(index)}
                    className="create-listing-page__remove-btn"
                  >
                    ✕
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={addImageField}
              className="create-listing-page__add-image-btn"
            >
              + Добавить ещё изображение
            </button>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="create-listing-page__submit-btn"
          >
            {isLoading ? 'Публикация...' : 'Опубликовать'}
          </button>
        </form>
      </main>
    </div>
  )
}

export default CreateListingPage
