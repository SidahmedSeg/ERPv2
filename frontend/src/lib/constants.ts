import {
    LayoutDashboard,
    Calculator,
    ShoppingCart,
    TrendingUp,
    Package,
    Users,
    Truck,
    Wrench,
    Warehouse,
    Settings,
    Building2,
    User,
    SlidersHorizontal,
    LogOut,
    Moon,
    Sun,
    Globe,
    FileText,
    Bell,
    Search,
    BarChart3,
    AlertTriangle,
    Box,
    ArrowUpDown,
    DollarSign,
    ClipboardCheck,
    FolderOpen,
    File,
    Trash2,
    Barcode,
    Layers,
    Receipt,
    Repeat,
    ShoppingBag,
    PackageCheck,
    ClipboardList,
    BookmarkIcon,
    Wallet,
    Tag,
    Mail,
} from 'lucide-react';

export const NAVIGATION_ITEMS = [
    {
        title: 'Dashboard',
        href: '/dashboard',
        icon: LayoutDashboard,
        // No permission required - everyone can see dashboard
    },
    {
        title: 'Inbox',
        href: '/dashboard/emails',
        icon: Mail,
        // No permission required - everyone can access their emails
    },
    {
        title: 'Accounting',
        href: '/dashboard/accounting',
        icon: Calculator,
        // No permission defined yet
    },
    {
        title: 'Buy',
        href: '/dashboard/buy',
        icon: ShoppingCart,
        permission: { resource: 'suppliers', action: 'view' },
        submenu: [
            {
                title: 'Suppliers',
                href: '/dashboard/suppliers',
                icon: Truck,
                permission: { resource: 'suppliers', action: 'view' },
            },
            {
                title: 'Purchase Orders',
                href: '/dashboard/purchase-orders',
                icon: ClipboardList,
                permission: { resource: 'purchase_orders', action: 'view' },
            },
            {
                title: 'Purchase Receipts',
                href: '/dashboard/purchase-receipts',
                icon: PackageCheck,
                permission: { resource: 'purchase_receipts', action: 'view' },
            },
        ],
    },
    {
        title: 'Sell',
        href: '/dashboard/sell',
        icon: TrendingUp,
        permission: { resource: 'customers', action: 'view' },
        submenu: [
            {
                title: 'Estimates',
                href: '/dashboard/estimates',
                icon: FileText,
                permission: { resource: 'estimates', action: 'view' },
            },
            {
                title: 'Sales Orders',
                href: '/dashboard/sales-orders',
                icon: ShoppingBag,
                permission: { resource: 'sales_orders', action: 'view' },
            },
            {
                title: 'Invoices',
                href: '/dashboard/invoices',
                icon: Receipt,
                permission: { resource: 'invoices', action: 'view' },
            },
            {
                title: 'Recurring Invoices',
                href: '/dashboard/recurring-invoices',
                icon: Repeat,
                permission: { resource: 'recurring_invoices', action: 'view' },
            },
        ],
    },
    {
        title: 'Stock',
        href: '/dashboard/stock',
        icon: Package,
        permission: { resource: 'stock', action: 'view_movements' },
        submenu: [
            {
                title: 'Overview',
                href: '/dashboard/stock',
                icon: LayoutDashboard,
                permission: { resource: 'stock', action: 'view_movements' },
            },
            {
                title: 'Products',
                href: '/dashboard/products',
                icon: Box,
                permission: { resource: 'products', action: 'view' },
            },
            {
                title: 'Warehouses',
                href: '/dashboard/stock/warehouses',
                icon: Warehouse,
                permission: { resource: 'stock', action: 'manage_warehouses' },
            },
            {
                title: 'Stock Movements',
                href: '/dashboard/stock/movements',
                icon: ArrowUpDown,
                permission: { resource: 'stock', action: 'view_movements' },
            },
            {
                title: 'Inventory Valuation',
                href: '/dashboard/stock/valuation',
                icon: DollarSign,
                permission: { resource: 'stock', action: 'view_valuation' },
            },
            {
                title: 'Physical Counts',
                href: '/dashboard/stock/counts',
                icon: ClipboardCheck,
                permission: { resource: 'stock', action: 'count' },
            },
            {
                title: 'Low Stock Alerts',
                href: '/dashboard/stock/alerts',
                icon: AlertTriangle,
                permission: { resource: 'stock', action: 'view_movements' },
            },
            {
                title: 'Analytics',
                href: '/dashboard/stock/analytics',
                icon: BarChart3,
                permission: { resource: 'stock', action: 'view_movements' },
            },
            {
                title: 'Serial Numbers',
                href: '/dashboard/serial-numbers',
                icon: Barcode,
                permission: { resource: 'stock', action: 'view_movements' },
            },
            {
                title: 'Batches',
                href: '/dashboard/batches',
                icon: Layers,
                permission: { resource: 'stock', action: 'view_movements' },
            },
        ],
    },
    {
        title: 'Customers',
        href: '/dashboard/customers',
        icon: Users,
        permission: { resource: 'customers', action: 'view' },
    },
    {
        title: 'Analytics',
        href: '/dashboard/analytics',
        icon: BarChart3,
        submenu: [
            {
                title: 'Sales by Period',
                href: '/dashboard/analytics/sales/by-period',
                icon: TrendingUp,
            },
            {
                title: 'Sales by Product',
                href: '/dashboard/analytics/sales/by-product',
                icon: Package,
            },
            {
                title: 'Sales by Customer',
                href: '/dashboard/analytics/sales/by-customer',
                icon: Users,
            },
            {
                title: 'Sales by Team',
                href: '/dashboard/analytics/sales/by-team',
                icon: Users,
            },
            {
                title: 'Customer Segmentation',
                href: '/dashboard/analytics/customers/segmentation',
                icon: Users,
            },
            {
                title: 'Saved Reports',
                href: '/dashboard/analytics/saved-reports',
                icon: BookmarkIcon,
            },
        ],
    },
    {
        title: 'Expenses',
        href: '/dashboard/expenses',
        icon: Wallet,
        permission: { resource: 'expenses', action: 'view' },
        submenu: [
            {
                title: 'All Expenses',
                href: '/dashboard/expenses',
                icon: Wallet,
                permission: { resource: 'expenses', action: 'view' },
            },
            {
                title: 'Expense Categories',
                href: '/dashboard/expense-categories',
                icon: Tag,
                permission: { resource: 'expenses', action: 'view' },
            },
        ],
    },
    {
        title: 'ParaDrive',
        href: '/dashboard/paradrive',
        icon: FolderOpen,
        permission: { resource: 'files', action: 'view' },
        submenu: [
            {
                title: 'My Files',
                href: '/dashboard/paradrive',
                icon: File,
                permission: { resource: 'files', action: 'view' },
            },
            {
                title: 'Trash',
                href: '/dashboard/paradrive/trash',
                icon: Trash2,
                permission: { resource: 'files', action: 'view' },
            },
        ],
    },
    {
        title: 'Services',
        href: '/dashboard/services',
        icon: Wrench,
        // No permission defined yet
    },
    {
        title: 'Warehouse',
        href: '/dashboard/warehouse',
        icon: Warehouse,
        permission: { resource: 'stock', action: 'manage_warehouses' },
    },
];

export const COMPANY_MENU_ITEMS = [
    {
        title: 'Company Settings',
        icon: SlidersHorizontal,
        href: '/dashboard/settings/company',
    },
];

export const USER_MENU_ITEMS = [
    {
        title: 'My Profile',
        icon: User,
        href: '/dashboard/profile',
    },
    {
        title: 'Documentation',
        icon: FileText,
        href: '/docs',
        external: true,
    },
];

export const LANGUAGES = [
    { code: 'en', name: 'English', flag: '🇺🇸' },
    { code: 'fr', name: 'Français', flag: '🇫🇷' },
    { code: 'es', name: 'Español', flag: '🇪🇸' },
    { code: 'ar', name: 'العربية', flag: '🇩🇿' },
];