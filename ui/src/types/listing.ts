export interface ListingImageResponse {
  url: string
  position: number
}

export interface ListingResponse {
  id: string
  title: string
  description: string
  price: number
  currency: string
  images: ListingImageResponse[]
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
  imageKeys: string[]
}

export interface CreateListingResponse {
  id: string
  createdAt: string
}

export interface GenerateUploadUrlRequest {
  contentType: string
}

export interface GenerateUploadUrlResponse {
  key: string
  uploadUrl: string
}
