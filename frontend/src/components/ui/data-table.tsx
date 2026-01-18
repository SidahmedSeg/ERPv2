"use client";

import { ReactNode } from "react";

export interface Column<TData> {
  id: string;
  header: string;
  accessor?: string | ((row: TData) => any);
  cell?: (row: TData) => ReactNode;
  sortable?: boolean;
  sortKey?: string;
  align?: "left" | "center" | "right";
  width?: string;
  headerClassName?: string;
  className?: string;
}

interface DataTableProps<TData> {
  data?: TData[] | null;
  columns: Column<TData>[];
  loading?: boolean;
  emptyMessage?: string;
  emptySearchMessage?: string;
  onRowClick?: (row: TData) => void;
  rowKey: keyof TData;
  rowClassName?: (row: TData) => string;
  actions?: (row: TData) => ReactNode;
  sortBy?: string | null;
  sortOrder?: "asc" | "desc" | null;
  onSort?: (sortKey: string) => void;
}

export function DataTable<TData extends Record<string, any>>({
  data = [],
  columns,
  loading = false,
  emptyMessage = "No data available",
  emptySearchMessage = "No results found",
  onRowClick,
  rowKey,
  rowClassName,
  actions,
  sortBy,
  sortOrder,
  onSort,
}: DataTableProps<TData>) {
  // Helper to get cell value
  const getCellValue = (row: TData, column: Column<TData>) => {
    if (column.cell) {
      return column.cell(row);
    }
    if (column.accessor) {
      if (typeof column.accessor === "function") {
        return column.accessor(row);
      }
      return row[column.accessor];
    }
    return null;
  };

  // Helper to get alignment classes
  const getAlignmentClass = (align?: "left" | "center" | "right") => {
    switch (align) {
      case "center":
        return "text-center";
      case "right":
        return "text-right";
      default:
        return "text-left";
    }
  };

  // Helper to handle column sorting
  const handleSort = (column: Column<TData>) => {
    if (!column.sortable || !onSort) return;

    const sortKey = column.sortKey || column.id;
    onSort(sortKey);
  };

  // Helper to render sort indicator
  const renderSortIndicator = (column: Column<TData>) => {
    if (!column.sortable) return null;

    const sortKey = column.sortKey || column.id;
    const isActive = sortBy === sortKey;

    if (!isActive) {
      return (
        <span className="ml-2 text-gray-400">
          <svg className="w-3 h-3 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
          </svg>
        </span>
      );
    }

    if (sortOrder === "asc") {
      return (
        <span className="ml-2 text-primary">
          <svg className="w-3 h-3 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
          </svg>
        </span>
      );
    }

    if (sortOrder === "desc") {
      return (
        <span className="ml-2 text-primary">
          <svg className="w-3 h-3 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </span>
      );
    }

    return null;
  };

  return (
    <div className="overflow-x-auto rounded-lg border border-border bg-white">
      <table className="min-w-full divide-y divide-border">
        <thead className="bg-gray-50 border-b border-border">
          <tr>
            {columns.map((column) => (
              <th
                key={column.id}
                className={`px-6 py-3 text-xs font-semibold text-gray-600 uppercase tracking-wider ${getAlignmentClass(
                  column.align
                )} ${column.headerClassName || ""} ${
                  column.sortable ? "cursor-pointer select-none hover:bg-gray-100" : ""
                }`}
                style={column.width ? { width: column.width } : undefined}
                onClick={() => handleSort(column)}
              >
                <div className="flex items-center justify-between">
                  <span>{column.header}</span>
                  {renderSortIndicator(column)}
                </div>
              </th>
            ))}
            {actions && (
              <th className="px-6 py-3 text-right text-xs font-semibold text-gray-600 uppercase tracking-wider w-20">
                Actions
              </th>
            )}
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-border">
          {loading ? (
            <tr>
              <td
                colSpan={columns.length + (actions ? 1 : 0)}
                className="px-6 py-12 text-center"
              >
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  <span className="ml-3 text-gray-500">Loading...</span>
                </div>
              </td>
            </tr>
          ) : !data || data.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length + (actions ? 1 : 0)}
                className="px-6 py-12 text-center text-gray-500"
              >
                {emptySearchMessage || emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((row) => {
              const key = String(row[rowKey]);
              const baseRowClassName =
                "hover:bg-gray-50 transition-colors border-b border-border last:border-b-0";
              const customRowClassName = rowClassName ? rowClassName(row) : "";
              const clickableClassName = onRowClick ? "cursor-pointer" : "";

              return (
                <tr
                  key={key}
                  onClick={() => onRowClick?.(row)}
                  className={`${baseRowClassName} ${customRowClassName} ${clickableClassName}`}
                >
                  {columns.map((column) => (
                    <td
                      key={column.id}
                      className={`px-6 py-4 whitespace-nowrap text-sm ${getAlignmentClass(
                        column.align
                      )} ${column.className || ""}`}
                    >
                      {getCellValue(row, column)}
                    </td>
                  ))}
                  {actions && (
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                      <div
                        onClick={(e) => {
                          e.stopPropagation();
                        }}
                      >
                        {actions(row)}
                      </div>
                    </td>
                  )}
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
}
