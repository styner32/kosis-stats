// API Response Types

export interface Company {
  ID?: number;
  CorpCode?: string;
  CorpName?: string;
  CorpEngName?: string;
  LastModifiedDate?: string;
  CreatedAt?: string;
  UpdatedAt?: string;
  // Fallback fields
  id?: number;
  corp_code?: string;
  corp_name?: string;
  corp_name_eng?: string;
  name?: string;
}

export interface RawReport {
  ID?: number;
  ReceiptNumber?: string;
  CorpCode?: string;
  BlobData?: string | Uint8Array;
  BlobSize?: number;
  JSONData?: string | Record<string, unknown>;
  CreatedAt?: string;
  UpdatedAt?: string;
  // Fallback fields
  id?: number;
  raw_report_id?: number;
  receipt_number?: string;
  receiptNumber?: string;
  recept_no?: string;
  corp_code?: string;
  corpCode?: string;
  blob_data?: string | Uint8Array;
  blob_size?: number;
  json_data?: string | Record<string, unknown>;
  createdAt?: string;
  updatedAt?: string;
  date?: string;
  recept_dt?: string;
  report_nm?: string;
  reportName?: string;
}

export interface Report {
  corp_name?: string;
  corp_code?: string;
  report_name?: string;
  raw_report_id?: number;
  receipt_number?: string;
  analysis?: unknown;
  [key: string]: unknown;
}

export interface RawReportResponse {
  raw_report: string;
}

export interface CompaniesResponse {
  companies: Company[];
}

export interface ReportsResponse {
  reports: RawReport[];
}

export interface RawReportResponse {
  raw_report: string;
}

export interface HealthResponse {
  status: string;
}

export interface FinancialsResponse {
  [key: string]: unknown;
}

// State Type
export interface AppState {
  companies: Company[];
  reports: Report[];
  selectedCompany: string | null;
  selectedReport: string | null;
  selectedYear: string | null;
  loading: boolean;
  error: string | null;
  currentTab: "dashboard" | "reports";
}

// DOM Elements Type
export interface DOMElements {
  // Tabs
  tabButtons: NodeListOf<Element>;
  dashboardView: HTMLElement;
  reportsView: HTMLElement;

  // Dashboard Elements
  companyInput: HTMLInputElement;
  companyList: HTMLUListElement;
  reportSelect: HTMLSelectElement;
  yearSelect: HTMLSelectElement;
  loading: HTMLElement;
  error: HTMLElement;
  results: HTMLElement;
  reportDetails: HTMLElement;

  // Report List Elements
  listCompanyInput: HTMLInputElement;
  listCompanyList: HTMLUListElement;
  dateStart: HTMLInputElement;
  dateEnd: HTMLInputElement;
  sortOrder: HTMLSelectElement;
  reportListContainer: HTMLElement;
}

// Request Options
export interface RequestOptions extends RequestInit {
  headers?: HeadersInit;
}
