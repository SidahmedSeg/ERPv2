"use client"

import * as React from "react"
import { Check, ChevronsUpDown } from "lucide-react"
import { cn } from "@/lib/utils"

interface ComboboxOption {
  label: string
  value: string
}

interface ComboboxProps {
  value?: string
  onChange: (value: string) => void
  options: readonly string[] | string[] | ComboboxOption[]
  placeholder?: string
  searchPlaceholder?: string
  emptyText?: string
  className?: string
}

export function Combobox({
  value,
  onChange,
  options,
  placeholder = "Select...",
  searchPlaceholder = "Search...",
  emptyText = "No results found.",
  className
}: ComboboxProps) {
  const [open, setOpen] = React.useState(false)
  const [search, setSearch] = React.useState("")
  const ref = React.useRef<HTMLDivElement>(null)

  // Normalize options to always work with { label, value } format
  const normalizedOptions: ComboboxOption[] = React.useMemo(() => {
    return options.map(option => {
      if (typeof option === 'string') {
        return { label: option, value: option }
      }
      return option as ComboboxOption
    })
  }, [options])

  const filteredOptions = normalizedOptions.filter(option =>
      option.label.toLowerCase().includes(search.toLowerCase())
  )

  // Find the label for the current value
  const selectedOption = normalizedOptions.find(opt => opt.value === value)

  // Close dropdown when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setOpen(false)
        setSearch("")
      }
    }

    if (open) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [open])

  return (
      <div className="relative" ref={ref}>
        <button
            type="button"
            onClick={() => setOpen(!open)}
            className={cn(
                "flex h-9 w-full items-center justify-between rounded-lg border border-border bg-white px-3 py-2 text-sm ring-offset-white placeholder:text-text-muted focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary disabled:cursor-not-allowed disabled:opacity-50",
                className
            )}
        >
        <span className={value ? "text-text-primary" : "text-text-muted"}>
          {selectedOption ? selectedOption.label : placeholder}
        </span>
          <ChevronsUpDown className="h-4 w-4 opacity-50" />
        </button>

        {open && (
            <div className="absolute z-[9999] mt-1 w-full bg-white border border-border rounded-lg shadow-lg max-h-60 overflow-hidden">
                <div className="p-2 border-b border-border">
                <input
                    type="text"
                    placeholder={searchPlaceholder}
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-primary"
                    onClick={(e) => e.stopPropagation()}
                />
              </div>
              <div className="max-h-48 overflow-y-auto p-1">
                {filteredOptions.length === 0 ? (
                    <div className="px-3 py-2 text-sm text-text-secondary">{emptyText}</div>
                ) : (
                    filteredOptions.map((option) => (
                        <button
                            key={option.value}
                            type="button"
                            onClick={() => {
                              onChange(option.value)
                              setOpen(false)
                              setSearch("")
                            }}
                            className={cn(
                                "w-full px-3 py-2 text-left text-sm hover:bg-gray-100 rounded-md flex items-center justify-between",
                                value === option.value && "bg-gray-100"
                            )}
                        >
                          {option.label}
                          {value === option.value && <Check className="h-4 w-4" />}
                        </button>
                    ))
                )}
              </div>
            </div>
        )}
      </div>
  )
}