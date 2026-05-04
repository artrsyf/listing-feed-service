import { Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import ListingPage from './pages/ListingPage'
import CreateListingPage from './pages/CreateListingPage'
import EditListingPage from './pages/EditListingPage'

function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/listing/:id" element={<ListingPage />} />
      <Route path="/listing/:id/edit" element={<EditListingPage />} />
      <Route path="/create" element={<CreateListingPage />} />
    </Routes>
  )
}

export default App
