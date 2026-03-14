import { Menu } from 'lucide-react'
import { useOutletContext } from 'react-router-dom'

interface Props {
  title: string
  children?: React.ReactNode
}

export function PageHeader({ title, children }: Props) {
  const { openSidebar } = useOutletContext<{ openSidebar: () => void }>()

  return (
    <div className="flex items-center justify-between mb-6">
      <div className="flex items-center gap-3">
        <button
          onClick={openSidebar}
          className="text-slate-400 hover:text-slate-200 transition-colors md:hidden"
        >
          <Menu size={22} />
        </button>
        <h2 className="text-xl font-semibold">{title}</h2>
      </div>
      {children && <div className="flex gap-2">{children}</div>}
    </div>
  )
}
