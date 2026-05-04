import type {
  ListingResponse,
  GetAllListingsResponse,
  CreateListingRequest,
  CreateListingResponse,
  GenerateUploadUrlRequest,
  GenerateUploadUrlResponse,
  SearchListingsRequest,
  SearchListingsResponse,
  UpdateListingRequest,
  UpdateListingResponse
} from '../types/listing'

const USE_FIXTURES = import.meta.env.VITE_USE_FIXTURES === 'true'

// Фиксированные цены для каждого товара (детерминированные фикстуры)
const FIXED_PRICES = [
  150000,  // iPhone 15 Pro Max
  250000,  // MacBook Pro 16"
  55000,   // Sony PlayStation 5
  35000,   // Велосипед горный
  45000,   // Диван угловой
  65000,   // Холодильник Samsung
  25000,   // Наушники AirPods Pro
  85000,   // Камера Canon EOS
  15000,   // Стол офисный
  12000,   // Кроссовки Nike Air
  120000,  // Телевизор LG OLED
  65000,   // Планшет iPad Air
  45000,   // Часы Apple Watch
  8000,    // Клавиатура механическая
  5000,    // Мышь игровая
  35000,   // Монитор 27" 4K
  55000,   // Кофемашина DeLonghi
  25000,   // Пылесос робот
  15000,   // Умная колонка
  5000     // Фитнес-браслет
]

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
      id: `fixture-${index}`,
      title: titles[index % titles.length],
      description: descriptions[index % descriptions.length],
      price: FIXED_PRICES[index % FIXED_PRICES.length],
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
      
      // Пытаемся найти объявление по ID из фикстур
      const allFixtures = generateFixtures(100)
      const fixture = allFixtures.listings.find(l => l.id === id)
      
      if (fixture) {
        return {
          ...fixture,
          description: 'Это фикстурное объявление для тестирования. Отличное состояние, полный комплект, гарантия.',
          images: fixture.images.map((img) => ({
            ...img,
            url: img.url.replace('400/300', '600/400') // Более крупное изображение для детальной страницы
          }))
        }
      }
      
      // Если не найдено - возвращаем дефолтное
      return {
        id,
        title: 'Объявление (Fixture)',
        description: 'Это фикстурное объявление для тестирования. Отличное состояние, полный комплект, гарантия.',
        price: 50000,
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
  },

  // Поиск объявлений
  async searchListings(params: SearchListingsRequest): Promise<SearchListingsResponse> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 300))
      // Фильтруем фикстуры по query
      const allFixtures = generateFixtures(100)
      let filtered = allFixtures.listings
      
      if (params.query) {
        const query = params.query.toLowerCase()
        filtered = filtered.filter(l => 
          l.title.toLowerCase().includes(query) || 
          l.description.toLowerCase().includes(query)
        )
      }
      
      if (params.minPrice !== undefined) {
        filtered = filtered.filter(l => l.price >= params.minPrice!)
      }
      
      if (params.maxPrice !== undefined) {
        filtered = filtered.filter(l => l.price <= params.maxPrice!)
      }
      
      const limit = params.limit || 20
      const sliced = filtered.slice(0, limit)
      
      return {
        listings: sliced,
        nextCursor: sliced.length < filtered.length ? String(limit) : null
      }
    }

    const queryParams = new URLSearchParams()
    if (params.query) queryParams.set('q', params.query)
    if (params.cursor) queryParams.set('cursor', params.cursor)
    if (params.limit) queryParams.set('limit', params.limit.toString())
    if (params.minPrice) queryParams.set('minPrice', params.minPrice.toString())
    if (params.maxPrice) queryParams.set('maxPrice', params.maxPrice.toString())

    const response = await fetch(`${API_BASE}/search?${queryParams.toString()}`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    return response.json()
  },

  // Обновление объявления
  async updateListing(id: string, data: UpdateListingRequest): Promise<UpdateListingResponse> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 500))
      return {
        id,
        updatedAt: new Date().toISOString()
      }
    }

    const response = await fetch(`${API_BASE}/listings/${id}`, {
      method: 'PUT',
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

  // Удаление объявления
  async deleteListing(id: string): Promise<void> {
    if (USE_FIXTURES) {
      await new Promise(resolve => setTimeout(resolve, 300))
      return
    }

    const response = await fetch(`${API_BASE}/listings/${id}`, {
      method: 'DELETE'
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
  }
}
