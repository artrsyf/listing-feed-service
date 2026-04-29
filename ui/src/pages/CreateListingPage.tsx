import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/listingApi'
import type { CreateListingRequest } from '../types/listing'
import './CreateListingPage.css'

interface UploadedImage {
  key: string
  previewUrl: string
  file: File
  isUploading: boolean
  uploadError?: string
}

function CreateListingPage() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    price: 0,
    currency: 'RUB'
  })
  const [images, setImages] = useState<UploadedImage[]>([])
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

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    if (files.length === 0) return

    const newImages: UploadedImage[] = files.map(file => ({
      key: '',
      previewUrl: URL.createObjectURL(file),
      file,
      isUploading: false
    }))

    setImages(prev => [...prev, ...newImages])

    // Загружаем каждое изображение
    for (const image of newImages) {
      await uploadImage(image)
    }

    // Очищаем input
    e.target.value = ''
  }

  const uploadImage = async (image: UploadedImage) => {
    setImages(prev => prev.map(img =>
      img.previewUrl === image.previewUrl
        ? { ...img, isUploading: true, uploadError: undefined }
        : img
    ))

    try {
      // Шаг 1: Получаем upload-url от бэкенда
      const { key, uploadUrl } = await api.generateUploadUrl({
        contentType: image.file.type
      })

      // Шаг 2: Загружаем файл в Minio
      await api.uploadFile(uploadUrl, image.file)

      // Шаг 3: Сохраняем ключ
      setImages(prev => prev.map(img =>
        img.previewUrl === image.previewUrl
          ? { ...img, key, isUploading: false }
          : img
      ))
    } catch (err) {
      setImages(prev => prev.map(img =>
        img.previewUrl === image.previewUrl
          ? {
            ...img,
            isUploading: false,
            uploadError: err instanceof Error ? err.message : 'Ошибка загрузки'
          }
          : img
      ))
    }
  }

  const removeImage = (previewUrl: string) => {
    setImages(prev => prev.filter(img => img.previewUrl !== previewUrl))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    setError(null)

    try {
      // Проверяем, что все изображения загружены
      const uploadedImages = images.filter(img => img.key && !img.isUploading)

      if (uploadedImages.length === 0) {
        setError('Добавьте хотя бы одно изображение')
        setIsLoading(false)
        return
      }

      // Проверяем, есть ли незагруженные изображения
      const hasUploadErrors = images.some(img => img.uploadError)
      if (hasUploadErrors) {
        setError('Не все изображения загружены. Удалите файлы с ошибками.')
        setIsLoading(false)
        return
      }

      const dataToSubmit: CreateListingRequest = {
        title: formData.title,
        description: formData.description,
        price: formData.price,
        currency: formData.currency,
        imageKeys: uploadedImages.map(img => img.key)
      }

      const response = await api.createListing(dataToSubmit)

      // TODO: Создание объявления согласованно в конечном счете, 
      // мгновенный редирект приводит к 404 до флаша в ES
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
              Изображения *
            </label>

            <div className="create-listing-page__image-preview">
              {images.map((image, index) => (
                <div key={image.previewUrl} className="create-listing-page__image-item">
                  <img
                    src={image.previewUrl}
                    alt={`Preview ${index}`}
                    className="create-listing-page__image-preview-item"
                  />
                  {image.isUploading && (
                    <div className="create-listing-page__image-overlay">
                      <div className="create-listing-page__spinner"></div>
                    </div>
                  )}
                  {image.uploadError && (
                    <div className="create-listing-page__image-error">
                      ⚠️
                    </div>
                  )}
                  {image.key && !image.isUploading && !image.uploadError && (
                    <div className="create-listing-page__image-success">
                      ✓
                    </div>
                  )}
                  <button
                    type="button"
                    onClick={() => removeImage(image.previewUrl)}
                    className="create-listing-page__image-remove"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>

            <label className="create-listing-page__file-input-label">
              <input
                type="file"
                accept="image/*"
                multiple
                onChange={handleFileSelect}
                className="create-listing-page__file-input"
              />
              <span className="create-listing-page__file-input-btn">
                📷 Выбрать фото
              </span>
            </label>

            {images.length > 0 && (
              <p className="create-listing-page__image-hint">
                Загружено: {images.filter(img => img.key && !img.uploadError).length} из {images.length}
              </p>
            )}
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
