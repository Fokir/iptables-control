import { clsx } from 'clsx'
import { useState, useCallback } from 'react'

interface Props {
  label?: string
  error?: string
  value: string
  onChange: (value: string) => void
  placeholder?: string
  required?: boolean
  className?: string
}

function isIpv6(value: string): boolean {
  return value.includes(':')
}

function maskIpv4(raw: string): string {
  // Allow only digits and dots
  let cleaned = raw.replace(/[^0-9.]/g, '')

  // Prevent leading dots
  if (cleaned.startsWith('.')) cleaned = cleaned.slice(1)

  // Prevent consecutive dots
  cleaned = cleaned.replace(/\.{2,}/g, '.')

  // Split into octets and limit to 4
  const parts = cleaned.split('.')
  const limited = parts.slice(0, 4)

  // Clamp each octet to 0-255 and limit to 3 digits
  const clamped = limited.map((part) => {
    if (part === '') return part
    // Remove leading zeros (but keep single "0")
    let num = part.replace(/^0+(?=\d)/, '')
    if (num === '') num = '0'
    // Limit to 3 digits
    num = num.slice(0, 3)
    // Clamp to 255
    const n = parseInt(num, 10)
    if (n > 255) return '255'
    return num
  })

  return clamped.join('.')
}

function maskIpv6(raw: string): string {
  // Allow only hex digits and colons
  let cleaned = raw.replace(/[^0-9a-fA-F:]/g, '')

  // Prevent leading colons (unless ::)
  if (cleaned.startsWith(':') && !cleaned.startsWith('::')) {
    cleaned = cleaned.slice(1)
  }

  // Prevent more than 2 consecutive colons
  cleaned = cleaned.replace(/:{3,}/g, '::')

  // Prevent multiple :: groups
  const doubleColonCount = (cleaned.match(/::/g) || []).length
  if (doubleColonCount > 1) {
    const firstIdx = cleaned.indexOf('::')
    cleaned = cleaned.slice(0, firstIdx + 2) + cleaned.slice(firstIdx + 2).replace(/::/g, ':')
  }

  // Limit each group to 4 hex digits
  const hasDoubleColon = cleaned.includes('::')
  const segments = cleaned.split(':')
  const maxGroups = 8
  const limitedSegments = segments.slice(0, maxGroups + (hasDoubleColon ? 1 : 0))
  const clamped = limitedSegments.map(seg => seg.slice(0, 4))

  // Max length 39 chars (xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx)
  let result = clamped.join(':')
  if (result.length > 39) result = result.slice(0, 39)

  return result
}

function validateIp(value: string): string | undefined {
  if (!value) return undefined

  if (isIpv6(value)) {
    // Basic IPv6 validation
    const expanded = value.replace('::', ':DOUBLE:')
    const parts = expanded.split(':').filter(Boolean)
    const regularParts = parts.filter(p => p !== 'DOUBLE')
    const hasDouble = parts.includes('DOUBLE')

    if (!hasDouble && regularParts.length !== 8) return 'Incomplete IPv6 address'
    if (hasDouble && regularParts.length >= 8) return 'Too many IPv6 groups'

    for (const part of regularParts) {
      if (!/^[0-9a-fA-F]{1,4}$/.test(part)) return 'Invalid IPv6 group'
    }
    return undefined
  }

  // IPv4 validation
  const parts = value.split('.')
  if (parts.length !== 4) return 'Incomplete IPv4 address'
  for (const part of parts) {
    if (part === '') return 'Incomplete IPv4 address'
    const n = parseInt(part, 10)
    if (isNaN(n) || n < 0 || n > 255) return 'Octet must be 0-255'
  }
  return undefined
}

export function IpInput({ label, error: externalError, value, onChange, placeholder, required, className }: Props) {
  const [touched, setTouched] = useState(false)

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value
    if (!raw) {
      onChange('')
      return
    }

    const masked = isIpv6(raw) ? maskIpv6(raw) : maskIpv4(raw)
    onChange(masked)
  }, [onChange])

  const validationError = touched && value ? validateIp(value) : undefined
  const displayError = externalError || validationError

  return (
    <div className="flex flex-col gap-1">
      {label && <label className="text-sm text-slate-400">{label}</label>}
      <input
        value={value}
        onChange={handleChange}
        onBlur={() => setTouched(true)}
        className={clsx(
          'bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-slate-100 placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors font-mono',
          displayError && 'border-red-500',
          className,
        )}
        placeholder={placeholder ?? '0.0.0.0'}
        required={required}
        autoComplete="off"
        spellCheck={false}
      />
      {displayError && <span className="text-sm text-red-400">{displayError}</span>}
    </div>
  )
}
