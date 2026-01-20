"use client"

import * as React from "react"
import { Calendar as CalendarIcon, ChevronDown } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface DatePickerProps {
  date?: Date
  onDateChange: (date: Date | undefined) => void
  placeholder?: string
  disabled?: boolean
  className?: string
  /**
   * Format function for displaying the selected date
   * @default (date) => date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
   */
  formatDate?: (date: Date) => string
}

/**
 * A refined date picker component with month/year dropdown selectors.
 *
 * Features:
 * - Dropdown month and year selectors for easy navigation
 * - Auto-closes when a date is selected
 * - Smooth transitions and hover states
 * - Fully accessible with keyboard navigation
 * - Supports dark mode
 *
 * @example
 * ```tsx
 * const [date, setDate] = useState<Date>()
 *
 * <DatePicker
 *   date={date}
 *   onDateChange={setDate}
 *   placeholder="Select a date"
 * />
 * ```
 */
export function DatePicker({
  date,
  onDateChange,
  placeholder = "Select date",
  disabled = false,
  className,
  formatDate,
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false)

  // Default date formatter with elegant output
  const defaultFormatDate = (date: Date) => {
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }

  const formatter = formatDate || defaultFormatDate

  const handleSelect = (selectedDate: Date | undefined) => {
    onDateChange(selectedDate)
    // Auto-close the popover when a date is selected
    if (selectedDate) {
      setOpen(false)
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            "w-full justify-between font-normal transition-all duration-200",
            "hover:bg-accent hover:border-primary/20",
            "focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
            !date && "text-muted-foreground",
            disabled && "cursor-not-allowed opacity-50",
            className
          )}
          aria-label={date ? `Selected date: ${formatter(date)}` : "Select date"}
        >
          <div className="flex items-center gap-2">
            <CalendarIcon className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
            <span className="truncate">
              {date ? formatter(date) : placeholder}
            </span>
          </div>
          <ChevronDown
            className={cn(
              "h-4 w-4 flex-shrink-0 text-muted-foreground transition-transform duration-200",
              open && "rotate-180"
            )}
          />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className={cn(
          "w-auto p-0 shadow-lg border border-border/50",
          "animate-in fade-in-0 zoom-in-95",
          "data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95"
        )}
        align="start"
        sideOffset={4}
      >
        <div className="rounded-lg bg-popover overflow-hidden">
          <Calendar
            mode="single"
            selected={date}
            onSelect={handleSelect}
            captionLayout="dropdown"
            fromYear={1900}
            toYear={new Date().getFullYear() + 10}
            initialFocus
            className="rounded-lg"
            classNames={{
              months: "flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0 p-3",
              month: "space-y-4 min-w-[280px]",
              caption: "flex justify-center pt-1 relative items-center gap-1",
              caption_label: "hidden",
              caption_dropdowns: "flex gap-2",
              dropdown: cn(
                "h-9 px-3 py-1 text-sm rounded-md",
                "bg-transparent border border-input hover:bg-accent hover:border-primary/20",
                "transition-colors duration-200",
                "focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
                "appearance-none cursor-pointer"
              ),
              dropdown_month: "flex-1 min-w-[120px]",
              dropdown_year: "flex-1 min-w-[90px]",
              nav: "space-x-1 flex items-center",
              nav_button: cn(
                "h-7 w-7 bg-transparent p-0 hover:bg-accent hover:text-accent-foreground",
                "transition-colors duration-200 rounded-md"
              ),
              nav_button_previous: "absolute left-1",
              nav_button_next: "absolute right-1",
              table: "w-full border-collapse mt-4",
              head_row: "flex",
              head_cell: cn(
                "text-muted-foreground rounded-md w-9 font-medium text-[0.8rem]",
                "uppercase tracking-wider"
              ),
              row: "flex w-full mt-1",
              cell: cn(
                "relative p-0 text-center text-sm",
                "focus-within:relative focus-within:z-20",
                "has-[[aria-selected]]:bg-accent first:[&:has-[aria-selected])]:rounded-l-md",
                "last:[&:has-[aria-selected])]:rounded-r-md"
              ),
              day: cn(
                "h-9 w-9 p-0 font-normal rounded-md",
                "hover:bg-accent hover:text-accent-foreground",
                "transition-all duration-150",
                "aria-selected:opacity-100"
              ),
              day_selected: cn(
                "bg-primary text-primary-foreground",
                "hover:bg-primary hover:text-primary-foreground",
                "focus:bg-primary focus:text-primary-foreground",
                "shadow-sm"
              ),
              day_today: cn(
                "bg-accent/50 text-accent-foreground font-semibold",
                "after:absolute after:bottom-1 after:left-1/2 after:-translate-x-1/2",
                "after:h-1 after:w-1 after:rounded-full after:bg-primary"
              ),
              day_outside: cn(
                "text-muted-foreground/40 opacity-50",
                "aria-selected:bg-accent/50 aria-selected:text-muted-foreground",
                "aria-selected:opacity-30"
              ),
              day_disabled: "text-muted-foreground/30 opacity-50 cursor-not-allowed hover:bg-transparent",
              day_range_middle: cn(
                "aria-selected:bg-accent aria-selected:text-accent-foreground",
                "rounded-none"
              ),
              day_hidden: "invisible",
            }}
          />
        </div>
      </PopoverContent>
    </Popover>
  )
}
