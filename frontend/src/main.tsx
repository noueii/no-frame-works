import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './index.css'
import { AuthProvider } from './contexts/AuthContext'
import { Login } from './pages/Login'
import { Register } from './pages/Register'
import { Home } from './pages/Home'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/*"
          element={
            <AuthProvider>
              <Routes>
                <Route path="/" element={<Home />} />
              </Routes>
            </AuthProvider>
          }
        />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
