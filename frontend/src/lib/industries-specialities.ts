/**
 * Comprehensive Industry-Speciality Mapping
 * 15 core industries, each with 5-10 specialities
 */

export interface IndustrySpeciality {
  industry: string;
  specialities: string[];
}

export const INDUSTRIES_SPECIALITIES: IndustrySpeciality[] = [
  {
    industry: "Technology & Software",
    specialities: [
      "Software Development",
      "IT Consulting",
      "Cybersecurity",
      "Cloud Computing",
      "Data Analytics & AI",
      "Computer Hardware",
      "Telecommunications",
      "Internet Services",
      "Semiconductors",
      "Wireless Technology",
    ],
  },
  {
    industry: "Manufacturing & Industrial",
    specialities: [
      "Automotive Manufacturing",
      "Electronics Manufacturing",
      "Chemical Manufacturing",
      "Machinery & Equipment",
      "Building Materials",
      "Textiles & Apparel",
      "Furniture Manufacturing",
      "Packaging",
      "Metal Fabrication",
      "Plastics & Polymers",
    ],
  },
  {
    industry: "Healthcare & Medical",
    specialities: [
      "Hospitals & Clinics",
      "Pharmaceuticals",
      "Medical Devices",
      "Biotechnology",
      "Healthcare Services",
      "Dental Services",
      "Mental Health Services",
      "Medical Research",
      "Veterinary Services",
    ],
  },
  {
    industry: "Financial Services",
    specialities: [
      "Banking",
      "Insurance",
      "Investment Banking",
      "Accounting",
      "Financial Planning",
      "Asset Management",
      "Payment Processing",
      "Fintech",
      "Venture Capital",
    ],
  },
  {
    industry: "Professional Services",
    specialities: [
      "Management Consulting",
      "Legal Services",
      "Human Resources",
      "Marketing & Advertising",
      "Public Relations",
      "Market Research",
      "Translation Services",
      "Event Management",
      "Business Process Outsourcing",
    ],
  },
  {
    industry: "Retail & E-commerce",
    specialities: [
      "Online Retail",
      "Department Stores",
      "Specialty Retail",
      "Luxury Goods",
      "Consumer Electronics",
      "Fashion & Apparel",
      "Home & Garden",
      "Wholesale Trade",
      "Import/Export",
    ],
  },
  {
    industry: "Food & Hospitality",
    specialities: [
      "Restaurants & Cafes",
      "Hotels & Resorts",
      "Food Production",
      "Food & Beverage Distribution",
      "Catering Services",
      "Fast Food Chains",
      "Breweries & Wineries",
      "Tourism Services",
      "Event Venues",
    ],
  },
  {
    industry: "Construction & Real Estate",
    specialities: [
      "Residential Construction",
      "Commercial Construction",
      "Civil Engineering",
      "Real Estate Development",
      "Property Management",
      "Architecture & Design",
      "Interior Design",
      "Facilities Management",
      "Infrastructure",
    ],
  },
  {
    industry: "Transportation & Logistics",
    specialities: [
      "Freight & Shipping",
      "Airlines & Aviation",
      "Warehousing & Storage",
      "Supply Chain Management",
      "Courier & Delivery",
      "Public Transportation",
      "Automotive Services",
      "Maritime Services",
      "Rail Transport",
    ],
  },
  {
    industry: "Media & Entertainment",
    specialities: [
      "Broadcasting & Television",
      "Film & Video Production",
      "Publishing",
      "Music & Audio Production",
      "Gaming & Esports",
      "Digital Media",
      "Sports Management",
      "Advertising Agencies",
      "Content Creation",
    ],
  },
  {
    industry: "Energy & Utilities",
    specialities: [
      "Oil & Gas",
      "Renewable Energy",
      "Electric Utilities",
      "Water & Waste Management",
      "Mining",
      "Nuclear Energy",
      "Energy Trading",
      "Environmental Services",
    ],
  },
  {
    industry: "Education & Training",
    specialities: [
      "Higher Education",
      "K-12 Education",
      "Online Learning Platforms",
      "Corporate Training",
      "Vocational Training",
      "Educational Technology",
      "Language Schools",
      "Tutoring Services",
      "Educational Publishing",
    ],
  },
  {
    industry: "Agriculture & Natural Resources",
    specialities: [
      "Crop Production",
      "Livestock & Dairy",
      "Forestry",
      "Fisheries & Aquaculture",
      "Agricultural Technology",
      "Farm Equipment",
      "Organic Farming",
      "Horticulture",
    ],
  },
  {
    industry: "Government & Public Sector",
    specialities: [
      "Federal Government",
      "State/Provincial Government",
      "Local Government",
      "Public Safety",
      "Military & Defense",
      "Public Administration",
      "Regulatory Agencies",
      "International Organizations",
    ],
  },
  {
    industry: "Other Services",
    specialities: [
      "Security Services",
      "Cleaning Services",
      "Repair & Maintenance",
      "Personal Services",
      "Cosmetics & Beauty",
      "Funeral Services",
      "Non-Profit Organizations",
      "Think Tanks & Research",
      "Consulting (Other)",
    ],
  },
];

// Extract just the industry names for the first dropdown
export const INDUSTRIES = INDUSTRIES_SPECIALITIES.map((item) => item.industry);

// Helper function to get specialities for a selected industry
export const getSpecialities = (industry: string): string[] => {
  const found = INDUSTRIES_SPECIALITIES.find((item) => item.industry === industry);
  return found ? found.specialities : [];
};

// Type definitions
export type Industry = typeof INDUSTRIES[number];
export type Speciality = string;
