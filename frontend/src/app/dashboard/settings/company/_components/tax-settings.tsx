"use client";

import { useState, useEffect } from "react";
import { useAuthStore } from "@/store/auth-store";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { Loader2, Plus, X, MoreVertical, Edit, Trash2, Copy, Star } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

interface TaxType {
  id: string;
  tenant_id: string;
  name: string;
  rate: number;
  is_active: boolean;
}

interface TaxRate {
  id: string;
  tenant_id: string;
  tax_name: string;
  rate: number;
  tax_type: string;
  applied_to?: string;
  region?: string;
  is_default: boolean;
  is_active: boolean;
  description?: string;
}

export function TaxSettings() {
  const { accessToken } = useAuthStore();
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:18080/api';
  const [taxTypes, setTaxTypes] = useState<TaxType[]>([]);
  const [taxRates, setTaxRates] = useState<TaxRate[]>([]);
  const [loading, setLoading] = useState(true);

  // Tax Type Dialog
  const [isTypeDialogOpen, setIsTypeDialogOpen] = useState(false);
  const [typeForm, setTypeForm] = useState({ name: "" });

  // Tax Rate Dialog
  const [isRateDialogOpen, setIsRateDialogOpen] = useState(false);
  const [editingRate, setEditingRate] = useState<TaxRate | null>(null);
  const [rateForm, setRateForm] = useState({
    tax_name: "",
    rate: 0,
    tax_type: "",
    applied_to: "",
    region: "",
    description: "",
  });

  useEffect(() => {
    fetchTaxTypes();
    fetchTaxRates();
  }, []);

  const fetchTaxTypes = async () => {
    try {
      const response = await fetch("`${apiUrl}/settings/tax-types`", {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      const data = await response.json();
      if (data.success) {
        setTaxTypes(data.data || []);
      }
    } catch (error) {
      console.error("Failed to fetch tax types:", error);
    } finally {
      setLoading(false);
    }
  };

  const fetchTaxRates = async () => {
    try {
      const response = await fetch("`${apiUrl}/settings/tax-rates`", {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      const data = await response.json();
      if (data.success) {
        setTaxRates(data.data || []);
      }
    } catch (error) {
      console.error("Failed to fetch tax rates:", error);
    }
  };

  // Tax Type operations
  const handleCreateType = async () => {
    try {
      const response = await fetch("`${apiUrl}/settings/tax-types`", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify(typeForm),
      });

      const data = await response.json();
      if (data.success) {
        toast.success("Tax type created successfully");
        setIsTypeDialogOpen(false);
        setTypeForm({ name: "" });
        fetchTaxTypes();
      } else {
        throw new Error(data.error);
      }
    } catch (error: any) {
      toast.error(error.message || "Failed to create tax type");
    }
  };

  const handleDeleteType = async (id: string) => {
    if (!confirm("Are you sure you want to delete this tax type?")) return;

    try {
      const response = await fetch(``${apiUrl}/settings/tax-types`/${id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${accessToken}` },
      });

      const data = await response.json();
      if (data.success) {
        toast.success("Tax type deleted successfully");
        fetchTaxTypes();
      }
    } catch (error) {
      toast.error("Failed to delete tax type");
    }
  };

  // Tax Rate operations
  const handleCreateRate = async () => {
    try {
      const response = await fetch("`${apiUrl}/settings/tax-rates`", {
        method: editingRate ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify(rateForm),
      });

      const data = await response.json();
      if (data.success) {
        toast.success(editingRate ? "Tax rate updated successfully" : "Tax rate created successfully");
        setIsRateDialogOpen(false);
        setEditingRate(null);
        setRateForm({
          tax_name: "",
          rate: 0,
          tax_type: "",
          applied_to: "",
          region: "",
          description: "",
        });
        fetchTaxRates();
      }
    } catch (error) {
      toast.error("Failed to save tax rate");
    }
  };

  const handleEditRate = (rate: TaxRate) => {
    setEditingRate(rate);
    setRateForm({
      tax_name: rate.tax_name,
      rate: rate.rate,
      tax_type: rate.tax_type,
      applied_to: rate.applied_to || "",
      region: rate.region || "",
      description: rate.description || "",
    });
    setIsRateDialogOpen(true);
  };

  const handleDeleteRate = async (id: string) => {
    if (!confirm("Are you sure you want to delete this tax rate?")) return;

    try {
      const response = await fetch(``${apiUrl}/settings/tax-rates`/${id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${accessToken}` },
      });

      const data = await response.json();
      if (data.success) {
        toast.success("Tax rate deleted successfully");
        fetchTaxRates();
      }
    } catch (error) {
      toast.error("Failed to delete tax rate");
    }
  };

  const handleSetDefault = async (id: string) => {
    try {
      const response = await fetch(``${apiUrl}/settings/tax-rates`/${id}/default`, {
        method: "PATCH",
        headers: { Authorization: `Bearer ${accessToken}` },
      });

      const data = await response.json();
      if (data.success) {
        toast.success("Default tax rate updated");
        fetchTaxRates();
      }
    } catch (error) {
      toast.error("Failed to set default tax rate");
    }
  };

  const handleDuplicateRate = (rate: TaxRate) => {
    setRateForm({
      tax_name: `${rate.tax_name} (Copy)`,
      rate: rate.rate,
      tax_type: rate.tax_type,
      applied_to: rate.applied_to || "",
      region: rate.region || "",
      description: rate.description || "",
    });
    setIsRateDialogOpen(true);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
      {/* Left Column: Tax Types (25%) */}
      <div className="lg:col-span-1">
        <Card className="h-full">
          <CardHeader>
            <CardTitle className="text-lg">Tax Types</CardTitle>
            <CardDescription>User-defined tax categories</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="space-y-2">
              {taxTypes.map((type) => (
                <div
                  key={type.id}
                  className="flex items-center justify-between p-2 pl-4 border rounded-lg hover:bg-gray-50"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{type.name}</p>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDeleteType(type.id)}
                    className="h-8 w-8 p-0"
                  >
                    <X className="h-4 w-4 text-gray-500 hover:text-red-600" />
                  </Button>
                </div>
              ))}

              {taxTypes.length === 0 && (
                <p className="text-sm text-gray-500 text-center py-4">No tax types yet</p>
              )}
            </div>

            <Button
              onClick={() => setIsTypeDialogOpen(true)}
              variant="outline"
              className="w-full bg-white hover:bg-gray-100"
              size="sm"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Tax Type
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Right Column: Tax Rates (75%) */}
      <div className="lg:col-span-3">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Tax Rates</CardTitle>
                <CardDescription>Detailed tax rate configuration</CardDescription>
              </div>
              <Button onClick={() => setIsRateDialogOpen(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Add Tax Rate
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Tax Name</TableHead>
                  <TableHead>Rate (%)</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Applied To</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {taxRates.map((rate) => (
                  <TableRow key={rate.id}>
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        {rate.tax_name}
                        {rate.is_default && (
                          <Badge variant="default" className="text-xs">
                            <Star className="h-3 w-3 mr-1" />
                            Default
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>{rate.rate}%</TableCell>
                    <TableCell>{rate.tax_type}</TableCell>
                    <TableCell>{rate.applied_to || "-"}</TableCell>
                    <TableCell>
                      <Badge variant={rate.is_active ? "default" : "secondary"}>
                        {rate.is_active ? "Active" : "Inactive"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleEditRate(rate)}>
                            <Edit className="h-4 w-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          {!rate.is_default && (
                            <DropdownMenuItem onClick={() => handleSetDefault(rate.id)}>
                              <Star className="h-4 w-4 mr-2" />
                              Set as Default
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuItem onClick={() => handleDuplicateRate(rate)}>
                            <Copy className="h-4 w-4 mr-2" />
                            Duplicate
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleDeleteRate(rate.id)}
                            className="text-red-600"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}

                {taxRates.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center text-gray-500 py-8">
                      No tax rates configured yet. Click "Add Tax Rate" to get started.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>

      {/* Tax Type Dialog */}
      <Dialog open={isTypeDialogOpen} onOpenChange={setIsTypeDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Tax Type</DialogTitle>
            <DialogDescription>Create a new tax type category</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="type_name">Name</Label>
              <Input
                id="type_name"
                value={typeForm.name}
                onChange={(e) => setTypeForm({ ...typeForm, name: e.target.value })}
                placeholder="e.g., TVA, VAT, Sales Tax"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsTypeDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateType}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Tax Rate Dialog */}
      <Dialog open={isRateDialogOpen} onOpenChange={setIsRateDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editingRate ? "Edit Tax Rate" : "Add Tax Rate"}</DialogTitle>
            <DialogDescription>Configure detailed tax rate settings</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="rate_tax_name">Tax Name</Label>
                <Input
                  id="rate_tax_name"
                  value={rateForm.tax_name}
                  onChange={(e) => setRateForm({ ...rateForm, tax_name: e.target.value })}
                  placeholder="e.g., TVA 19%"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="rate_rate">Rate (%)</Label>
                <Input
                  id="rate_rate"
                  type="number"
                  step="0.01"
                  value={rateForm.rate}
                  onChange={(e) =>
                    setRateForm({ ...rateForm, rate: parseFloat(e.target.value) || 0 })
                  }
                  placeholder="19.00"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="rate_tax_type">Tax Type</Label>
                <Select
                  value={rateForm.tax_type}
                  onValueChange={(value) => setRateForm({ ...rateForm, tax_type: value })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    {taxTypes.map((type) => (
                      <SelectItem key={type.id} value={type.name}>
                        {type.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="rate_applied_to">Applied To</Label>
                <Input
                  id="rate_applied_to"
                  value={rateForm.applied_to}
                  onChange={(e) => setRateForm({ ...rateForm, applied_to: e.target.value })}
                  placeholder="e.g., Products, Services"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="rate_region">Region (Optional)</Label>
              <Input
                id="rate_region"
                value={rateForm.region}
                onChange={(e) => setRateForm({ ...rateForm, region: e.target.value })}
                placeholder="e.g., Algeria, France"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="rate_description">Description (Optional)</Label>
              <Input
                id="rate_description"
                value={rateForm.description}
                onChange={(e) => setRateForm({ ...rateForm, description: e.target.value })}
                placeholder="Additional notes"
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsRateDialogOpen(false);
                setEditingRate(null);
              }}
            >
              Cancel
            </Button>
            <Button onClick={handleCreateRate}>
              {editingRate ? "Update" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
