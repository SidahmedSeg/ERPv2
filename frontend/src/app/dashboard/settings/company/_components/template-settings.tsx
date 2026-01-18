"use client";

import { useState, useEffect, useRef } from "react";
import { useAuthStore } from "@/store/auth-store";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { Loader2, Upload, ImageIcon, Trash2 } from "lucide-react";
import Image from "next/image";

interface CompanySettings {
  id: string;
  tenant_id: string;
  logo_url?: string;
}

export function TemplateSettings() {
  const { accessToken } = useAuthStore();
  const [settings, setSettings] = useState<CompanySettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    fetchSettings();
  }, [accessToken]);

  const fetchSettings = async () => {
    try {
      setLoading(true);
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/settings/company`,
        {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        }
      );

      if (!response.ok) {
        const errorText = await response.text();
        console.error("Company settings fetch failed:", response.status, errorText);
        throw new Error(`Failed to fetch company settings: ${response.status}`);
      }

      const result = await response.json();

      if (!result.success) {
        console.error("Company settings response not successful:", result);
        throw new Error(result.error || "Failed to fetch company settings");
      }

      setSettings(result.data);

      if (result.data?.logo_url) {
        setLogoPreview(result.data.logo_url);
      }
    } catch (error) {
      console.error("Error fetching settings:", error);
      toast.error("Failed to load company settings");
    } finally {
      setLoading(false);
    }
  };

  const handleFileSelect = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      toast.error("Please select an image file");
      return;
    }

    // Validate file size (max 5MB)
    if (file.size > 5 * 1024 * 1024) {
      toast.error("File size must be less than 5MB");
      return;
    }

    try {
      setUploading(true);

      // Upload file
      const formData = new FormData();
      formData.append('file', file);

      const uploadResponse = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/files/upload`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
          body: formData,
        }
      );

      if (!uploadResponse.ok) {
        throw new Error("Failed to upload file");
      }

      const uploadResult = await uploadResponse.json();
      const fileUrl = uploadResult.data?.url;

      if (!fileUrl) {
        console.error("Upload result:", uploadResult);
        throw new Error("No file URL returned from upload");
      }

      // Update company settings with new logo URL
      const updateResponse = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/settings/company`,
        {
          method: 'PUT',
          headers: {
            Authorization: `Bearer ${accessToken}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            logo_url: fileUrl,
          }),
        }
      );

      if (!updateResponse.ok) {
        throw new Error("Failed to update company settings");
      }

      setLogoPreview(fileUrl);
      toast.success("Company logo updated successfully");
      fetchSettings();
    } catch (error) {
      console.error("Error uploading logo:", error);
      toast.error("Failed to upload logo");
    } finally {
      setUploading(false);
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleRemoveLogo = async () => {
    try {
      setUploading(true);

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/settings/company`,
        {
          method: 'PUT',
          headers: {
            Authorization: `Bearer ${accessToken}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            logo_url: null,
          }),
        }
      );

      if (!response.ok) {
        throw new Error("Failed to remove logo");
      }

      setLogoPreview(null);
      toast.success("Company logo removed successfully");
      fetchSettings();
    } catch (error) {
      console.error("Error removing logo:", error);
      toast.error("Failed to remove logo");
    } finally {
      setUploading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="pt-6 space-y-6">
          {/* Row 1: Title and Upload Button */}
          <div className="flex items-start justify-between">
            <div>
              <h3 className="text-sm font-medium mb-1">Company Logo</h3>
              <p className="text-xs text-muted-foreground">
                Upload your company logo to appear on invoices, estimates, and other documents
              </p>
            </div>
            <div className="flex gap-2">
              {logoPreview && (
                <Button
                  onClick={handleRemoveLogo}
                  disabled={uploading}
                  variant="outline"
                  size="sm"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Remove
                </Button>
              )}
              <Button
                onClick={handleFileSelect}
                disabled={uploading}
                size="sm"
              >
                {uploading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Uploading...
                  </>
                ) : (
                  <>
                    <Upload className="mr-2 h-4 w-4" />
                    {logoPreview ? 'Change' : 'Upload'}
                  </>
                )}
              </Button>
            </div>
          </div>

          {/* Row 2: Logo Preview (only shown after upload) */}
          {logoPreview && (
            <div className="w-[40%]">
              <div className="border-2 border-border rounded-lg p-6 bg-gray-50">
                <div className="relative w-full h-32 bg-white rounded border border-border flex items-center justify-center p-4">
                  <Image
                    src={logoPreview}
                    alt="Company Logo"
                    fill
                    className="object-contain p-2"
                  />
                </div>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Recommended: PNG or JPG format, max 5MB. Square images work best.
              </p>
            </div>
          )}

          {/* Hidden file input */}
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={handleFileChange}
            className="hidden"
          />
        </CardContent>
      </Card>
    </div>
  );
}
