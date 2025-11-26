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
  RawReportResponse,
} from "./types";

type AnalysisRecord = {
  RawReportID?: number;
  Analysis?: unknown;
  CreatedAt?: string;
  created_at?: string;
};

// Configuration
const API_BASE_URL = "http://localhost:8080";
const API_VERSION = "/api/v1";

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
  companyInput: document.getElementById(
    "company-combobox-input"
  ) as HTMLInputElement,
  companyList: document.getElementById(
    "company-combobox-list"
  ) as HTMLUListElement,
  reportSelect: document.getElementById("report-select") as HTMLSelectElement,
  yearSelect: document.getElementById("year-select") as HTMLSelectElement,
  loading: document.getElementById("loading")!,
  error: document.getElementById("error")!,
  results: document.getElementById("results")!,
  reportDetails: document.getElementById("report-details")!,
};

// Debounce utility
function debounce<T extends (...args: any[]) => void>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
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
async function apiRequest<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  const url = `${API_BASE_URL}${API_VERSION}${endpoint}`;

  const defaultOptions: RequestOptions = {
    headers: {
      "Content-Type": "application/json",
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
    console.error("API request failed:", error);
    throw error;
  }
}

async function fetchCompanies(search: string = ""): Promise<void> {
  try {
    // Don't show global loading for search interactions to keep UI responsive
    // showLoading();
    const endpoint = search
      ? `/companies?search=${encodeURIComponent(search)}`
      : "/companies";
    const data = await apiRequest<CompaniesResponse | Company[]>(endpoint);
    state.companies = Array.isArray(data) ? data : data.companies || [];
    renderCompanyList();
    // hideLoading();
  } catch (error) {
    console.error("Failed to load companies:", error);
    state.companies = [];
    renderCompanyList();
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
    const data = await apiRequest<ReportsResponse | AnalysisRecord[]>(
      `/reports/${corpCode}`
    );
    const reports = Array.isArray(data)
      ? data
      : (data as { reports?: AnalysisRecord[] }).reports || [];
    state.reports = reports;
    populateReportSelect();
    populateYearSelect();
    hideLoading();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    showError(`Failed to load reports: ${message}`);
    state.reports = [];
    populateReportSelect();
  }
}

async function fetchRawReport(
  corpCode: string,
  reportId: string
): Promise<RawReport | null> {
  if (!corpCode || !reportId) {
    return null;
  }

  try {
    showLoading();
    const data = await apiRequest<RawReportResponse>(
      `/reports/${corpCode}/${reportId}`
    );
    hideLoading();
    return { BlobData: data.raw_report } as RawReport;
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    showError(`Failed to load raw report: ${message}`);
    return null;
  }
}

// UI update functions
function renderCompanyList(): void {
  elements.companyList.innerHTML = "";

  if (state.companies.length === 0) {
    elements.companyList.innerHTML =
      '<li class="combobox-empty">No companies found</li>';
    elements.companyList.classList.remove("hidden");
    return;
  }

  state.companies.forEach((company) => {
    const li = document.createElement("li");
    li.className = "combobox-item";

    const corpCode =
      getFieldValue<string>(
        company as Record<string, unknown>,
        "CorpCode",
        "corp_code",
        "corpCode",
        "id"
      ) || "";
    const corpName =
      getFieldValue<string>(
        company as Record<string, unknown>,
        "CorpName",
        "corp_name",
        "corpName",
        "name"
      ) || `Company ${corpCode}`;

    li.textContent = corpName;
    li.dataset.value = corpCode;

    if (state.selectedCompany === corpCode) {
      li.classList.add("selected");
    }

    li.addEventListener("click", () => {
      selectCompany(corpCode, corpName);
    });

    elements.companyList.appendChild(li);
  });

  elements.companyList.classList.remove("hidden");
}

function selectCompany(corpCode: string, corpName: string): void {
  state.selectedCompany = corpCode;
  elements.companyInput.value = corpName;
  elements.companyList.classList.add("hidden");

  // Clear reports
  state.selectedReport = null;
  elements.reportSelect.value = "";
  clearReportDetails();

  fetchReports(corpCode);
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
          "CreatedAt",
          "created_at",
          "createdAt"
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
    elements.reportSelect.innerHTML =
      '<option value="">No reports available</option>';
    elements.reportSelect.disabled = true;
    return;
  }

  elements.reportSelect.disabled = false;

  filteredReports.forEach((report) => {
    const option = document.createElement("option");
    const rawReportId =
      getFieldValue<string | number>(
        report as Record<string, unknown>,
        "RawReportID",
        "raw_report_id",
        "rawReportId"
      ) || "";
    const date = getFieldValue<string>(
      report as Record<string, unknown>,
      "CreatedAt",
      "created_at",
      "createdAt"
    );
    let dateStr = "";
    if (date) {
      try {
        dateStr = new Date(date).toLocaleDateString();
      } catch {
        dateStr = date;
      }
    }
    option.value = String(rawReportId);
    option.textContent = `Report ${rawReportId}${
      dateStr ? " - " + dateStr : ""
    }`;
    elements.reportSelect.appendChild(option);
  });
}

function populateYearSelect(): void {
  const years = new Set<string>();

  state.reports.forEach((report) => {
    const date = getFieldValue<string>(
      report as Record<string, unknown>,
      "CreatedAt",
      "created_at",
      "createdAt"
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
    const option = document.createElement("option");
    option.value = year;
    option.textContent = year;
    elements.yearSelect.appendChild(option);
  });
}

let currentObjectUrl: string | null = null;

function displayReportDetails(
  data: AnalysisRecord | null,
  rawReport: RawReport | null
): void {
  if (!data) {
    elements.reportDetails.innerHTML =
      '<div class="empty-state">No report details available</div>';
    return;
  }

  let html = '<div class="report-card">';

  html += "<h3>Report Details</h3>";
  html += '<div class="meta">';

  const rawReportId = getFieldValue<string | number>(
    data as Record<string, unknown>,
    "RawReportID",
    "raw_report_id",
    "rawReportId"
  );
  const receiptNumber = getFieldValue<string>(
    data as Record<string, unknown>,
    "ReceiptNumber",
    "receipt_number",
    "receiptNumber"
  );
  const createdAt = getFieldValue<string>(
    data as Record<string, unknown>,
    "CreatedAt",
    "created_at",
    "createdAt"
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
  html += "</div>";

  const analysisData = getFieldValue<unknown>(
    data as Record<string, unknown>,
    "Analysis",
    "analysis"
  );
  let parsedAnalysis: Record<string, unknown> | unknown = analysisData;
  if (typeof analysisData === "string") {
    try {
      parsedAnalysis = JSON.parse(analysisData);
    } catch {
      parsedAnalysis = analysisData;
    }
  }

  if (parsedAnalysis) {
    html += '<div class="json-viewer">';
    html += "<h4>Analysis</h4>";
    html += `<pre>${escapeHtml(JSON.stringify(parsedAnalysis, null, 2))}</pre>`;
    html += "</div>";
  } else {
    html += '<div class="empty-state">No analysis available</div>';
  }

  if (rawReport && rawReport.BlobData) {
    // Clean up previous object URL
    if (currentObjectUrl) {
      URL.revokeObjectURL(currentObjectUrl);
      currentObjectUrl = null;
    }

    html += '<div class="raw-report-viewer">';
    html += "<h4>Raw Report Content</h4>";

    try {
      // Backend now returns Base64 string to preserve original bytes
      const binaryString = atob(rawReport.BlobData.toString());
      const len = binaryString.length;
      const bytes = new Uint8Array(len);
      for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }

      // Decode bytes to string, handling EUC-KR if necessary
      // We try to detect if it's likely EUC-KR (common for DART)
      let reportContent = "";
      const decoder = new TextDecoder("utf-8", { fatal: false });
      const tempContent = decoder.decode(bytes);

      // Check if UTF-8 decoding resulted in many replacement characters ()
      // If the file is actually EUC-KR but we decode as UTF-8, we'll get many of these.
      // If the file is UTF-8 (even with a lying meta tag), we'll get few/none.
      const replacementCount = (tempContent.match(/\uFFFD/g) || []).length;

      // Heuristic: If > 1% of characters are replacements, it's probably not UTF-8
      const isLikelyBrokenUtf8 = replacementCount > tempContent.length * 0.01;

      // Check for EUC-KR meta tag
      const hasEucKrTag =
        tempContent.includes("charset=euc-kr") ||
        tempContent.includes('charset="euc-kr"') ||
        tempContent.includes("charset='euc-kr'");

      if (hasEucKrTag && isLikelyBrokenUtf8) {
        try {
          console.log(
            "Detected EUC-KR tag and broken UTF-8. Re-decoding as EUC-KR."
          );
          const eucDecoder = new TextDecoder("euc-kr");
          reportContent = eucDecoder.decode(bytes);
        } catch (e) {
          console.warn("Failed to decode as EUC-KR, falling back to UTF-8", e);
          reportContent = tempContent;
        }
      } else {
        // It's either valid UTF-8 (regardless of tag) or we don't know what it is.
        // If it was valid UTF-8 but had the EUC-KR tag (the lying tag case),
        // we stick with tempContent (which is correct) and just fix the tag below.
        if (hasEucKrTag) {
          console.log(
            "Detected EUC-KR tag but content appears to be valid UTF-8. Ignoring tag."
          );
        }
        reportContent = tempContent;
      }

      // Always fix the meta tag to utf-8 if we have rendered it as such (which we have, effectively)
      // Matches: charset=euc-kr, charset="euc-kr", charset='euc-kr'
      reportContent = reportContent.replace(
        /(charset\s*=\s*["']?)euc-kr(["']?)/gi,
        "$1utf-8$2"
      );

      const lowerContent = reportContent.toLowerCase();

      // Improved content detection
      // If it contains common HTML structural tags, treat it as HTML regardless of XML headers
      const hasHtmlTags =
        lowerContent.includes("<table") ||
        lowerContent.includes("<body") ||
        lowerContent.includes("<div") ||
        lowerContent.includes("<span");

      const firstChars = reportContent.substring(0, 100).trim().toLowerCase();
      const looksLikeXml =
        firstChars.startsWith("<?xml") ||
        firstChars.includes("<xbrl") ||
        firstChars.includes("<dart-receipt");

      if (hasHtmlTags) {
        // Inject Search/Filter UI and Logic
        const searchScript = `
          <style>
            #report-search-bar {
              position: fixed;
              top: 0;
              left: 0;
              right: 0;
              background: #f8f9fa;
              padding: 10px;
              border-bottom: 1px solid #ddd;
              box-shadow: 0 2px 5px rgba(0,0,0,0.1);
              z-index: 10000;
              display: flex;
              gap: 10px;
              align-items: center;
              font-family: system-ui, -apple-system, sans-serif;
            }
            #report-search-input {
              padding: 6px 12px;
              border: 1px solid #ccc;
              border-radius: 4px;
              flex-grow: 1;
              font-size: 14px;
            }
            #report-search-stats {
              font-size: 12px;
              color: #666;
            }
            body {
              padding-top: 60px !important; /* Make space for search bar */
            }
            .highlight {
              background-color: #fff3cd;
              padding: 2px;
            }
            .match-row {
              background-color: #e8f0fe;
            }
          </style>
          <div id="report-search-bar">
            <input type="text" id="report-search-input" placeholder="Search text to filter rows... (e.g., 'Sales', 'Revenue')">
            <span id="report-search-stats"></span>
          </div>
          <script>
            document.addEventListener('DOMContentLoaded', () => {
              const input = document.getElementById('report-search-input');
              const stats = document.getElementById('report-search-stats');
              const tables = document.querySelectorAll('table');
              let timeout = null;

              input.addEventListener('input', (e) => {
                clearTimeout(timeout);
                timeout = setTimeout(() => {
                  const term = e.target.value.toLowerCase();
                  let matchCount = 0;
                  let totalRows = 0;

                  tables.forEach(table => {
                    const rows = table.querySelectorAll('tr');
                    
                    rows.forEach(row => {
                      // Skip header rows if possible (heuristic)
                      if(row.querySelector('th')) return;
                      
                      totalRows++;
                      const text = row.textContent.toLowerCase();
                      
                      if (term === '') {
                        row.style.display = '';
                        row.classList.remove('match-row');
                      } else if (text.includes(term)) {
                        row.style.display = '';
                        row.classList.add('match-row');
                        matchCount++;
                      } else {
                        row.style.display = 'none';
                        row.classList.remove('match-row');
                      }
                    });
                  });

                  if (term !== '') {
                    stats.textContent = \`Found \${matchCount} matches\`;
                  } else {
                    stats.textContent = '';
                  }
                }, 300);
              });
            });
          </script>
        `;

        // Insert before </body>, or append if not found
        if (reportContent.includes("</body>")) {
          reportContent = reportContent.replace(
            "</body>",
            searchScript + "</body>"
          );
        } else {
          reportContent += searchScript;
        }

        // Render as HTML in iframe
        // Use UTF-8 as we have normalized the string
        const blob = new Blob([reportContent], {
          type: "text/html; charset=utf-8",
        });
        currentObjectUrl = URL.createObjectURL(blob);

        html += `<iframe 
              src="${currentObjectUrl}" 
              sandbox="allow-scripts"
              style="width: 100%; height: 600px; border: 1px solid #ccc; background-color: white;"
            ></iframe>`;
        html +=
          '<p class="note" style="font-size: 0.8em; color: #666; margin-top: 5px;">Rendering as HTML.</p>';
      } else if (looksLikeXml) {
        // Render as XML
        const blob = new Blob([reportContent], {
          type: "text/xml; charset=utf-8",
        });
        currentObjectUrl = URL.createObjectURL(blob);

        html += `<iframe 
              src="${currentObjectUrl}" 
              sandbox="allow-scripts"
              style="width: 100%; height: 600px; border: 1px solid #ccc; background-color: white;"
            ></iframe>`;
        html +=
          '<p class="note" style="font-size: 0.8em; color: #666; margin-top: 5px;">Rendering as XML.</p>';
      } else {
        // Fallback to text/plain
        html += `<pre style="white-space: pre-wrap; max-height: 600px; overflow: auto;">${escapeHtml(
          reportContent
        )}</pre>`;
      }
    } catch (e) {
      console.error("Failed to render report:", e);
      html += `<div class="error-state">Failed to render raw report content. Error: ${e}</div>`;
    }

    html += "</div>";
  } else {
    html += '<div class="action-area">';
    html += `<button id="btn-load-raw" class="primary-button" data-report-id="${rawReportId}">Load Raw Report</button>`;
    html += "</div>";
  }

  html += "</div>";
  elements.reportDetails.innerHTML = html;
  elements.results.classList.remove("hidden");

  // Attach event listener to the button if it exists
  const loadBtn = document.getElementById("btn-load-raw");
  if (loadBtn) {
    loadBtn.addEventListener("click", async () => {
      if (state.selectedCompany) {
        const idToFetch = loadBtn.getAttribute("data-report-id");
        if (idToFetch) {
          const fetchedRawReport = await fetchRawReport(
            state.selectedCompany,
            idToFetch
          );
          displayReportDetails(data, fetchedRawReport);
        }
      }
    });
  }
  // Removed Prism re-highlighting logic as we are using iframes primarily now
}

function clearReportDetails(): void {
  elements.reportDetails.innerHTML = "";
  elements.results.classList.add("hidden");
}

function showLoading(): void {
  state.loading = true;
  elements.loading.classList.remove("hidden");
  elements.error.classList.add("hidden");
  elements.results.classList.add("hidden");
}

function hideLoading(): void {
  state.loading = false;
  elements.loading.classList.add("hidden");
}

function showError(message: string): void {
  state.error = message;
  elements.error.textContent = message;
  elements.error.classList.remove("hidden");
  elements.loading.classList.add("hidden");
}

function escapeHtml(text: string): string {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

// Event handlers
if (elements.companyInput) {
  // Input handler for search
  elements.companyInput.addEventListener(
    "input",
    debounce((e: Event) => {
      const target = e.target as HTMLInputElement;
      fetchCompanies(target.value);
    }, 300)
  );

  // Focus handler to show list
  elements.companyInput.addEventListener("focus", () => {
    if (state.companies.length > 0) {
      elements.companyList.classList.remove("hidden");
    } else {
      fetchCompanies(elements.companyInput.value);
    }
  });

  // Click outside to close
  document.addEventListener("click", (e: Event) => {
    const target = e.target as HTMLElement;
    if (
      !elements.companyInput.contains(target) &&
      !elements.companyList.contains(target)
    ) {
      elements.companyList.classList.add("hidden");
    }
  });
}

elements.reportSelect.addEventListener("change", async (e: Event) => {
  const target = e.target as HTMLSelectElement;
  state.selectedReport = target.value || null;

  if (state.selectedReport && state.selectedCompany) {
    const report = state.reports.find((r) => {
      const id = getFieldValue<string | number>(
        r as Record<string, unknown>,
        "RawReportID",
        "raw_report_id",
        "rawReportId"
      );
      return String(id) === state.selectedReport;
    });

    if (report) {
      displayReportDetails(report, null);
    } else {
      clearReportDetails();
    }
  } else {
    clearReportDetails();
  }
});

elements.yearSelect.addEventListener("change", (e: Event) => {
  const target = e.target as HTMLSelectElement;
  state.selectedYear = target.value || null;
  populateReportSelect();

  if (
    state.selectedReport &&
    elements.reportSelect.value &&
    state.selectedCompany
  ) {
    const report = state.reports.find((r) => {
      const id = getFieldValue<string | number>(
        r as Record<string, unknown>,
        "RawReportID",
        "raw_report_id",
        "rawReportId"
      );
      return String(id) === state.selectedReport;
    });
    if (report) {
      displayReportDetails(report, null);
    }
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
    console.log("API health check:", health);

    // Load initial data
    await fetchCompanies();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    showError(
      `Failed to connect to API: ${message}. Make sure the backend server is running on ${API_BASE_URL}`
    );
  }
}

// Start the app
init().catch(console.error);
