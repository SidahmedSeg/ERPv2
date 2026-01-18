'use client';

import { useEffect, useState, useCallback, useRef } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import {
    Package,
    Users,
    Truck,
    TrendingUp,
    Search,
    Loader2,
    FileText,
    FileCheck,
    ShoppingCart,
    Receipt,
    ReceiptText,
    Repeat,
    DollarSign
} from 'lucide-react';
import '@/Loader/Redirect_loader.css';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface SearchResult {
    id: string;
    type: 'product' | 'customer' | 'supplier' | 'stock_movement' | 'file' | 'estimate' | 'sales_order' | 'invoice' | 'recurring_invoice' | 'purchase_order' | 'expense';
    title: string;
    subtitle?: string;
    metadata?: string;
}

export function GlobalSearch() {
    const router = useRouter();
    const pathname = usePathname();
    const { accessToken } = useAuthStore();
    const [query, setQuery] = useState('');
    const [loading, setLoading] = useState(false);
    const [showResults, setShowResults] = useState(false);
    const [navigating, setNavigating] = useState(false);
    const searchRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);
    const [results, setResults] = useState<{
        products: SearchResult[];
        customers: SearchResult[];
        suppliers: SearchResult[];
        stock_movements: SearchResult[];
        files: SearchResult[];
        estimates: SearchResult[];
        sales_orders: SearchResult[];
        invoices: SearchResult[];
        recurring_invoices: SearchResult[];
        purchase_orders: SearchResult[];
        expenses: SearchResult[];
    }>({
        products: [],
        customers: [],
        suppliers: [],
        stock_movements: [],
        files: [],
        estimates: [],
        sales_orders: [],
        invoices: [],
        recurring_invoices: [],
        purchase_orders: [],
        expenses: [],
    });

    // Keyboard shortcut to focus search (Cmd+K or Ctrl+K)
    useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                inputRef.current?.focus();
            }
        };

        document.addEventListener('keydown', down);
        return () => document.removeEventListener('keydown', down);
    }, []);

    // Click outside to close results
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
                setShowResults(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    // Reset navigating state when pathname changes (route navigation completes)
    useEffect(() => {
        setNavigating(false);
    }, [pathname]);

    // Debounced search
    useEffect(() => {
        if (!query || query.length < 2) {
            setResults({
                products: [],
                customers: [],
                suppliers: [],
                stock_movements: [],
                files: [],
                estimates: [],
                sales_orders: [],
                invoices: [],
                recurring_invoices: [],
                purchase_orders: [],
                expenses: [],
            });
            setShowResults(false);
            return;
        }

        const timeoutId = setTimeout(() => {
            performSearch(query);
        }, 300);

        return () => clearTimeout(timeoutId);
    }, [query]);

    const performSearch = async (searchQuery: string) => {
        if (!token) return;

        setLoading(true);
        setShowResults(true);
        try {
            const response = await fetch(
                `${API_URL}/api/search?q=${encodeURIComponent(searchQuery)}&limit=5`,
                {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                }
            );

            if (!response.ok) {
                setResults({
                    products: [],
                    customers: [],
                    suppliers: [],
                    stock_movements: [],
                    files: [],
                    estimates: [],
                    sales_orders: [],
                    invoices: [],
                    recurring_invoices: [],
                    purchase_orders: [],
                    expenses: [],
                });
                return;
            }

            const data = await response.json();

            if (data.success && data.data) {
                // Transform products
                const products = (data.data.products?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'product' as const,
                    title: hit.name,
                    subtitle: `SKU: ${hit.sku}`,
                    metadata: hit.barcode ? `Barcode: ${hit.barcode}` : undefined,
                }));

                // Transform customers
                const customers = (data.data.customers?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'customer' as const,
                    title: hit.name,
                    subtitle: hit.email,
                    metadata: hit.phone,
                }));

                // Transform suppliers
                const suppliers = (data.data.suppliers?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'supplier' as const,
                    title: hit.company_name,
                    subtitle: hit.email,
                    metadata: hit.phone,
                }));

                // Transform stock movements
                const stock_movements = (data.data.stock_movements?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'stock_movement' as const,
                    title: `${hit.movement_number} - ${hit.product_name}`,
                    subtitle: `Type: ${hit.movement_type}`,
                    metadata: `Qty: ${hit.quantity}`,
                }));

                // Transform files
                const files = (data.data.files?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'file' as const,
                    title: hit.original_name || hit.file_name,
                    subtitle: hit.file_type ? `Type: ${hit.file_type}` : undefined,
                    metadata: hit.file_size ? `${(hit.file_size / 1024).toFixed(1)} KB` : undefined,
                }));

                // Transform estimates
                const estimates = (data.data.estimates?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'estimate' as const,
                    title: hit.estimate_number,
                    subtitle: hit.customer_name,
                    metadata: hit.total_amount ? `$${hit.total_amount}` : undefined,
                }));

                // Transform sales orders
                const sales_orders = (data.data.sales_orders?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'sales_order' as const,
                    title: hit.order_number,
                    subtitle: hit.customer_name,
                    metadata: hit.total_amount ? `$${hit.total_amount}` : undefined,
                }));

                // Transform invoices
                const invoices = (data.data.invoices?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'invoice' as const,
                    title: hit.invoice_number,
                    subtitle: hit.customer_name,
                    metadata: hit.total_amount ? `$${hit.total_amount}` : undefined,
                }));

                // Transform recurring invoices
                const recurring_invoices = (data.data.recurring_invoices?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'recurring_invoice' as const,
                    title: hit.template_name,
                    subtitle: hit.customer_name,
                    metadata: hit.amount ? `$${hit.amount}/${hit.frequency}` : undefined,
                }));

                // Transform purchase orders
                const purchase_orders = (data.data.purchase_orders?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'purchase_order' as const,
                    title: hit.po_number,
                    subtitle: hit.supplier_name,
                    metadata: hit.total_amount ? `$${hit.total_amount}` : undefined,
                }));

                // Transform expenses
                const expenses = (data.data.expenses?.hits || []).map((hit: any) => ({
                    id: hit.id,
                    type: 'expense' as const,
                    title: hit.title,
                    subtitle: hit.merchant_name || hit.category_name,
                    metadata: hit.amount ? `$${hit.amount}` : undefined,
                }));

                setResults({
                    products,
                    customers,
                    suppliers,
                    stock_movements,
                    files,
                    estimates,
                    sales_orders,
                    invoices,
                    recurring_invoices,
                    purchase_orders,
                    expenses,
                });
            }
        } catch (error) {
            setResults({
                products: [],
                customers: [],
                suppliers: [],
                stock_movements: [],
                files: [],
                estimates: [],
                sales_orders: [],
                invoices: [],
                recurring_invoices: [],
                purchase_orders: [],
                expenses: [],
            });
        } finally{
            setLoading(false);
        }
    };

    const handleSelect = useCallback((result: SearchResult) => {
        setNavigating(true);
        setShowResults(false);
        setQuery('');

        // Navigate to the appropriate page based on result type
        switch (result.type) {
            case 'product':
                router.push(`/dashboard/products?id=${result.id}`);
                break;
            case 'customer':
                router.push(`/dashboard/customers/${result.id}`);
                break;
            case 'supplier':
                router.push(`/dashboard/suppliers?id=${result.id}`);
                break;
            case 'stock_movement':
                router.push(`/dashboard/stock/movements?id=${result.id}`);
                break;
            case 'file':
                router.push(`/dashboard/paradrive?id=${result.id}`);
                break;
            case 'estimate':
                router.push(`/dashboard/estimates/${result.id}`);
                break;
            case 'sales_order':
                router.push(`/dashboard/sales-orders/${result.id}`);
                break;
            case 'invoice':
                router.push(`/dashboard/invoices/${result.id}`);
                break;
            case 'recurring_invoice':
                router.push(`/dashboard/recurring-invoices/${result.id}`);
                break;
            case 'purchase_order':
                router.push(`/dashboard/purchase-orders/${result.id}`);
                break;
            case 'expense':
                router.push(`/dashboard/expenses/${result.id}`);
                break;
        }
    }, [router]);

    const totalResults = results.products.length +
                        results.customers.length +
                        results.suppliers.length +
                        results.stock_movements.length +
                        results.files.length +
                        results.estimates.length +
                        results.sales_orders.length +
                        results.invoices.length +
                        results.recurring_invoices.length +
                        results.purchase_orders.length +
                        results.expenses.length;

    const renderResultSection = (
        title: string,
        items: SearchResult[],
        Icon: any,
        iconColor: string,
        bgColor: string
    ) => {
        if (!loading && items.length === 0) return null;

        return (
            <div className="border-b border-gray-100 last:border-0">
                <div className="px-3 py-2 text-xs font-semibold text-gray-500 bg-gray-50">
                    {title}
                </div>
                <div className="py-1">
                    {items.map((result) => (
                        <button
                            key={result.id}
                            onClick={() => handleSelect(result)}
                            className="w-full text-left px-3 py-1.5 hover:bg-gray-50 transition-colors flex items-center gap-2"
                        >
                            <div className={`flex-shrink-0 w-6 h-6 ${bgColor} rounded flex items-center justify-center`}>
                                <Icon className={`h-3.5 w-3.5 ${iconColor}`} />
                            </div>
                            <div className="flex-1 min-w-0">
                                <p className="text-sm font-medium text-gray-900 truncate">
                                    {result.title}
                                </p>
                                {result.subtitle && (
                                    <p className="text-xs text-gray-500 truncate">{result.subtitle}</p>
                                )}
                            </div>
                            {result.metadata && (
                                <div className="flex-shrink-0">
                                    <span className="text-xs text-gray-400">{result.metadata}</span>
                                </div>
                            )}
                        </button>
                    ))}
                </div>
            </div>
        );
    };

    return (
        <div ref={searchRef} className="relative w-80">
            {/* Search input */}
            <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                    ref={inputRef}
                    type="text"
                    placeholder="Search everything..."
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    onFocus={() => query.length >= 2 && setShowResults(true)}
                    className="w-full pl-9 pr-16 py-1.5 text-sm bg-gray-50 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white transition-colors"
                />
                {loading && (
                    <Loader2 className="absolute right-10 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-gray-400" />
                )}
                <kbd className="absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border border-gray-300 bg-white px-1.5 font-mono text-[10px] font-medium text-gray-400">
                    <span className="text-xs">⌘</span>K
                </kbd>
            </div>

            {/* Results dropdown */}
            {showResults && query.length >= 2 && (
                <div className="absolute top-full left-0 right-0 mt-2 bg-white border border-gray-200 rounded-lg shadow-lg max-h-[500px] overflow-y-auto z-50">
                    {loading && (
                        <div className="flex items-center justify-center py-6">
                            <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
                            <span className="ml-2 text-sm text-gray-500">Searching...</span>
                        </div>
                    )}

                    {!loading && totalResults === 0 && (
                        <div className="py-6 text-center">
                            <Search className="h-6 w-6 mx-auto mb-2 text-gray-300" />
                            <p className="text-sm text-gray-600">No results found for "{query}"</p>
                            <p className="text-xs text-gray-400 mt-1">Try a different search term</p>
                        </div>
                    )}

                    {renderResultSection('Estimates', results.estimates, FileCheck, 'text-indigo-500', 'bg-indigo-50')}
                    {renderResultSection('Sales Orders', results.sales_orders, ShoppingCart, 'text-blue-500', 'bg-blue-50')}
                    {renderResultSection('Invoices', results.invoices, Receipt, 'text-green-500', 'bg-green-50')}
                    {renderResultSection('Recurring Invoices', results.recurring_invoices, Repeat, 'text-teal-500', 'bg-teal-50')}
                    {renderResultSection('Purchase Orders', results.purchase_orders, ReceiptText, 'text-orange-500', 'bg-orange-50')}
                    {renderResultSection('Expenses', results.expenses, DollarSign, 'text-red-500', 'bg-red-50')}
                    {renderResultSection('Products', results.products, Package, 'text-blue-500', 'bg-blue-50')}
                    {renderResultSection('Customers', results.customers, Users, 'text-green-500', 'bg-green-50')}
                    {renderResultSection('Suppliers', results.suppliers, Truck, 'text-yellow-600', 'bg-yellow-50')}
                    {renderResultSection('Stock Movements', results.stock_movements, TrendingUp, 'text-orange-500', 'bg-orange-50')}
                    {renderResultSection('Files', results.files, FileText, 'text-purple-500', 'bg-purple-50')}
                </div>
            )}

            {/* Navigation Loader Overlay */}
            {navigating && (
                <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[9999] flex items-center justify-center">
                    <div className="flex flex-col items-center gap-4">
                        <div className="loader"></div>
                        <p className="text-sm text-white font-medium">Redirecting...</p>
                    </div>
                </div>
            )}
        </div>
    );
}
