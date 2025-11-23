import {
    Company,
    RawReport,
    CompaniesResponse,
    ReportsResponse,
    HealthResponse,
    FinancialsResponse,
    AppState,
    DOMElements,
    RequestOptions,
} from './types';

type AnalysisRecord = {
    RawReportID?: number;
    Analysis?: unknown;
    CreatedAt?: string;
    created_at?: string;
};

// Configuration
const API_BASE_URL = 'http://localhost:8080';
const API_VERSION = '/api/v1';

// State management
const state: AppState = {
    companies: [],
    reports: [],
    selectedCompany: null,
    selectedReport: null,
    selectedYear: null,
    loading: false,
    error: null,
};

// DOM elements
const elements: DOMElements = {
    companySelect: document.getElementById('company-select') as HTMLSelectElement,
    reportSelect: document.getElementById('report-select') as HTMLSelectElement,
    yearSelect: document.getElementById('year-select') as HTMLSelectElement,
    loading: document.getElementById('loading')!,
    error: document.getElementById('error')!,
    results: document.getElementById('results')!,
    reportDetails: document.getElementById('report-details')!,
};

// Type guard to check if element exists
function assertElement(element: HTMLElement | null, name: string): asserts element is HTMLElement {
    if (!element) {
        throw new Error(`Element ${name} not found`);
    }
}

// Helper function to get field value with fallbacks
function getFieldValue<T>(
    obj: Record<string, unknown>,
    ...keys: string[]
): T | undefined {
    for (const key of keys) {
        const value = obj[key];
        if (value !== undefined && value !== null) {
            return value as T;
        }
    }
    return undefined;
}

// API functions
async function apiRequest<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const url = `${API_BASE_URL}${API_VERSION}${endpoint}`;

    const defaultOptions: RequestOptions = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const config: RequestOptions = { ...defaultOptions, ...options };

    try {
        const response = await fetch(url, config);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('API request failed:', error);
        throw error;
    }
}

async function fetchCompanies(): Promise<void> {
    try {
        showLoading();
        const data = await apiRequest<CompaniesResponse | Company[]>('/companies');
        state.companies = Array.isArray(data) ? data : (data.companies || []);
        populateCompanySelect();
        hideLoading();
    } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        showError(`Failed to load companies: ${message}`);
        state.companies = [];
        populateCompanySelect();
    }
}

async function fetchReports(corpCode: string): Promise<void> {
    if (!corpCode) {
        state.reports = [];
        populateReportSelect();
        return;
    }

    try {
        showLoading();
        const data = await apiRequest<ReportsResponse | AnalysisRecord[]>(`/reports/${corpCode}`);
        const reports = Array.isArray(data)
            ? data
            : ((data as { reports?: AnalysisRecord[] }).reports || []);
        state.reports = reports;
        populateReportSelect();
        populateYearSelect();
        hideLoading();
    } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        showError(`Failed to load reports: ${message}`);
        state.reports = [];
        populateReportSelect();
    }
}

async function fetchFinancials(corpCode: string): Promise<FinancialsResponse | null> {
    if (!corpCode) {
        return null;
    }

    try {
        const data = await apiRequest<FinancialsResponse>(`/financials?corp_code=${corpCode}`);
        return data;
    } catch (error) {
        console.error('Failed to load financials:', error);
        return null;
    }
}

// UI update functions
function populateCompanySelect(): void {
    elements.companySelect.innerHTML = '<option value="">Select a company</option>';

    if (state.companies.length === 0) {
        elements.companySelect.innerHTML = '<option value="">No companies available</option>';
        elements.companySelect.disabled = true;
        return;
    }

    elements.companySelect.disabled = false;

    state.companies.forEach((company) => {
        const option = document.createElement('option');
        const corpCode = getFieldValue<string>(company as Record<string, unknown>, 'CorpCode', 'corp_code', 'corpCode', 'id') || '';
        const corpName = getFieldValue<string>(company as Record<string, unknown>, 'CorpName', 'corp_name', 'corpName', 'name') || `Company ${corpCode}`;
        
        option.value = corpCode;
        option.textContent = corpName;
        elements.companySelect.appendChild(option);
    });
}

function populateReportSelect(): void {
    elements.reportSelect.innerHTML = '<option value="">Select a report</option>';

    if (!state.selectedCompany) {
        elements.reportSelect.disabled = true;
        return;
    }

    const filteredReports = state.selectedYear
        ? state.reports.filter((r) => {
              const date = getFieldValue<string>(
                  r as Record<string, unknown>,
                  'CreatedAt',
                  'created_at',
                  'createdAt'
              );
              if (!date) return false;
              try {
                  const dateObj = new Date(date);
                  return (
                      !isNaN(dateObj.getTime()) &&
                      dateObj.getFullYear().toString() === state.selectedYear
                  );
              } catch {
                  return false;
              }
          })
        : state.reports;

    if (filteredReports.length === 0) {
        elements.reportSelect.innerHTML = '<option value="">No reports available</option>';
        elements.reportSelect.disabled = true;
        return;
    }

    elements.reportSelect.disabled = false;

    filteredReports.forEach((report) => {
        const option = document.createElement('option');
        const rawReportId =
            getFieldValue<string | number>(
                report as Record<string, unknown>,
                'RawReportID',
                'raw_report_id',
                'rawReportId'
            ) || '';
        const date = getFieldValue<string>(
            report as Record<string, unknown>,
            'CreatedAt',
            'created_at',
            'createdAt'
        );
        let dateStr = '';
        if (date) {
            try {
                dateStr = new Date(date).toLocaleDateString();
            } catch {
                dateStr = date;
            }
        }
        option.value = String(rawReportId);
        option.textContent = `Report ${rawReportId}${dateStr ? ' - ' + dateStr : ''}`;
        elements.reportSelect.appendChild(option);
    });
}

function populateYearSelect(): void {
    const years = new Set<string>();

    state.reports.forEach((report) => {
        const date = getFieldValue<string>(
            report as Record<string, unknown>,
            'CreatedAt',
            'created_at',
            'createdAt'
        );
        if (date) {
            try {
                const dateObj = new Date(date);
                if (!isNaN(dateObj.getTime())) {
                    years.add(dateObj.getFullYear().toString());
                }
            } catch {
                // Skip invalid dates
            }
        }
    });

    const sortedYears = Array.from(years).sort((a, b) => b.localeCompare(a));

    elements.yearSelect.innerHTML = '<option value="">All Years</option>';
    sortedYears.forEach((year) => {
        const option = document.createElement('option');
        option.value = year;
        option.textContent = year;
        elements.yearSelect.appendChild(option);
    });
}

function displayReportDetails(data: AnalysisRecord | null): void {
    if (!data) {
        elements.reportDetails.innerHTML = '<div class="empty-state">No report details available</div>';
        return;
    }

    let html = '<div class="report-card">';

    html += '<h3>Report Details</h3>';
    html += '<div class="meta">';

    const rawReportId = getFieldValue<string | number>(
        data as Record<string, unknown>,
        'RawReportID',
        'raw_report_id',
        'rawReportId'
    );
    const createdAt = getFieldValue<string>(
        data as Record<string, unknown>,
        'CreatedAt',
        'created_at',
        'createdAt'
    );

    if (rawReportId !== undefined) {
        html += `<span><strong>Raw Report ID:</strong> ${rawReportId}</span>`;
    }

    if (createdAt) {
        try {
            const date = new Date(createdAt);
            html += `<span><strong>Date:</strong> ${date.toLocaleDateString()}</span>`;
        } catch {
            html += `<span><strong>Date:</strong> ${createdAt}</span>`;
        }
    }
    html += '</div>';

    const analysisData = getFieldValue<unknown>(data as Record<string, unknown>, 'Analysis', 'analysis');
    let parsedAnalysis: Record<string, unknown> | unknown = analysisData;
    if (typeof analysisData === 'string') {
        try {
            parsedAnalysis = JSON.parse(analysisData);
        } catch {
            parsedAnalysis = analysisData;
        }
    }

    if (parsedAnalysis) {
        html += '<div class="json-viewer">';
        html += '<h4>Analysis</h4>';
        html += `<pre>${escapeHtml(JSON.stringify(parsedAnalysis, null, 2))}</pre>`;
        html += '</div>';
    } else {
        html += '<div class="empty-state">No analysis available</div>';
    }

    html += '</div>';
    elements.reportDetails.innerHTML = html;
    elements.results.classList.remove('hidden');
}

function clearReportDetails(): void {
    elements.reportDetails.innerHTML = '';
    elements.results.classList.add('hidden');
}

function showLoading(): void {
    state.loading = true;
    elements.loading.classList.remove('hidden');
    elements.error.classList.add('hidden');
    elements.results.classList.add('hidden');
}

function hideLoading(): void {
    state.loading = false;
    elements.loading.classList.add('hidden');
}

function showError(message: string): void {
    state.error = message;
    elements.error.textContent = message;
    elements.error.classList.remove('hidden');
    elements.loading.classList.add('hidden');
}

function escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Event handlers
elements.companySelect.addEventListener('change', async (e: Event) => {
    const target = e.target as HTMLSelectElement;
    state.selectedCompany = target.value || null;
    state.selectedReport = null;
    elements.reportSelect.value = '';
    clearReportDetails();

    if (state.selectedCompany) {
        await fetchReports(state.selectedCompany);
    } else {
        state.reports = [];
        populateReportSelect();
    }
});

elements.reportSelect.addEventListener('change', async (e: Event) => {
    const target = e.target as HTMLSelectElement;
    state.selectedReport = target.value || null;

    if (state.selectedReport) {
        const report = state.reports.find((r) => {
            const id = getFieldValue<string | number>(
                r as Record<string, unknown>,
                'RawReportID',
                'raw_report_id',
                'rawReportId'
            );
            return String(id) === state.selectedReport;
        });
        displayReportDetails(report || null);
    } else {
        clearReportDetails();
    }
});

elements.yearSelect.addEventListener('change', (e: Event) => {
    const target = e.target as HTMLSelectElement;
    state.selectedYear = target.value || null;
    populateReportSelect();

    if (state.selectedReport && elements.reportSelect.value) {
        const report = state.reports.find((r) => {
            const id = getFieldValue<string | number>(
                r as Record<string, unknown>,
                'RawReportID',
                'raw_report_id',
                'rawReportId'
            );
            return String(id) === state.selectedReport;
        });
        displayReportDetails(report || null);
    }
});

// Health check function
async function checkHealth(): Promise<HealthResponse> {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return await response.json();
    } catch (error) {
        throw error;
    }
}

// Initialize
async function init(): Promise<void> {
    try {
        // Check if API is available
        const health = await checkHealth();
        console.log('API health check:', health);

        // Load initial data
        await fetchCompanies();
    } catch (error) {
        const message = error instanceof Error ? error.message : 'Unknown error';
        showError(`Failed to connect to API: ${message}. Make sure the backend server is running on ${API_BASE_URL}`);
    }
}

// Start the app
init().catch(console.error);
