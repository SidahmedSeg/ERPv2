"use client";

import { useState } from "react";
import { GeneralSettings } from "./_components/general-settings";
import { TaxSettings } from "./_components/tax-settings";
import { TemplateSettings } from "./_components/template-settings";

export default function CompanySettingsPage() {
  const [activeTab, setActiveTab] = useState("general");

  const tabs = [
    { id: "general", label: "General" },
    { id: "taxes", label: "Taxes" },
    { id: "template", label: "Template" },
  ];

  return (
    <div className="h-full w-full flex flex-col bg-white">
      {/* Tabs */}
      <div className="border-b border-border px-6">
        <div className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-1 py-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? "border-primary text-primary"
                  : "border-transparent text-text-secondary hover:text-text-primary"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto p-6">
        {activeTab === "general" && <GeneralSettings />}
        {activeTab === "taxes" && <TaxSettings />}
        {activeTab === "template" && <TemplateSettings />}
      </div>
    </div>
  );
}
