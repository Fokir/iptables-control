import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../../api/client'
import type { User } from '../../types'

interface Props {
  children: ReactNode
}

export function AuthGuard({ children }: Props) {
  const navigate = useNavigate()
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    api.get<User>('/api/auth/me')
      .then(() => setChecking(false))
      .catch(() => navigate('/login', { replace: true }))
  }, [navigate])

  if (checking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <p className="text-slate-400">Loading...</p>
      </div>
    )
  }

  return <>{children}</>
}
