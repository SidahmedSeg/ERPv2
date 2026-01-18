'use client';

import * as React from 'react';
import { Loader2 } from 'lucide-react';
import { Button, buttonVariants } from './button';
import type { VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

interface LoadingButtonProps
  extends React.ComponentProps<'button'>,
    VariantProps<typeof buttonVariants> {
  loading?: boolean;
  loadingText?: string;
  asChild?: boolean;
}

const LoadingButton = React.forwardRef<HTMLButtonElement, LoadingButtonProps>(
  ({
    className,
    variant,
    size,
    children,
    loading = false,
    loadingText,
    disabled,
    asChild = false,
    ...props
  }, ref) => {
    return (
      <Button
        ref={ref}
        className={cn(className)}
        variant={variant}
        size={size}
        disabled={disabled || loading}
        asChild={asChild}
        {...props}
      >
        {loading ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            {loadingText || children}
          </>
        ) : (
          children
        )}
      </Button>
    );
  }
);

LoadingButton.displayName = 'LoadingButton';

export { LoadingButton };
