"use client";

import { useState, useEffect, useRef } from "react";
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
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { toast } from "sonner";
import { Loader2, Upload, Building2, Mail, MapPin, Clock, DollarSign, Edit, MapPinIcon } from "lucide-react";
import { FileSelectorDialog } from "@/components/paradrive/file-selector-dialog";
import { Combobox } from "@/components/ui/combobox";
import { DatePicker } from "@/components/ui/date-picker";
import { INDUSTRIES } from "@/lib/industries";
import { COUNTRY_CODES } from "@/lib/country-codes";
import Image from "next/image";
import { format } from "date-fns";

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
  const { accessToken } = useAuthStore();
  const [settings, setSettings] = useState<CompanySettings | null>(null);
  const [loading, setLoading] = useState(true);

  // Dialog states
  const [isCompanyInfoDialogOpen, setIsCompanyInfoDialogOpen] = useState(false);
  const [isContactDialogOpen, setIsContactDialogOpen] = useState(false);
  const [isAddressDialogOpen, setIsAddressDialogOpen] = useState(false);
  const [isBusinessHoursDialogOpen, setIsBusinessHoursDialogOpen] = useState(false);
  const [isFiscalDialogOpen, setIsFiscalDialogOpen] = useState(false);
  const [isParaDriveOpen, setIsParaDriveOpen] = useState(false);

  // Form states for each section
  const [companyInfoForm, setCompanyInfoForm] = useState<any>({});
  const [contactForm, setContactForm] = useState<any>({});
  const [addressForm, setAddressForm] = useState<any>({});
  const [businessHoursForm, setBusinessHoursForm] = useState<any>({});
  const [fiscalForm, setFiscalForm] = useState<any>({});

  const [saving, setSaving] = useState(false);

  // Google Places Autocomplete
  const [googleLoaded, setGoogleLoaded] = useState(false);
  const autocompleteInputRef = useRef<HTMLInputElement | null>(null);
  const autocompleteInstanceRef = useRef<any>(null);
  const preventBlurHandlerRef = useRef<((e: MouseEvent) => void) | null>(null);

  // Phone number states with country code
  const [phoneCountryCode, setPhoneCountryCode] = useState("+213");
  const [faxCountryCode, setFaxCountryCode] = useState("+213");

  useEffect(() => {
    if (accessToken) {
      fetchSettings();
    }
  }, [accessToken]);

  // Load Google Places API
  useEffect(() => {
    if (typeof window !== "undefined" && !(window as any).google) {
      const apiKey = process.env.NEXT_PUBLIC_GOOGLE_PLACES_API_KEY;
      if (!apiKey) {
        console.error("Google Places API key is not configured");
        return;
      }

      const script = document.createElement("script");
      script.src = `https://maps.googleapis.com/maps/api/js?key=${apiKey}&libraries=places`;
      script.async = true;
      script.defer = true;
      script.onload = () => {
        console.log("Google Maps API loaded successfully");
        setGoogleLoaded(true);
      };
      script.onerror = () => {
        console.error("Failed to load Google Maps API");
      };
      document.head.appendChild(script);
    } else if ((window as any).google) {
      console.log("Google Maps API already loaded");
      setGoogleLoaded(true);
    }
  }, []);

  // Initialize Google Places Autocomplete
  useEffect(() => {
    console.log("Autocomplete useEffect triggered:", {
      googleLoaded,
      isAddressDialogOpen,
      hasInputRef: !!autocompleteInputRef.current,
      hasInstance: !!autocompleteInstanceRef.current,
    });

    if (!googleLoaded) {
      console.log("Google not loaded yet");
      return;
    }

    if (!isAddressDialogOpen) {
      // Cleanup when dialog closes
      if (autocompleteInstanceRef.current) {
        console.log("Cleaning up Google Places Autocomplete");
        autocompleteInstanceRef.current = null;
      }
      // Remove the mousedown event listener from pac-container
      if (preventBlurHandlerRef.current) {
        const pacContainers = document.querySelectorAll('.pac-container');
        pacContainers.forEach(container => {
          container.removeEventListener('mousedown', preventBlurHandlerRef.current as any);
        });
        preventBlurHandlerRef.current = null;
      }
      return;
    }

    // Dialog is open and Google is loaded
    if (!autocompleteInstanceRef.current) {
      console.log("Attempting to initialize autocomplete...");

      // Add a small delay to ensure the dialog is fully rendered and ref is attached
      const timeoutId = setTimeout(() => {
        if (!autocompleteInputRef.current) {
          console.warn("Input ref still not available after delay");
          return;
        }

        console.log("Input ref is now available, initializing autocomplete");

        try {
          // Ensure the autocomplete dropdown appears correctly
          const options = {
            types: ["address"],
            fields: ["address_components", "formatted_address"],
          };

          const autocomplete = new (window as any).google.maps.places.Autocomplete(autocompleteInputRef.current, options);

          // Handle pac-container interaction properly
          const handlePacContainerSetup = () => {
            // Wait a bit for pac-container to be created
            setTimeout(() => {
              const pacContainers = document.querySelectorAll('.pac-container');
              if (pacContainers.length > 0) {
                const pacContainer = pacContainers[pacContainers.length - 1] as HTMLElement;

                // Prevent blur on mousedown but allow the click to go through
                pacContainer.addEventListener('mousedown', (e) => {
                  // Don't prevent default - let Google handle it
                  // Just keep the input focused
                  if (autocompleteInputRef.current) {
                    setTimeout(() => {
                      autocompleteInputRef.current?.focus();
                    }, 0);
                  }
                });

                console.log('Attached event handlers to pac-container');
              }
            }, 150);
          };

          // Listen for input events to know when dropdown appears
          autocompleteInputRef.current.addEventListener('input', handlePacContainerSetup);

          // Also setup on initialization
          handlePacContainerSetup();

          autocomplete.addListener("place_changed", () => {
            console.log("Place changed event fired");
            const place = autocomplete.getPlace();
            console.log("Selected place:", place);

            if (!place.address_components) {
              console.warn("No address components found");
              return;
            }

            let street = "";
            let city = "";
            let state = "";
            let postal = "";
            let country = "";

            for (const component of place.address_components) {
              const types = component.types;
              if (types.includes("street_number")) {
                street = component.long_name + " ";
              }
              if (types.includes("route")) {
                street += component.long_name;
              }
              if (types.includes("locality")) {
                city = component.long_name;
              }
              if (types.includes("administrative_area_level_1")) {
                state = component.long_name;
              }
              if (types.includes("postal_code")) {
                postal = component.long_name;
              }
              if (types.includes("country")) {
                country = component.long_name;
              }
            }

            console.log("Parsed address:", { street, city, state, postal, country });

            setAddressForm((prev: any) => ({
              ...prev,
              street_address: street.trim(),
              city,
              state,
              postal_code: postal,
              country,
            }));

            // Clear the search input after selection
            if (autocompleteInputRef.current) {
              autocompleteInputRef.current.value = "";
            }
          });

          autocompleteInstanceRef.current = autocomplete;
          console.log("Google Places Autocomplete initialized successfully");
        } catch (error) {
          console.error("Error initializing Google Places Autocomplete:", error);
        }
      }, 200);

      return () => clearTimeout(timeoutId);
    }
  }, [googleLoaded, isAddressDialogOpen]);

  const fetchSettings = async () => {
    try {
      const response = await fetch("http://localhost:8080/api/settings/company", {
        headers: {
          Authorization: `Bearer ${accessToken}`,
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

  const updateSettings = async (updates: Partial<CompanySettings>) => {
    setSaving(true);
    try {
      const response = await fetch("http://localhost:8080/api/settings/company", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",

          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify(updates),
      });

      const data = await response.json();
      if (data.success) {
        toast.success("Settings updated successfully");
        setSettings(data.data);
        return true;
      } else {
        throw new Error(data.error || "Failed to save settings");
      }
    } catch (error: any) {
      toast.error(error.message || "Failed to save settings");
      return false;
    } finally {
      setSaving(false);
    }
  };

  // Section update handlers
  const handleUpdateCompanyInfo = async () => {
    const success = await updateSettings(companyInfoForm);
    if (success) {
      setIsCompanyInfoDialogOpen(false);
    }
  };

  const handleUpdateContact = async () => {
    // Combine country code with phone number
    const updatedContact = {
      ...contactForm,
      phone_number: contactForm.phone_number ? `${phoneCountryCode} ${contactForm.phone_number}` : undefined,
      fax: contactForm.fax ? `${faxCountryCode} ${contactForm.fax}` : undefined,
    };
    const success = await updateSettings(updatedContact);
    if (success) {
      setIsContactDialogOpen(false);
    }
  };

  const handleUpdateAddress = async () => {
    const success = await updateSettings(addressForm);
    if (success) {
      setIsAddressDialogOpen(false);
    }
  };

  const handleUpdateBusinessHours = async () => {
    const success = await updateSettings(businessHoursForm);
    if (success) {
      setIsBusinessHoursDialogOpen(false);
    }
  };

  const handleUpdateFiscal = async () => {
    const success = await updateSettings(fiscalForm);
    if (success) {
      setIsFiscalDialogOpen(false);
    }
  };

  const handleLogoSelect = async (files: any[]) => {
    if (files.length === 0 || !settings) return;

    const selectedFile = files[0];
    const success = await updateSettings({ logo_url: selectedFile.url });
    if (success) {
      setIsParaDriveOpen(false);
    }
  };

  // Open dialogs with current values
  const openCompanyInfoDialog = () => {
    setCompanyInfoForm({
      company_name: settings?.company_name,
      legal_business_name: settings?.legal_business_name,
      industry: settings?.industry,
      company_size: settings?.company_size,
      founded_date: settings?.founded_date,
      website_url: settings?.website_url,
    });
    setIsCompanyInfoDialogOpen(true);
  };

  const openContactDialog = () => {
    // Parse phone number to extract country code
    const phone = settings?.phone_number || "";
    const fax = settings?.fax || "";

    // Try to extract country code from phone number
    const phoneCode = COUNTRY_CODES.find(cc => phone.startsWith(cc.code))?.code || "+213";
    const phoneNum = phone.replace(phoneCode, "").trim();

    const faxCode = COUNTRY_CODES.find(cc => fax.startsWith(cc.code))?.code || "+213";
    const faxNum = fax.replace(faxCode, "").trim();

    setPhoneCountryCode(phoneCode);
    setFaxCountryCode(faxCode);

    setContactForm({
      primary_email: settings?.primary_email,
      support_email: settings?.support_email,
      phone_number: phoneNum,
      fax: faxNum,
    });
    setIsContactDialogOpen(true);
  };

  const openAddressDialog = () => {
    setAddressForm({
      street_address: settings?.street_address,
      city: settings?.city,
      state: settings?.state,
      postal_code: settings?.postal_code,
      country: settings?.country,
    });
    setIsAddressDialogOpen(true);
  };

  const openBusinessHoursDialog = () => {
    setBusinessHoursForm({
      timezone: settings?.timezone,
      working_days: settings?.working_days,
      working_hours_start: settings?.working_hours_start,
      working_hours_end: settings?.working_hours_end,
    });
    setIsBusinessHoursDialogOpen(true);
  };

  const openFiscalDialog = () => {
    setFiscalForm({
      fiscal_year_start: settings?.fiscal_year_start,
      default_currency: settings?.default_currency,
      date_format: settings?.date_format,
      number_format: settings?.number_format,
      rc_number: settings?.rc_number,
      nif_number: settings?.nif_number,
      nis_number: settings?.nis_number,
      ai_number: settings?.ai_number,
      capital_social: settings?.capital_social,
    });
    setIsFiscalDialogOpen(true);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!settings) {
    return null;
  }

  return (
    <div className="space-y-6">
      {/* Row 1: Company Information + Contact Details */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Company Information - Read Only */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
            <div className="flex items-center gap-2">
              <Building2 className="h-5 w-5 text-primary" />
              <CardTitle>Company Information</CardTitle>
            </div>
            <Button variant="ghost" size="sm" onClick={openCompanyInfoDialog} className="hover:bg-gray-100">
              <Edit className="h-4 w-4" />
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-4">
              {settings.logo_url ? (
                <div className="relative h-16 w-16 border rounded-lg overflow-hidden">
                  <Image src={settings.logo_url} alt="Company Logo" fill className="object-cover" />
                </div>
              ) : (
                <div className="h-16 w-16 border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center bg-gray-50">
                  <Building2 className="h-8 w-8 text-gray-400" />
                </div>
              )}
              <Button variant="outline" size="sm" onClick={() => setIsParaDriveOpen(true)} className="hover:bg-gray-100">
                <Upload className="h-4 w-4 mr-2" />
                Upload Logo
              </Button>
            </div>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div>
                <p className="text-gray-500">Company Name</p>
                <p className="font-medium">{settings.company_name || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">Legal Name</p>
                <p className="font-medium">{settings.legal_business_name || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">Industry</p>
                <p className="font-medium">{settings.industry || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">Company Size</p>
                <p className="font-medium">{settings.company_size || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">Founded Date</p>
                <p className="font-medium">{settings.founded_date || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">Website</p>
                <p className="font-medium truncate">{settings.website_url || "-"}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Contact Details - Read Only */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
            <div className="flex items-center gap-2">
              <Mail className="h-5 w-5 text-primary" />
              <CardTitle>Contact Details</CardTitle>
            </div>
            <Button variant="ghost" size="sm" onClick={openContactDialog} className="hover:bg-gray-100">
              <Edit className="h-4 w-4" />
            </Button>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div>
              <p className="text-gray-500">Primary Email</p>
              <p className="font-medium">{settings.primary_email || "-"}</p>
            </div>
            <div>
              <p className="text-gray-500">Support Email</p>
              <p className="font-medium">{settings.support_email || "-"}</p>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <p className="text-gray-500">Phone Number</p>
                <p className="font-medium">{settings.phone_number || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">Fax</p>
                <p className="font-medium">{settings.fax || "-"}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Row 2: Address + Business Hours */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Address - Read Only */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
            <div className="flex items-center gap-2">
              <MapPin className="h-5 w-5 text-primary" />
              <CardTitle>Address</CardTitle>
            </div>
            <Button variant="ghost" size="sm" onClick={openAddressDialog} className="hover:bg-gray-100">
              <Edit className="h-4 w-4" />
            </Button>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div>
              <p className="text-gray-500">Street Address</p>
              <p className="font-medium">{settings.street_address || "-"}</p>
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <p className="text-gray-500">City</p>
                <p className="font-medium">{settings.city || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">State</p>
                <p className="font-medium">{settings.state || "-"}</p>
              </div>
              <div>
                <p className="text-gray-500">Postal Code</p>
                <p className="font-medium">{settings.postal_code || "-"}</p>
              </div>
            </div>
            <div>
              <p className="text-gray-500">Country</p>
              <p className="font-medium">{settings.country || "-"}</p>
            </div>
          </CardContent>
        </Card>

        {/* Business Hours - Read Only */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
            <div className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-primary" />
              <CardTitle>Business Hours</CardTitle>
            </div>
            <Button variant="ghost" size="sm" onClick={openBusinessHoursDialog} className="hover:bg-gray-100">
              <Edit className="h-4 w-4" />
            </Button>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div>
              <p className="text-gray-500">Timezone</p>
              <p className="font-medium">{settings.timezone}</p>
            </div>
            <div>
              <p className="text-gray-500">Working Days</p>
              <div className="flex gap-2 flex-wrap mt-1">
                {Object.entries(settings.working_days || {}).map(([day, isWorking]) => (
                  <span
                    key={day}
                    className={`px-2 py-1 rounded text-xs ${
                      isWorking ? "bg-primary/10 text-primary" : "bg-gray-100 text-gray-500"
                    }`}
                  >
                    {day.charAt(0).toUpperCase() + day.slice(1, 3)}
                  </span>
                ))}
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <p className="text-gray-500">Start Time</p>
                <p className="font-medium">{settings.working_hours_start}</p>
              </div>
              <div>
                <p className="text-gray-500">End Time</p>
                <p className="font-medium">{settings.working_hours_end}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Row 3: Fiscal Settings (full width) */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div className="flex items-center gap-2">
            <DollarSign className="h-5 w-5 text-primary" />
            <CardTitle>Fiscal Settings</CardTitle>
          </div>
          <Button variant="ghost" size="sm" onClick={openFiscalDialog} className="hover:bg-gray-100">
            <Edit className="h-4 w-4" />
          </Button>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-gray-500">Fiscal Year Start</p>
              <p className="font-medium">{settings.fiscal_year_start || "-"}</p>
            </div>
            <div>
              <p className="text-gray-500">Default Currency</p>
              <p className="font-medium">{settings.default_currency}</p>
            </div>
            <div>
              <p className="text-gray-500">Date Format</p>
              <p className="font-medium">{settings.date_format}</p>
            </div>
            <div>
              <p className="text-gray-500">Number Format</p>
              <p className="font-medium">{settings.number_format}</p>
            </div>
            <div>
              <p className="text-gray-500">RC Number</p>
              <p className="font-medium">{settings.rc_number || "-"}</p>
            </div>
            <div>
              <p className="text-gray-500">NIF Number</p>
              <p className="font-medium">{settings.nif_number || "-"}</p>
            </div>
            <div>
              <p className="text-gray-500">NIS Number</p>
              <p className="font-medium">{settings.nis_number || "-"}</p>
            </div>
            <div>
              <p className="text-gray-500">AI Number</p>
              <p className="font-medium">{settings.ai_number || "-"}</p>
            </div>
            <div>
              <p className="text-gray-500">Capital Social</p>
              <p className="font-medium">
                {settings.capital_social ? `${settings.capital_social.toLocaleString()}` : "-"}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Company Information Dialog */}
      <Dialog open={isCompanyInfoDialogOpen} onOpenChange={setIsCompanyInfoDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Company Information</DialogTitle>
            <DialogDescription>Update your company details</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="company_name">Company Name *</Label>
              <Input
                id="company_name"
                value={companyInfoForm.company_name || ""}
                onChange={(e) => setCompanyInfoForm({ ...companyInfoForm, company_name: e.target.value })}
                placeholder="My Company Ltd."
                className="bg-white"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="legal_business_name">Legal Business Name</Label>
              <Input
                id="legal_business_name"
                value={companyInfoForm.legal_business_name || ""}
                onChange={(e) =>
                  setCompanyInfoForm({ ...companyInfoForm, legal_business_name: e.target.value })
                }
                placeholder="Company Legal Name Inc."
                className="bg-white"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="industry">Industry</Label>
                <Combobox
                  value={companyInfoForm.industry || ""}
                  onChange={(value) => setCompanyInfoForm({ ...companyInfoForm, industry: value })}
                  options={INDUSTRIES}
                  placeholder="Select industry..."
                  className="w-full bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="company_size">Company Size</Label>
                <Select
                  value={companyInfoForm.company_size || ""}
                  onValueChange={(value) => setCompanyInfoForm({ ...companyInfoForm, company_size: value })}
                >
                  <SelectTrigger className="bg-white">
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
                <DatePicker
                  date={companyInfoForm.founded_date ? new Date(companyInfoForm.founded_date) : undefined}
                  onDateChange={(date) =>
                    setCompanyInfoForm({
                      ...companyInfoForm,
                      founded_date: date ? format(date, "yyyy-MM-dd") : ""
                    })
                  }
                  placeholder="Select founded date"
                  className="bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="website_url">Website URL</Label>
                <Input
                  id="website_url"
                  type="url"
                  value={companyInfoForm.website_url || ""}
                  onChange={(e) => setCompanyInfoForm({ ...companyInfoForm, website_url: e.target.value })}
                  placeholder="https://example.com"
                  className="bg-white"
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsCompanyInfoDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateCompanyInfo} disabled={saving}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Update
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Contact Details Dialog */}
      <Dialog open={isContactDialogOpen} onOpenChange={setIsContactDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Contact Details</DialogTitle>
            <DialogDescription>Update contact information</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="primary_email">Primary Email</Label>
              <Input
                id="primary_email"
                type="email"
                value={contactForm.primary_email || ""}
                onChange={(e) => setContactForm({ ...contactForm, primary_email: e.target.value })}
                placeholder="info@company.com"
                className="bg-white"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="support_email">Support Email</Label>
              <Input
                id="support_email"
                type="email"
                value={contactForm.support_email || ""}
                onChange={(e) => setContactForm({ ...contactForm, support_email: e.target.value })}
                placeholder="support@company.com"
                className="bg-white"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="phone_number">Phone Number</Label>
              <div className="flex gap-2">
                <Select value={phoneCountryCode} onValueChange={setPhoneCountryCode}>
                  <SelectTrigger className="w-[140px] bg-white">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {COUNTRY_CODES.map((cc) => (
                      <SelectItem key={cc.code} value={cc.code}>
                        {cc.flag} {cc.code}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Input
                  id="phone_number"
                  type="tel"
                  value={contactForm.phone_number || ""}
                  onChange={(e) => setContactForm({ ...contactForm, phone_number: e.target.value })}
                  placeholder="555 123 456"
                  className="flex-1 bg-white"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="fax">Fax</Label>
              <div className="flex gap-2">
                <Select value={faxCountryCode} onValueChange={setFaxCountryCode}>
                  <SelectTrigger className="w-[140px] bg-white">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {COUNTRY_CODES.map((cc) => (
                      <SelectItem key={cc.code} value={cc.code}>
                        {cc.flag} {cc.code}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Input
                  id="fax"
                  type="tel"
                  value={contactForm.fax || ""}
                  onChange={(e) => setContactForm({ ...contactForm, fax: e.target.value })}
                  placeholder="555 123 456"
                  className="flex-1 bg-white"
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsContactDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateContact} disabled={saving}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Update
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Address Dialog */}
      <Dialog open={isAddressDialogOpen} onOpenChange={setIsAddressDialogOpen}>
        <DialogContent onInteractOutside={(e) => {
          // Prevent dialog from closing when clicking on Google Places dropdown
          const target = e.target as HTMLElement;
          if (target.closest('.pac-container')) {
            e.preventDefault();
          }
        }}>
          <DialogHeader>
            <DialogTitle>Edit Address</DialogTitle>
            <DialogDescription>Update company address</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {googleLoaded && (
              <div className="space-y-2">
                <Label>Search Address (Google Places)</Label>
                <div className="relative">
                  <MapPinIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400 z-10 pointer-events-none" />
                  <input
                    ref={autocompleteInputRef}
                    type="text"
                    placeholder="Search for an address..."
                    className="w-full px-3 py-2 pl-9 border border-border rounded-lg focus:ring-1 focus:ring-primary focus:border-primary outline-none transition-all text-sm bg-white"
                  />
                </div>
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="street_address">Street Address</Label>
              <Textarea
                id="street_address"
                value={addressForm.street_address || ""}
                onChange={(e) => setAddressForm({ ...addressForm, street_address: e.target.value })}
                placeholder="123 Main Street, Suite 100"
                rows={2}
                className="bg-white"
              />
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="city">City</Label>
                <Input
                  id="city"
                  value={addressForm.city || ""}
                  onChange={(e) => setAddressForm({ ...addressForm, city: e.target.value })}
                  placeholder="New York"
                  className="bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="state">State/Province</Label>
                <Input
                  id="state"
                  value={addressForm.state || ""}
                  onChange={(e) => setAddressForm({ ...addressForm, state: e.target.value })}
                  placeholder="NY"
                  className="bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="postal_code">Postal Code</Label>
                <Input
                  id="postal_code"
                  value={addressForm.postal_code || ""}
                  onChange={(e) => setAddressForm({ ...addressForm, postal_code: e.target.value })}
                  placeholder="10001"
                  className="bg-white"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="country">Country</Label>
              <Input
                id="country"
                value={addressForm.country || ""}
                onChange={(e) => setAddressForm({ ...addressForm, country: e.target.value })}
                placeholder="United States"
                className="bg-white"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsAddressDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateAddress} disabled={saving}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Update
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Business Hours Dialog */}
      <Dialog open={isBusinessHoursDialogOpen} onOpenChange={setIsBusinessHoursDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Business Hours</DialogTitle>
            <DialogDescription>Update operating hours and timezone</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="timezone">Timezone</Label>
              <Select
                value={businessHoursForm.timezone || ""}
                onValueChange={(value) => setBusinessHoursForm({ ...businessHoursForm, timezone: value })}
              >
                <SelectTrigger className="bg-white">
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
                      checked={businessHoursForm.working_days?.[day] || false}
                      onCheckedChange={(checked) =>
                        setBusinessHoursForm({
                          ...businessHoursForm,
                          working_days: {
                            ...businessHoursForm.working_days,
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
                  value={businessHoursForm.working_hours_start || ""}
                  onChange={(e) =>
                    setBusinessHoursForm({ ...businessHoursForm, working_hours_start: e.target.value })
                  }
                  className="bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="working_hours_end">End Time</Label>
                <Input
                  id="working_hours_end"
                  type="time"
                  value={businessHoursForm.working_hours_end || ""}
                  onChange={(e) =>
                    setBusinessHoursForm({ ...businessHoursForm, working_hours_end: e.target.value })
                  }
                  className="bg-white"
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsBusinessHoursDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateBusinessHours} disabled={saving}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Update
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Fiscal Settings Dialog */}
      <Dialog open={isFiscalDialogOpen} onOpenChange={setIsFiscalDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Fiscal Settings</DialogTitle>
            <DialogDescription>Update financial and legal information</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="fiscal_year_start">Fiscal Year Start</Label>
                <DatePicker
                  date={fiscalForm.fiscal_year_start ? new Date(fiscalForm.fiscal_year_start) : undefined}
                  onDateChange={(date) =>
                    setFiscalForm({
                      ...fiscalForm,
                      fiscal_year_start: date ? format(date, "MMM d") : ""
                    })
                  }
                  placeholder="Select date"
                  className="bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="default_currency">Default Currency</Label>
                <Select
                  value={fiscalForm.default_currency || ""}
                  onValueChange={(value) => setFiscalForm({ ...fiscalForm, default_currency: value })}
                >
                  <SelectTrigger className="bg-white">
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
                  value={fiscalForm.date_format || ""}
                  onValueChange={(value) => setFiscalForm({ ...fiscalForm, date_format: value })}
                >
                  <SelectTrigger className="bg-white">
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
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="number_format">Number Format</Label>
                <Select
                  value={fiscalForm.number_format || ""}
                  onValueChange={(value) => setFiscalForm({ ...fiscalForm, number_format: value })}
                >
                  <SelectTrigger className="bg-white">
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
                  value={fiscalForm.capital_social || ""}
                  onChange={(e) =>
                    setFiscalForm({ ...fiscalForm, capital_social: parseFloat(e.target.value) || 0 })
                  }
                  placeholder="1000000.00"
                  className="bg-white"
                />
              </div>
            </div>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="rc_number">
                    RC Number <span className="text-xs text-gray-500">(Alphanumeric)</span>
                  </Label>
                  <Input
                    id="rc_number"
                    value={fiscalForm.rc_number || ""}
                    onChange={(e) => setFiscalForm({ ...fiscalForm, rc_number: e.target.value })}
                    placeholder="RC123456"
                    className="bg-white"
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
                    value={fiscalForm.nif_number || ""}
                    onChange={(e) => {
                      const value = e.target.value.replace(/\D/g, "");
                      setFiscalForm({ ...fiscalForm, nif_number: value });
                    }}
                    placeholder="123456789"
                    className="bg-white"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="nis_number">
                    NIS Number <span className="text-xs text-gray-500">(Numeric)</span>
                  </Label>
                  <Input
                    id="nis_number"
                    type="text"
                    pattern="[0-9]*"
                    value={fiscalForm.nis_number || ""}
                    onChange={(e) => {
                      const value = e.target.value.replace(/\D/g, "");
                      setFiscalForm({ ...fiscalForm, nis_number: value });
                    }}
                    placeholder="987654321"
                    className="bg-white"
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
                    value={fiscalForm.ai_number || ""}
                    onChange={(e) => {
                      const value = e.target.value.replace(/\D/g, "");
                      setFiscalForm({ ...fiscalForm, ai_number: value });
                    }}
                    placeholder="111222333"
                    className="bg-white"
                  />
                </div>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsFiscalDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateFiscal} disabled={saving}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Update
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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
