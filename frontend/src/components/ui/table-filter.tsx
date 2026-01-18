"use client";

import { ChevronDown } from "lucide-react";

interface FilterOption<T = string> {
  label: string;
  value: T;
}

interface TableFilterProps<T = string> {
  value: T;
  onChange: (value: T) => void;
  options: FilterOption<T>[];
  placeholder?: string;
}

export function TableFilter<T extends string = string>({
  value,
  onChange,
  options,
  placeholder = "Filter",
}: TableFilterProps<T>) {
  const selectedOption = options.find((opt) => opt.value === value);

  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value as T)}
        className="appearance-none px-4 py-2 pr-10 border border-border rounded-lg focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all text-sm bg-white cursor-pointer"
      >
        <option value="">{placeholder}</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <ChevronDown className="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 pointer-events-none" />
    </div>
  );
}
