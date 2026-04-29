import { useEffect, useRef } from 'react'
import './InfiniteScroll.css'

interface InfiniteScrollProps {
  hasMore: boolean
  isLoading: boolean
  onLoadMore: () => void
  children: React.ReactNode
  listingsCount?: number
}

function InfiniteScroll({ hasMore, isLoading, onLoadMore, children, listingsCount = 0 }: InfiniteScrollProps) {
  const observerRef = useRef<IntersectionObserver | null>(null)
  const loadMoreRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (isLoading) return

    observerRef.current = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore) {
          onLoadMore()
        }
      },
      {
        root: null,
        rootMargin: '200px',
        threshold: 0.1
      }
    )

    if (loadMoreRef.current) {
      observerRef.current.observe(loadMoreRef.current)
    }

    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect()
      }
    }
  }, [hasMore, isLoading, onLoadMore])

  return (
    <>
      {children}
      <div ref={loadMoreRef} className="infinite-scroll-trigger">
        {isLoading && hasMore && (
          <div className="infinite-scroll-loader">
            <div className="infinite-scroll-spinner"></div>
          </div>
        )}
        {!hasMore && listingsCount > 0 && (
          <p className="infinite-scroll-end">Больше объявлений нет</p>
        )}
      </div>
    </>
  )
}

export default InfiniteScroll
