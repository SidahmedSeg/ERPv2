'use client';

import { useState, useEffect, useCallback } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { toast } from 'sonner';
import { Loader2, Upload, Image as ImageIcon, FileText, Check, X, Folder, ChevronRight, Home } from 'lucide-react';
import { useRef } from 'react';

interface FileItem {
    id: string;
    file_name: string;
    original_name: string;
    file_type: 'image' | 'document' | 'video' | 'other';
    mime_type: string;
    file_size: number;
    url: string;
    thumbnail_url?: string;
    uploaded_by_name: string;
    created_at: string;
    folder_id?: string | null;
}

interface FolderItem {
    id: string;
    name: string;
    color: string;
    parent_folder_id: string | null;
    path: string;
    subfolder_count: number;
    file_count: number;
    created_at: string;
}

interface FileSelectorDialogProps {
    open: boolean;
    onClose: () => void;
    onSelect: (files: FileItem[]) => void;
    fileType?: 'image' | 'document' | 'video' | 'all';
    multiple?: boolean;
    title?: string;
    description?: string;
}

export function FileSelectorDialog({
    open,
    onClose,
    onSelect,
    fileType = 'all',
    multiple = true,
    title = 'Select Files',
    description = 'Choose files from ParaDrive or upload new ones',
}: FileSelectorDialogProps) {
    const { accessToken } = useAuthStore();
    const [files, setFiles] = useState<FileItem[]>([]);
    const [folders, setFolders] = useState<FolderItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [uploading, setUploading] = useState(false);
    const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set());
    const [fileTypeFilter, setFileTypeFilter] = useState<string>(
        fileType === 'all' ? 'all' : fileType
    );
    const [searchQuery, setSearchQuery] = useState('');
    const fileInputRef = useRef<HTMLInputElement>(null);

    // Folder navigation state
    const [currentFolderId, setCurrentFolderId] = useState<string | null>(null);
    const [folderPath, setFolderPath] = useState<{ id: string | null; name: string }[]>([
        { id: null, name: 'My Files' },
    ]);

    useEffect(() => {
        if (open) {
            fetchFolderContents();
        }
    }, [open, fileTypeFilter, currentFolderId]);

    const fetchFolderContents = async () => {
        setLoading(true);
        try {
            const folderId = currentFolderId || 'root';
            const response = await fetch(
                `${process.env.NEXT_PUBLIC_API_URL}/api/folders/${folderId}/contents`,
                {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                }
            );

            const result = await response.json();
            if (result.success) {
                setFolders(result.data.folders || []);
                // Apply file type filter to files
                let fetchedFiles = result.data.files || [];
                if (fileTypeFilter && fileTypeFilter !== 'all') {
                    fetchedFiles = fetchedFiles.filter((f: FileItem) => f.file_type === fileTypeFilter);
                }
                setFiles(fetchedFiles);
            } else {
                throw new Error(result.error || 'Failed to fetch folder contents');
            }
        } catch (err: any) {
            console.error('Failed to fetch folder contents:', err);
            toast.error(err.message || 'Failed to fetch folder contents');
        } finally {
            setLoading(false);
        }
    };

    // Navigation functions
    const navigateToFolder = (folder: FolderItem) => {
        setCurrentFolderId(folder.id);
        setFolderPath([...folderPath, { id: folder.id, name: folder.name }]);
        setSearchQuery(''); // Clear search when navigating
    };

    const navigateToBreadcrumb = (index: number) => {
        const newPath = folderPath.slice(0, index + 1);
        const targetFolder = newPath[newPath.length - 1];
        setCurrentFolderId(targetFolder.id);
        setFolderPath(newPath);
        setSearchQuery(''); // Clear search when navigating
    };

    const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const files = event.target.files;
        if (!files || files.length === 0) return;

        setUploading(true);
        const uploadPromises = Array.from(files).map(async (file) => {
            const formData = new FormData();
            formData.append('file', file);

            try {
                const response = await fetch(
                    `${process.env.NEXT_PUBLIC_API_URL}/api/files/upload`,
                    {
                        method: 'POST',
                        headers: {
                            Authorization: `Bearer ${token}`,
                        },
                        body: formData,
                    }
                );

                const result = await response.json();
                if (!result.success) {
                    throw new Error(result.error || `Failed to upload ${file.name}`);
                }

                return result.data;
            } catch (err: any) {
                console.error('Upload error:', err);
                toast.error(err.message || `Failed to upload ${file.name}`);
                return null;
            }
        });

        const uploadedFiles = await Promise.all(uploadPromises);
        const successfulUploads = uploadedFiles.filter((f) => f !== null);

        if (successfulUploads.length > 0) {
            toast.success(`Successfully uploaded ${successfulUploads.length} file(s)`);
            fetchFolderContents();
            // Auto-select newly uploaded files
            const newSelection = new Set(selectedFiles);
            successfulUploads.forEach((f) => f && newSelection.add(f.id));
            setSelectedFiles(newSelection);
        }

        setUploading(false);
        // Reset input
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
    };

    const toggleFileSelection = (fileId: string) => {
        const newSelection = new Set(selectedFiles);
        if (newSelection.has(fileId)) {
            newSelection.delete(fileId);
        } else {
            if (!multiple) {
                newSelection.clear();
            }
            newSelection.add(fileId);
        }
        setSelectedFiles(newSelection);
    };

    const handleSelect = () => {
        const selected = files.filter((f) => selectedFiles.has(f.id));
        onSelect(selected);
        onClose();
        setSelectedFiles(new Set());
    };

    const handleCancel = () => {
        onClose();
        setSelectedFiles(new Set());
    };

    const formatFileSize = (bytes: number) => {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    };

    // Filter files by search query
    const filteredFiles = files.filter((file) =>
        file.original_name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <Dialog open={open} onOpenChange={handleCancel}>
            <DialogContent className="max-w-4xl max-h-[80vh] flex flex-col" hideClose>
                <DialogHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
                    <div>
                        <DialogTitle>{title}</DialogTitle>
                        <DialogDescription>{description}</DialogDescription>
                    </div>
                    <input
                        ref={fileInputRef}
                        type="file"
                        accept={fileType === 'image' ? 'image/*' : undefined}
                        multiple
                        onChange={handleFileUpload}
                        className="hidden"
                    />
                    <Button
                        onClick={() => fileInputRef.current?.click()}
                        disabled={uploading}
                        className="bg-primary text-white hover:bg-primary"
                    >
                        {uploading ? (
                            <>
                                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                Uploading...
                            </>
                        ) : (
                            <>
                                <Upload className="h-4 w-4 mr-2" />
                                Upload Files
                            </>
                        )}
                    </Button>
                </DialogHeader>

                <div className="flex-1 overflow-y-auto space-y-4">
                    {/* Breadcrumb Navigation */}
                    {!searchQuery && (
                        <div className="flex items-center gap-2 text-sm pb-2 border-b">
                            {folderPath.map((folder, index) => (
                                <div key={folder.id || 'root'} className="flex items-center gap-2">
                                    {index > 0 && <ChevronRight className="h-4 w-4 text-gray-400" />}
                                    <button
                                        onClick={() => navigateToBreadcrumb(index)}
                                        className={`flex items-center gap-1 hover:text-primary transition-colors ${
                                            index === folderPath.length - 1
                                                ? 'text-primary font-medium'
                                                : 'text-gray-600'
                                        }`}
                                    >
                                        {index === 0 && <Home className="h-4 w-4" />}
                                        {folder.name}
                                    </button>
                                </div>
                            ))}
                        </div>
                    )}

                    {/* Search and Selection Info */}
                    <div className="flex items-center gap-4">
                        <div className="w-64">
                            <input
                                type="text"
                                placeholder="Search files..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                        </div>
                        <div className="text-sm text-muted-foreground whitespace-nowrap">
                            {selectedFiles.size} selected
                        </div>
                    </div>

                    {/* Files Grid */}
                    {loading ? (
                        <div className="flex items-center justify-center py-8">
                            <Loader2 className="h-6 w-6 animate-spin text-primary" />
                        </div>
                    ) : files.length === 0 && folders.length === 0 ? (
                        <div className="text-center py-8 text-muted-foreground">
                            <p>No files found</p>
                            <p className="text-xs mt-1">Upload files to get started</p>
                        </div>
                    ) : filteredFiles.length === 0 && folders.length === 0 ? (
                        <div className="text-center py-8 text-muted-foreground">
                            <p>No files found</p>
                            {searchQuery && <p className="text-xs mt-1">Try a different search term</p>}
                        </div>
                    ) : (
                        <div className="grid grid-cols-6 gap-2">
                            {/* Folders - Not selectable, only clickable to navigate */}
                            {!searchQuery && folders.map((folder) => (
                                <div
                                    key={folder.id}
                                    className="relative group cursor-pointer border-2 border-transparent hover:border-gray-300 rounded-lg overflow-hidden transition-all"
                                    onClick={() => navigateToFolder(folder)}
                                >
                                    <div className="aspect-square bg-gray-50 relative flex items-center justify-center">
                                        <Folder className="h-12 w-12" style={{ color: folder.color }} />
                                    </div>
                                    <div className="p-2 bg-white">
                                        <p className="text-xs font-medium truncate">
                                            {folder.name}
                                        </p>
                                        <p className="text-xs text-muted-foreground">
                                            {folder.subfolder_count + folder.file_count} items
                                        </p>
                                    </div>
                                </div>
                            ))}

                            {/* Files - Selectable */}
                            {filteredFiles.map((file) => {
                                const isSelected = selectedFiles.has(file.id);
                                return (
                                    <div
                                        key={file.id}
                                        className={`relative group cursor-pointer border-2 rounded-lg overflow-hidden transition-all ${
                                            isSelected
                                                ? 'border-primary ring-2 ring-primary/20'
                                                : 'border-transparent hover:border-gray-300'
                                        }`}
                                        onClick={() => toggleFileSelection(file.id)}
                                    >
                                        <div className="aspect-square bg-gray-100 relative">
                                            {file.file_type === 'image' ? (
                                                <img
                                                    src={file.thumbnail_url || file.url}
                                                    alt={file.original_name}
                                                    className="w-full h-full object-cover"
                                                />
                                            ) : (
                                                <div className="w-full h-full flex items-center justify-center">
                                                    <FileText className="h-12 w-12 text-muted-foreground" />
                                                </div>
                                            )}
                                            {isSelected && (
                                                <div className="absolute inset-0 bg-primary/20 flex items-center justify-center">
                                                    <div className="bg-primary text-white rounded-full p-2">
                                                        <Check className="h-5 w-5" />
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                        <div className="p-2 bg-white">
                                            <p className="text-xs font-medium truncate">
                                                {file.original_name}
                                            </p>
                                            <p className="text-xs text-muted-foreground">
                                                {formatFileSize(file.file_size)}
                                            </p>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={handleCancel}>
                        Cancel
                    </Button>
                    <Button
                        onClick={handleSelect}
                        disabled={selectedFiles.size === 0}
                    >
                        Select {selectedFiles.size > 0 && `(${selectedFiles.size})`}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
