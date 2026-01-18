"use client";

import { useState, useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import { Loader2, Upload, Building2, Mail, MapPin, Clock, DollarSign } from "lucide-react";
import { FileSelectorDialog } from "@/components/paradrive/file-selector-dialog";
import Image from "next/image";

interface CompanySettings {
  id: string;
  tenant_id: string;
  company_name: string;
  legal_business_name?: string;
  industry?: string;
  company_size?: string;
  founded_date?: string;
  website_url?: string;
  logo_url?: string;
  primary_email?: string;
  support_email?: string;
  phone_number?: string;
  fax?: string;
  street_address?: string;
  city?: string;
  state?: string;
  postal_code?: string;
  country?: string;
  timezone: string;
  working_days: Record<string, boolean>;
  working_hours_start: string;
  working_hours_end: string;
  fiscal_year_start?: string;
  default_currency: string;
  date_format: string;
  number_format: string;
  rc_number?: string;
  nif_number?: string;
  nis_number?: string;
  ai_number?: string;
  capital_social?: number;
}

export function GeneralSettings() {
  const { token } = useAuthStore();
  const [settings, setSettings] = useState<CompanySettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [isParaDriveOpen, setIsParaDriveOpen] = useState(false);

  // Fetch settings on mount
  useEffect(() => {
    if (token) {
      fetchSettings();
    }
  }, [token]);

  const fetchSettings = async () => {
    try {
      const response = await fetch("http://localhost:8080/api/settings/company", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const data = await response.json();
      if (data.success) {
        if (data.data) {
          setSettings(data.data);
        } else {
          // No settings exist yet, create default settings
          setSettings({
            id: "",
            tenant_id: "",
            company_name: "",
            timezone: "UTC",
            working_days: {
              monday: true,
              tuesday: true,
              wednesday: true,
              thursday: true,
              friday: true,
              saturday: false,
              sunday: false,
            },
            working_hours_start: "09:00",
            working_hours_end: "17:00",
            default_currency: "USD",
            date_format: "DD/MM/YYYY",
            number_format: "1,000.00",
          });
        }
      }
    } catch (error) {
      console.error("Failed to fetch settings:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!settings) return;

    setSaving(true);
    try {
      const response = await fetch("http://localhost:8080/api/settings/company", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(settings),
      });

      const data = await response.json();
      if (data.success) {
        toast.success("Company settings updated successfully");
        setSettings(data.data);
      } else {
        throw new Error(data.error || "Failed to save settings");
      }
    } catch (error: any) {
      toast.error(error.message || "Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  const handleLogoSelect = (files: any[]) => {
    if (files.length === 0 || !settings) return;

    const selectedFile = files[0];
    setSettings({ ...settings, logo_url: selectedFile.url });
    setIsParaDriveOpen(false);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!settings) {
    return null; // This shouldn't happen as we always set default settings
  }

  return (
    <div className="space-y-6">
      {/* Row 1: Company Information + Contact Details */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Company Information */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Building2 className="h-5 w-5 text-primary" />
              <CardTitle>Company Information</CardTitle>
            </div>
            <CardDescription>Basic information about your company</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="company_name">Company Name *</Label>
              <Input
                id="company_name"
                value={settings.company_name}
                onChange={(e) => setSettings({ ...settings, company_name: e.target.value })}
                placeholder="My Company Ltd."
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="legal_business_name">Legal Business Name</Label>
              <Input
                id="legal_business_name"
                value={settings.legal_business_name || ""}
                onChange={(e) =>
                  setSettings({ ...settings, legal_business_name: e.target.value })
                }
                placeholder="Company Legal Name Inc."
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="industry">Industry</Label>
                <Input
                  id="industry"
                  value={settings.industry || ""}
                  onChange={(e) => setSettings({ ...settings, industry: e.target.value })}
                  placeholder="Technology"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="company_size">Company Size</Label>
                <Select
                  value={settings.company_size || ""}
                  onValueChange={(value) => setSettings({ ...settings, company_size: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select size" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1-10">1-10 employees</SelectItem>
                    <SelectItem value="11-50">11-50 employees</SelectItem>
                    <SelectItem value="51-200">51-200 employees</SelectItem>
                    <SelectItem value="201-500">201-500 employees</SelectItem>
                    <SelectItem value="500+">500+ employees</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="founded_date">Founded Date</Label>
                <Input
                  id="founded_date"
                  type="date"
                  value={settings.founded_date || ""}
                  onChange={(e) => setSettings({ ...settings, founded_date: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="website_url">Website URL</Label>
                <Input
                  id="website_url"
                  type="url"
                  value={settings.website_url || ""}
                  onChange={(e) => setSettings({ ...settings, website_url: e.target.value })}
                  placeholder="https://example.com"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label>Company Logo (500x500px)</Label>
              <div className="flex items-center gap-4">
                <div className="relative h-24 w-24 border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center bg-gray-50 group cursor-pointer overflow-hidden">
                  {settings.logo_url ? (
                    <>
                      <Image
                        src={settings.logo_url}
                        alt="Company Logo"
                        fill
                        className="object-cover"
                      />
                      <div className="absolute inset-0 bg-black bg-opacity-50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                        <Upload className="h-6 w-6 text-white" />
                      </div>
                    </>
                  ) : (
                    <Upload className="h-8 w-8 text-gray-400" />
                  )}
                  <button
                    type="button"
                    className="absolute inset-0 w-full h-full"
                    onClick={() => setIsParaDriveOpen(true)}
                  />
                </div>
                <div className="text-sm text-gray-500">
                  Click to {settings.logo_url ? "change" : "upload"} logo
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Contact Details */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Mail className="h-5 w-5 text-primary" />
              <CardTitle>Contact Details</CardTitle>
            </div>
            <CardDescription>How customers can reach you</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="primary_email">Primary Email</Label>
              <Input
                id="primary_email"
                type="email"
                value={settings.primary_email || ""}
                onChange={(e) => setSettings({ ...settings, primary_email: e.target.value })}
                placeholder="info@company.com"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="support_email">Support Email</Label>
              <Input
                id="support_email"
                type="email"
                value={settings.support_email || ""}
                onChange={(e) => setSettings({ ...settings, support_email: e.target.value })}
                placeholder="support@company.com"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="phone_number">Phone Number</Label>
                <Input
                  id="phone_number"
                  type="tel"
                  value={settings.phone_number || ""}
                  onChange={(e) => setSettings({ ...settings, phone_number: e.target.value })}
                  placeholder="+1 (555) 123-4567"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="fax">Fax</Label>
                <Input
                  id="fax"
                  type="tel"
                  value={settings.fax || ""}
                  onChange={(e) => setSettings({ ...settings, fax: e.target.value })}
                  placeholder="+1 (555) 123-4568"
                />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Row 2: Address + Business Hours */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Address */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <MapPin className="h-5 w-5 text-primary" />
              <CardTitle>Address</CardTitle>
            </div>
            <CardDescription>Company physical location</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="street_address">Street Address</Label>
              <Textarea
                id="street_address"
                value={settings.street_address || ""}
                onChange={(e) => setSettings({ ...settings, street_address: e.target.value })}
                placeholder="123 Main Street, Suite 100"
                rows={2}
              />
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="city">City</Label>
                <Input
                  id="city"
                  value={settings.city || ""}
                  onChange={(e) => setSettings({ ...settings, city: e.target.value })}
                  placeholder="New York"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="state">State/Province</Label>
                <Input
                  id="state"
                  value={settings.state || ""}
                  onChange={(e) => setSettings({ ...settings, state: e.target.value })}
                  placeholder="NY"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="postal_code">Postal Code</Label>
                <Input
                  id="postal_code"
                  value={settings.postal_code || ""}
                  onChange={(e) => setSettings({ ...settings, postal_code: e.target.value })}
                  placeholder="10001"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="country">Country</Label>
              <Input
                id="country"
                value={settings.country || ""}
                onChange={(e) => setSettings({ ...settings, country: e.target.value })}
                placeholder="United States"
              />
            </div>
          </CardContent>
        </Card>

        {/* Business Hours */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-primary" />
              <CardTitle>Business Hours</CardTitle>
            </div>
            <CardDescription>Operating hours and timezone</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="timezone">Timezone</Label>
              <Select
                value={settings.timezone}
                onValueChange={(value) => setSettings({ ...settings, timezone: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="UTC">UTC</SelectItem>
                  <SelectItem value="America/New_York">Eastern Time (ET)</SelectItem>
                  <SelectItem value="America/Chicago">Central Time (CT)</SelectItem>
                  <SelectItem value="America/Denver">Mountain Time (MT)</SelectItem>
                  <SelectItem value="America/Los_Angeles">Pacific Time (PT)</SelectItem>
                  <SelectItem value="Europe/London">London (GMT)</SelectItem>
                  <SelectItem value="Europe/Paris">Paris (CET)</SelectItem>
                  <SelectItem value="Africa/Algiers">Algiers (CET)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Working Days</Label>
              <div className="grid grid-cols-2 gap-3">
                {["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"].map((day) => (
                  <div key={day} className="flex items-center space-x-2">
                    <Checkbox
                      id={day}
                      checked={settings.working_days?.[day] || false}
                      onCheckedChange={(checked) =>
                        setSettings({
                          ...settings,
                          working_days: {
                            ...settings.working_days,
                            [day]: checked as boolean,
                          },
                        })
                      }
                    />
                    <label
                      htmlFor={day}
                      className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 capitalize"
                    >
                      {day}
                    </label>
                  </div>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="working_hours_start">Start Time</Label>
                <Input
                  id="working_hours_start"
                  type="time"
                  value={settings.working_hours_start}
                  onChange={(e) =>
                    setSettings({ ...settings, working_hours_start: e.target.value })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="working_hours_end">End Time</Label>
                <Input
                  id="working_hours_end"
                  type="time"
                  value={settings.working_hours_end}
                  onChange={(e) =>
                    setSettings({ ...settings, working_hours_end: e.target.value })
                  }
                />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Row 3: Fiscal Settings (full width) */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <DollarSign className="h-5 w-5 text-primary" />
            <CardTitle>Fiscal Settings</CardTitle>
          </div>
          <CardDescription>Financial and legal information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="fiscal_year_start">Fiscal Year Start</Label>
              <Input
                id="fiscal_year_start"
                value={settings.fiscal_year_start || ""}
                onChange={(e) => setSettings({ ...settings, fiscal_year_start: e.target.value })}
                placeholder="Jan 1"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="default_currency">Default Currency</Label>
              <Select
                value={settings.default_currency}
                onValueChange={(value) => setSettings({ ...settings, default_currency: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="USD">USD - US Dollar</SelectItem>
                  <SelectItem value="EUR">EUR - Euro</SelectItem>
                  <SelectItem value="GBP">GBP - British Pound</SelectItem>
                  <SelectItem value="DZD">DZD - Algerian Dinar</SelectItem>
                  <SelectItem value="CAD">CAD - Canadian Dollar</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="date_format">Date Format</Label>
              <Select
                value={settings.date_format}
                onValueChange={(value) => setSettings({ ...settings, date_format: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="DD/MM/YYYY">DD/MM/YYYY</SelectItem>
                  <SelectItem value="MM/DD/YYYY">MM/DD/YYYY</SelectItem>
                  <SelectItem value="YYYY-MM-DD">YYYY-MM-DD</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="number_format">Number Format</Label>
              <Select
                value={settings.number_format}
                onValueChange={(value) => setSettings({ ...settings, number_format: value })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1,000.00">1,000.00 (comma, period)</SelectItem>
                  <SelectItem value="1.000,00">1.000,00 (period, comma)</SelectItem>
                  <SelectItem value="1 000.00">1 000.00 (space, period)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="capital_social">Capital Social</Label>
              <Input
                id="capital_social"
                type="number"
                step="0.01"
                value={settings.capital_social || ""}
                onChange={(e) =>
                  setSettings({ ...settings, capital_social: parseFloat(e.target.value) || 0 })
                }
                placeholder="1000000.00"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="space-y-2">
              <Label htmlFor="rc_number">
                RC Number <span className="text-xs text-gray-500">(Alphanumeric)</span>
              </Label>
              <Input
                id="rc_number"
                value={settings.rc_number || ""}
                onChange={(e) => setSettings({ ...settings, rc_number: e.target.value })}
                placeholder="RC123456"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="nif_number">
                NIF Number <span className="text-xs text-gray-500">(Numeric)</span>
              </Label>
              <Input
                id="nif_number"
                type="text"
                pattern="[0-9]*"
                value={settings.nif_number || ""}
                onChange={(e) => {
                  const value = e.target.value.replace(/\D/g, "");
                  setSettings({ ...settings, nif_number: value });
                }}
                placeholder="123456789"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="nis_number">
                NIS Number <span className="text-xs text-gray-500">(Numeric)</span>
              </Label>
              <Input
                id="nis_number"
                type="text"
                pattern="[0-9]*"
                value={settings.nis_number || ""}
                onChange={(e) => {
                  const value = e.target.value.replace(/\D/g, "");
                  setSettings({ ...settings, nis_number: value });
                }}
                placeholder="987654321"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="ai_number">
                AI Number <span className="text-xs text-gray-500">(Numeric)</span>
              </Label>
              <Input
                id="ai_number"
                type="text"
                pattern="[0-9]*"
                value={settings.ai_number || ""}
                onChange={(e) => {
                  const value = e.target.value.replace(/\D/g, "");
                  setSettings({ ...settings, ai_number: value });
                }}
                placeholder="111222333"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Save Button */}
      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={saving} size="lg">
          {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Save Changes
        </Button>
      </div>

      {/* ParaDrive Dialog for Logo Upload */}
      <FileSelectorDialog
        open={isParaDriveOpen}
        onClose={() => setIsParaDriveOpen(false)}
        onSelect={handleLogoSelect}
        fileType="image"
        multiple={false}
        title="Select Company Logo"
        description="Choose an image from ParaDrive or upload a new one (500x500px recommended)"
      />
    </div>
  );
}
