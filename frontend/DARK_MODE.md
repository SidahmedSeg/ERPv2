# Dark Mode Implementation - Best Practices

## Overview

MyERP v2 implements dark mode using **next-themes**, the industry-standard solution for Next.js applications. This implementation follows real-world best practices to ensure a seamless user experience.

## Architecture

### 1. Theme Provider Setup

**Location:** `src/app/layout.tsx`

```tsx
<ThemeProvider
  attribute="class"           // Uses class-based theming (.dark)
  defaultTheme="system"       // Respects OS preference by default
  enableSystem                // Detects system preference changes
  storageKey="myerp-theme"    // Persistent storage in localStorage
>
  {children}
</ThemeProvider>
```

### 2. Theme Toggle Component

**Location:** `src/components/ui/theme-toggle.tsx`

Features:
- ✅ **Three-way toggle**: Light, Dark, System
- ✅ **Animated icons**: Smooth sun/moon rotation
- ✅ **Dropdown menu**: Clean UX following Vercel/Shadcn patterns
- ✅ **Hydration-safe**: Prevents flash of unstyled content (FOUC)
- ✅ **Keyboard accessible**: Full a11y support

```tsx
// Prevents hydration mismatch
const [mounted, setMounted] = React.useState(false);

React.useEffect(() => {
  setMounted(true);
}, []);

if (!mounted) {
  return <LoadingState />;
}
```

### 3. CSS Variable System

**Location:** `src/app/globals.css`

All colors are defined as HSL values in CSS variables:

```css
:root {
  /* Light mode colors */
  --primary: 243.6 75.4% 58.4%;      /* Indigo #4F46E5 */
  --background: 0 0% 100%;           /* White */
  --foreground: 222.2 47.4% 11.2%;  /* Dark slate */
  /* ... */
}

.dark {
  /* Dark mode colors */
  --primary: 238.7 83.5% 66.7%;      /* Lighter indigo */
  --background: 222.2 47.4% 11.2%;   /* Dark slate */
  --foreground: 210 40% 98.4%;       /* Off-white */
  /* ... */
}
```

### 4. Tailwind Integration

**Location:** `tailwind.config.ts`

Colors reference CSS variables for dynamic theming:

```ts
colors: {
  primary: {
    DEFAULT: 'hsl(var(--primary))',
    foreground: 'hsl(var(--primary-foreground))',
    hover: 'hsl(var(--primary-hover))',
    subtle: 'hsl(var(--primary-subtle))',
  },
  background: 'hsl(var(--background))',
  foreground: 'hsl(var(--foreground))',
  // ...
}
```

## Best Practices Implemented

### 1. ✅ Prevent Flash of Unstyled Content (FOUC)

```tsx
// Root layout
<html lang="en" suppressHydrationWarning>
```

The `suppressHydrationWarning` attribute prevents React hydration warnings when the theme is applied before hydration completes.

### 2. ✅ System Preference Detection

```tsx
defaultTheme="system"
enableSystem
```

Automatically detects and respects the user's OS dark mode preference. Updates in real-time when system preference changes.

### 3. ✅ Persistent Storage

```tsx
storageKey="myerp-theme"
```

User's theme preference is saved to `localStorage` and persists across sessions.

### 4. ✅ SSR-Safe Rendering

The ThemeToggle component uses a `mounted` state to prevent server-side rendering mismatches:

```tsx
const [mounted, setMounted] = React.useState(false);

React.useEffect(() => {
  setMounted(true);
}, []);

if (!mounted) {
  return <LoadingState />;  // Prevents hydration mismatch
}
```

### 5. ✅ Semantic Color Tokens

Instead of hardcoding colors, use semantic tokens:

```tsx
// ✅ Good - Adapts to theme
<div className="bg-background text-foreground">

// ❌ Bad - Hardcoded
<div className="bg-white text-gray-900">
```

### 6. ✅ Icon Transitions

Smooth icon rotation animations:

```tsx
<Sun className="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
<Moon className="absolute h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
```

### 7. ✅ Accessibility

- Semantic HTML (`<button>`, proper ARIA labels)
- Keyboard navigation support
- Screen reader friendly (`sr-only` text)
- Focus states for all interactive elements

## Usage in Components

### Reading Current Theme

```tsx
'use client';

import { useTheme } from 'next-themes';

export function MyComponent() {
  const { theme, setTheme, systemTheme } = useTheme();

  // Current theme: 'light' | 'dark' | 'system'
  console.log(theme);

  // Resolved theme (when theme is 'system')
  const currentTheme = theme === 'system' ? systemTheme : theme;

  // Change theme
  setTheme('dark');
}
```

### Conditional Rendering Based on Theme

```tsx
'use client';

import { useTheme } from 'next-themes';

export function ThemeAwareComponent() {
  const { resolvedTheme } = useTheme();

  return (
    <div>
      {resolvedTheme === 'dark' ? (
        <DarkModeLogo />
      ) : (
        <LightModeLogo />
      )}
    </div>
  );
}
```

### Dark Mode Variants in Tailwind

```tsx
// Different styles for dark mode
<div className="bg-white dark:bg-gray-900 text-gray-900 dark:text-white">
  Content
</div>

// Using semantic tokens (preferred)
<div className="bg-background text-foreground">
  Content
</div>
```

## Color Palette

### Light Mode
- **Primary (Indigo):** `#4F46E5` (HSL: `243.6 75.4% 58.4%`)
- **Secondary (Cyan):** `#06B6D4` (HSL: `187 94.5% 42.7%`)
- **Background:** `#FFFFFF` (HSL: `0 0% 100%`)
- **Foreground:** `#1E293B` (HSL: `222.2 47.4% 11.2%`)

### Dark Mode
- **Primary (Indigo):** `#818CF8` (HSL: `238.7 83.5% 66.7%`)
- **Secondary (Cyan):** `#22D3EE` (HSL: `187 79.5% 53.3%`)
- **Background:** `#1E293B` (HSL: `222.2 47.4% 11.2%`)
- **Foreground:** `#F8FAFC` (HSL: `210 40% 98.4%`)

### Semantic Colors
| Color   | Light Mode | Dark Mode  |
|---------|------------|------------|
| Success | `#16A34A`  | `#22C55E`  |
| Warning | `#F59E0B`  | `#FCD34D`  |
| Error   | `#EF4444`  | `#F87171`  |
| Info    | `#06B6D4`  | `#22D3EE`  |

## Testing Checklist

When implementing dark mode features:

- [ ] Test with `defaultTheme="light"`
- [ ] Test with `defaultTheme="dark"`
- [ ] Test with `defaultTheme="system"` (change OS preference)
- [ ] Verify no FOUC on page load
- [ ] Check localStorage persistence (refresh page)
- [ ] Test theme toggle in all states
- [ ] Verify colors are legible in both modes
- [ ] Check component-specific dark mode variants
- [ ] Test with browser DevTools forced colors
- [ ] Verify accessibility (keyboard navigation, screen readers)

## Common Patterns

### Card Component with Dark Mode

```tsx
<Card className="bg-card text-card-foreground border-border">
  <CardHeader>
    <CardTitle>Title</CardTitle>
  </CardHeader>
  <CardContent>
    Content
  </CardContent>
</Card>
```

### Button with Hover States

```tsx
<Button className="bg-primary text-primary-foreground hover:bg-primary-hover">
  Click me
</Button>
```

### Input with Dark Mode

```tsx
<Input
  className="bg-background border-input focus:ring-ring"
  placeholder="Enter text..."
/>
```

## Performance Considerations

1. **CSS Variables:** Using `hsl(var(--variable))` adds minimal overhead and enables instant theme switching
2. **LocalStorage:** Theme preference is cached, preventing flicker on subsequent visits
3. **System Detection:** Native browser API (`window.matchMedia`) is efficient and event-driven
4. **No JavaScript Flash:** Theme is applied before React hydration via `next-themes`

## Migration from Other Solutions

### From styled-components ThemeProvider

```tsx
// Old
const theme = useContext(ThemeContext);

// New
const { theme, resolvedTheme } = useTheme();
```

### From Manual localStorage

```tsx
// Old
const [theme, setTheme] = useState(localStorage.getItem('theme'));

// New - handled automatically
const { theme, setTheme } = useTheme();
```

## Resources

- **next-themes Documentation:** https://github.com/pacocoursey/next-themes
- **Tailwind Dark Mode:** https://tailwindcss.com/docs/dark-mode
- **shadcn/ui Theming:** https://ui.shadcn.com/docs/theming
- **Next.js App Router:** https://nextjs.org/docs/app

---

**Last Updated:** January 17, 2026
**Version:** 2.0.0
**Status:** Production Ready ✅
