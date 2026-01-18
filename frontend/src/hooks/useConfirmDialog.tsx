'use client';

import { create } from 'zustand';

interface ConfirmDialogStore {
    isOpen: boolean;
    title: string;
    description: string;
    confirmText: string;
    cancelText: string;
    variant: 'default' | 'destructive';
    onConfirm: () => void;
    onCancel: () => void;
    open: (options: {
        title: string;
        description: string;
        confirmText?: string;
        cancelText?: string;
        variant?: 'default' | 'destructive';
        onConfirm: () => void;
        onCancel?: () => void;
    }) => void;
    close: () => void;
}

export const useConfirmDialog = create<ConfirmDialogStore>((set) => ({
    isOpen: false,
    title: '',
    description: '',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    variant: 'default',
    onConfirm: () => {},
    onCancel: () => {},
    open: (options) =>
        set({
            isOpen: true,
            title: options.title,
            description: options.description,
            confirmText: options.confirmText || 'Confirm',
            cancelText: options.cancelText || 'Cancel',
            variant: options.variant || 'default',
            onConfirm: options.onConfirm,
            onCancel: options.onCancel || (() => {}),
        }),
    close: () => set({ isOpen: false }),
}));