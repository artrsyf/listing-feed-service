import './ListingGrid.css'

interface ListingGridProps {
  children: React.ReactNode
}

function ListingGrid({ children }: ListingGridProps) {
  return (
    <div className="listing-grid">
      {children}
    </div>
  )
}

export default ListingGrid
