import type {
  ListingResponse,
  GetAllListingsResponse,
  CreateListingRequest,
  CreateListingResponse,
  GenerateUploadUrlRequest,
  GenerateUploadUrlResponse
} from '../types/listing'

const USE_FIXTURES = import.meta.env.VITE_USE_FIXTURES === 'true'

const generateFixtures = (count: number, cursor?: string): GetAllListingsResponse => {
  const startIndex = cursor ? parseInt(cursor, 10) : 0
  const titles = [
    'iPhone 15 Pro Max',
    'MacBook Pro 16"',
    'Sony PlayStation 5',
    'Велосипед горный',
    'Диван угловой',
    'Холодильник Samsung',
    'Наушники AirPods Pro',
    'Камера Canon EOS',
    'Стол офисный',
    'Кроссовки Nike Air',
    'Телевизор LG OLED',
    'Планшет iPad Air',
    'Часы Apple Watch',
    'Клавиатура механическая',
    'Мышь игровая',
    'Монитор 27" 4K',
    'Кофемашина DeLonghi',
    'Пылесос робот',
    'Умная колонка',
    'Фитнес-браслет'
  ]
  
  const descriptions = [
    'Отличное состояние, использовался бережно. Полный комплект.',
    'Новый, в упаковке. Гарантия 1 год.',
    'Срочная продажа! Торг уместен.',
    'После одного владельца. Без дефектов.',
    'Подарок на день рождения, но не подошел.',
    'Профессиональное оборудование. Идеально для работы.',
    'Компактный и удобный. Отлично подходит для маленькой квартиры.',
    'Топовая модель с лучшими характеристиками.',
    'Классический дизайн, проверенное качество.',
    'Современный функционал по доступной цене.'
  ]
  
  const currencies = ['RUB', 'USD', 'EUR']
  const imageUrls = [
    'https://picsum.photos/400/300?random=1',
    'https://picsum.photos/400/300?random=2',
    'https://picsum.photos/400/300?random=3',
    'https://picsum.photos/400/300?random=4',
    'https://picsum.photos/400/300?random=5'
  ]

  const listings: ListingResponse[] = Array.from({ length: count }, (_, i) => {
    const index = startIndex + i
    // Генерируем 1-3 изображения для каждого объявления
    const imageCount = (index % 3) + 1
    const listingImages = Array.from({ length: imageCount }, (_, j) => ({
      url: imageUrls[(index + j) % imageUrls.length],
      position: j
    }))
    
    return {
      id: `fixture-${index}-${Date.now()}`,
      title: titles[index % titles.length],
      description: descriptions[index % descriptions.length],
      price: Math.floor(Math.random() * 100000) + 1000,
      currency: currencies[index % currencies.length],
      images: listingImages,
      createdAt: new Date(Date.now() - index * 86400000).toISOString()
    }
  })
  
  const nextCursor = startIndex + count < 100 ? String(startIndex + count) : null
  
  return {
    listings,
    cursor: nextCursor
  }
}

const API_BASE = '/api/feed'

export const api = {
  async getListings(limit: number = 20, cursor?: string): Promise<GetAllListingsResponse> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 300))
      return generateFixtures(limit, cursor)
    }
    
    const params = new URLSearchParams()
    params.set('limit', limit.toString())
    if (cursor) params.set('cursor', cursor)
    
    const response = await fetch(`${API_BASE}/listings?${params.toString()}`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },
  
  async getListing(id: string): Promise<ListingResponse> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 200))
      return {
        id,
        title: 'iPhone 15 Pro Max (Fixture)',
        description: 'Это фикстурное объявление для тестирования. Отличное состояние, полный комплект, гарантия.',
        price: 150000,
        currency: 'RUB',
        images: [
          { url: 'https://picsum.photos/600/400?random=fixture1', position: 0 },
          { url: 'https://picsum.photos/600/400?random=fixture2', position: 1 },
          { url: 'https://picsum.photos/600/400?random=fixture3', position: 2 }
        ],
        createdAt: new Date().toISOString()
      }
    }

    const response = await fetch(`${API_BASE}/listing/${id}`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },
  
  async createListing(data: CreateListingRequest): Promise<CreateListingResponse> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 500))
      return {
        id: `fixture-new-${Date.now()}`,
        createdAt: new Date().toISOString()
      }
    }

    const response = await fetch(`${API_BASE}/listings`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  // Получение URL для загрузки изображения
  async generateUploadUrl(data: GenerateUploadUrlRequest): Promise<GenerateUploadUrlResponse> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 200))
      // Генерируем фиктивный ключ и URL
      const key = `fixture-image-${Date.now()}-${Math.random().toString(36).slice(2)}`
      return {
        key,
        uploadUrl: `https://fixture-minio.example.com/${key}`
      }
    }

    const response = await fetch(`${API_BASE}/images/upload-url`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  // Загрузка файла в объектное хранилище (Minio)
  async uploadFile(url: string, file: File): Promise<void> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 300))
      return
    }

    const response = await fetch(url, {
      method: 'PUT',
      body: file,
      headers: {
        'Content-Type': file.type
      }
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  }
}
