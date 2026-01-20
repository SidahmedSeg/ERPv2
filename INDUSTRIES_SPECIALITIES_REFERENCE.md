# Industries & Specialities Reference

## Overview
**15 Core Industries** with **5-10 Specialities** each

---

## 1. Technology & Software
- Software Development
- IT Consulting
- Cybersecurity
- Cloud Computing
- Data Analytics & AI
- Computer Hardware
- Telecommunications
- Internet Services
- Semiconductors
- Wireless Technology

---

## 2. Manufacturing & Industrial
- Automotive Manufacturing
- Electronics Manufacturing
- Chemical Manufacturing
- Machinery & Equipment
- Building Materials
- Textiles & Apparel
- Furniture Manufacturing
- Packaging
- Metal Fabrication
- Plastics & Polymers

---

## 3. Healthcare & Medical
- Hospitals & Clinics
- Pharmaceuticals
- Medical Devices
- Biotechnology
- Healthcare Services
- Dental Services
- Mental Health Services
- Medical Research
- Veterinary Services

---

## 4. Financial Services
- Banking
- Insurance
- Investment Banking
- Accounting
- Financial Planning
- Asset Management
- Payment Processing
- Fintech
- Venture Capital

---

## 5. Professional Services
- Management Consulting
- Legal Services
- Human Resources
- Marketing & Advertising
- Public Relations
- Market Research
- Translation Services
- Event Management
- Business Process Outsourcing

---

## 6. Retail & E-commerce
- Online Retail
- Department Stores
- Specialty Retail
- Luxury Goods
- Consumer Electronics
- Fashion & Apparel
- Home & Garden
- Wholesale Trade
- Import/Export

---

## 7. Food & Hospitality
- Restaurants & Cafes
- Hotels & Resorts
- Food Production
- Food & Beverage Distribution
- Catering Services
- Fast Food Chains
- Breweries & Wineries
- Tourism Services
- Event Venues

---

## 8. Construction & Real Estate
- Residential Construction
- Commercial Construction
- Civil Engineering
- Real Estate Development
- Property Management
- Architecture & Design
- Interior Design
- Facilities Management
- Infrastructure

---

## 9. Transportation & Logistics
- Freight & Shipping
- Airlines & Aviation
- Warehousing & Storage
- Supply Chain Management
- Courier & Delivery
- Public Transportation
- Automotive Services
- Maritime Services
- Rail Transport

---

## 10. Media & Entertainment
- Broadcasting & Television
- Film & Video Production
- Publishing
- Music & Audio Production
- Gaming & Esports
- Digital Media
- Sports Management
- Advertising Agencies
- Content Creation

---

## 11. Energy & Utilities
- Oil & Gas
- Renewable Energy
- Electric Utilities
- Water & Waste Management
- Mining
- Nuclear Energy
- Energy Trading
- Environmental Services

---

## 12. Education & Training
- Higher Education
- K-12 Education
- Online Learning Platforms
- Corporate Training
- Vocational Training
- Educational Technology
- Language Schools
- Tutoring Services
- Educational Publishing

---

## 13. Agriculture & Natural Resources
- Crop Production
- Livestock & Dairy
- Forestry
- Fisheries & Aquaculture
- Agricultural Technology
- Farm Equipment
- Organic Farming
- Horticulture

---

## 14. Government & Public Sector
- Federal Government
- State/Provincial Government
- Local Government
- Public Safety
- Military & Defense
- Public Administration
- Regulatory Agencies
- International Organizations

---

## 15. Other Services
- Security Services
- Cleaning Services
- Repair & Maintenance
- Personal Services
- Cosmetics & Beauty
- Funeral Services
- Non-Profit Organizations
- Think Tanks & Research
- Consulting (Other)

---

## Total Count
- **Industries:** 15
- **Specialities:** 137 total (average 9.1 per industry)
- **Range:** 8-10 specialities per industry

---

## UI Behavior

### Cascading Selection Flow:
1. User opens "Company Information" dialog
2. **First Combobox (Industry):** Shows 15 industries
3. User selects industry (e.g., "Technology & Software")
4. **Second Combobox (Speciality):** Becomes enabled, shows 10 specialities
5. User selects speciality (e.g., "Software Development")
6. If user changes industry → speciality resets to empty

### Database Storage:
```json
{
  "industry": "Technology & Software",
  "speciality": "Software Development"
}
```

### Display:
```
Industry: Technology & Software
Speciality: Software Development
```

---

**Created:** 2026-01-19
**File Location:** `/Users/intelifoxdz/Zyndra/myerp-v2/frontend/src/lib/industries-specialities.ts`
