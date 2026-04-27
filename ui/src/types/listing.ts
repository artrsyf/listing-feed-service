export interface ListingResponse {
  id: string
  title: string
  description: string
  price: number
  currency: string
  images: string[]
  createdAt: string
}

export interface GetAllListingsResponse {
  listings: ListingResponse[]
  cursor: string | null
}

export interface SearchListingsResponse {
  listings: ListingResponse[]
  nextCursor: string | null
}

export interface CreateListingRequest {
  title: string
  description: string
  price: number
  currency: string
  imageUrls: string[]
}

export interface CreateListingResponse {
  id: string
  createdAt: string
}
