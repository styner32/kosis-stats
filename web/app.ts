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
  [key: string]: unknown;
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
  currentTab: "dashboard",
};

// DOM elements
const elements: DOMElements = {
  // Tabs
  tabButtons: document.querySelectorAll(".tab-btn"),
  dashboardView: document.getElementById("dashboard-view")!,
  reportsView: document.getElementById("reports-view")!,

  // Dashboard
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

  // Report List
  listCompanyInput: document.getElementById(
    "list-company-input"
  ) as HTMLInputElement,
  listCompanyList: document.getElementById(
    "list-company-list"
  ) as HTMLUListElement,
  dateStart: document.getElementById("date-start") as HTMLInputElement,
  dateEnd: document.getElementById("date-end") as HTMLInputElement,
  sortOrder: document.getElementById("sort-order") as HTMLSelectElement,
  reportListContainer: document.getElementById("report-list-container")!,
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
    const endpoint = search
      ? `/companies?search=${encodeURIComponent(search)}`
      : "/companies";
    const data = await apiRequest<CompaniesResponse | Company[]>(endpoint);
    state.companies = Array.isArray(data) ? data : data.companies || [];
    renderCompanyList();
  } catch (error) {
    console.error("Failed to load companies:", error);
    state.companies = [];
    renderCompanyList();
  }
}

async function fetchReports(
  corpCode: string = "",
  limit: number = 100
): Promise<void> {
  try {
    showLoading();
    const endpoint = corpCode
      ? `/reports/${corpCode}?limit=${limit}`
      : `/reports?limit=${limit}`;

    const data = await apiRequest<ReportsResponse | AnalysisRecord[]>(endpoint);
    const reports = Array.isArray(data)
      ? data
      : (data as { reports?: AnalysisRecord[] }).reports || [];
    state.reports = reports;
    updateUIWithReports();
    hideLoading();
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    showError(`Failed to load reports: ${message}`);
    state.reports = [];
    updateUIWithReports();
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
function updateUIWithReports(): void {
  // Update Dashboard
  populateReportSelect();
  populateYearSelect();

  // Update List View
  renderReportTable();
}

function renderCompanyList(): void {
  // Render to both lists (Dashboard and Reports View)
  [elements.companyList, elements.listCompanyList].forEach((list) => {
    if (!list) return;
    list.innerHTML = "";

    if (state.companies.length === 0) {
      list.innerHTML = '<li class="combobox-empty">No companies found</li>';
      // list.classList.remove("hidden"); // Don't auto-show empty
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

      list.appendChild(li);
    });
  });
}

function selectCompany(corpCode: string, corpName: string): void {
  state.selectedCompany = corpCode;

  // Update both inputs
  if (elements.companyInput) elements.companyInput.value = corpName;
  if (elements.listCompanyInput) elements.listCompanyInput.value = corpName;

  // Hide lists
  if (elements.companyList) elements.companyList.classList.add("hidden");
  if (elements.listCompanyList)
    elements.listCompanyList.classList.add("hidden");

  // Clear selections
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

function renderReportTable(): void {
  const container = elements.reportListContainer;
  if (!container) return;

  let reports = [...state.reports];

  // Filter by Date
  const startDate = elements.dateStart.value;
  const endDate = elements.dateEnd.value;

  if (startDate) {
    reports = reports.filter((r) => {
      const d = getFieldValue<string>(
        r as Record<string, unknown>,
        "CreatedAt",
        "created_at"
      );
      return d ? new Date(d) >= new Date(startDate) : false;
    });
  }

  if (endDate) {
    reports = reports.filter((r) => {
      const d = getFieldValue<string>(
        r as Record<string, unknown>,
        "CreatedAt",
        "created_at"
      );
      if (!d) return false;
      const rDate = new Date(d);
      const eDate = new Date(endDate);
      eDate.setHours(23, 59, 59, 999); // End of that day
      return rDate <= eDate;
    });
  }

  // Sort
  const sortOrder = elements.sortOrder.value; // 'asc' or 'desc'
  reports.sort((a, b) => {
    const d1 =
      getFieldValue<string>(
        a as Record<string, unknown>,
        "CreatedAt",
        "created_at"
      ) || "";
    const d2 =
      getFieldValue<string>(
        b as Record<string, unknown>,
        "CreatedAt",
        "created_at"
      ) || "";

    // If dates are missing (e.g. from /reports endpoint), sort might be unstable or grouped
    if (!d1 && !d2) return 0;
    if (!d1) return 1;
    if (!d2) return -1;

    const t1 = new Date(d1).getTime();
    const t2 = new Date(d2).getTime();
    return sortOrder === "asc" ? t1 - t2 : t2 - t1;
  });

  if (reports.length === 0) {
    container.innerHTML = '<div class="empty-state">No matching reports</div>';
    return;
  }

  let html = '<table class="data-table">';
  html +=
    "<thead><tr><th>Company</th><th>Name</th><th>Date</th><th>Receipt #</th></tr></thead>";
  html += "<tbody>";

  reports.forEach((r) => {
    const corpCode =
      getFieldValue(r as Record<string, unknown>, "CorpCode", "corp_code") ||
      state.selectedCompany ||
      "-";
    const name =
      getFieldValue(
        r as Record<string, unknown>,
        "ReportName",
        "report_name"
      ) || "-";

    const receipt =
      getFieldValue(
        r as Record<string, unknown>,
        "ReceiptNumber",
        "receipt_number"
      ) ||
      getFieldValue(
        r as Record<string, unknown>,
        "RawReportID",
        "raw_report_id"
      ) ||
      "-";

    const dateRaw = (receipt as string).substring(0, 8);
    const date = `${dateRaw.substring(0, 4)}-${dateRaw.substring(
      4,
      6
    )}-${dateRaw.substring(6, 8)}`;

    html += `<tr>
            <td>${corpCode}</td>
            <td>${name}</td>
            <td>${date}</td>
            <td>${receipt}</td>
        </tr>`;
  });

  html += "</tbody></table>";
  container.innerHTML = html;
}

let currentObjectUrl: string | null = null;

function displayReportDetails(
  data: AnalysisRecord | RawReport | null,
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
      const binaryString = atob(rawReport.BlobData.toString());
      const len = binaryString.length;
      const bytes = new Uint8Array(len);
      for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }

      let reportContent = "";
      const decoder = new TextDecoder("utf-8", { fatal: false });
      const tempContent = decoder.decode(bytes);
      const replacementCount = (tempContent.match(/\uFFFD/g) || []).length;
      const isLikelyBrokenUtf8 = replacementCount > tempContent.length * 0.01;
      const hasEucKrTag =
        tempContent.includes("charset=euc-kr") ||
        tempContent.includes('charset="euc-kr"') ||
        tempContent.includes("charset='euc-kr'");

      if (hasEucKrTag && isLikelyBrokenUtf8) {
        try {
          const eucDecoder = new TextDecoder("euc-kr");
          reportContent = eucDecoder.decode(bytes);
        } catch (e) {
          reportContent = tempContent;
        }
      } else {
        reportContent = tempContent;
      }

      reportContent = reportContent.replace(
        /(charset\s*=\s*["']?)euc-kr(["']?)/gi,
        "$1utf-8$2"
      );

      const lowerContent = reportContent.toLowerCase();
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
        const searchScript = `
          <style>
            #report-search-bar { position: fixed; top: 0; left: 0; right: 0; background: #f8f9fa; padding: 10px; border-bottom: 1px solid #ddd; z-index: 10000; display: flex; gap: 10px; align-items: center; font-family: system-ui; }
            #report-search-input { padding: 6px; flex-grow: 1; border: 1px solid #ccc; border-radius: 4px; }
            body { padding-top: 60px !important; }
            .match-row { background-color: #e8f0fe; }
          </style>
          <div id="report-search-bar"><input type="text" id="report-search-input" placeholder="Search..."></div>
        `;
        if (reportContent.includes("</body>")) {
          reportContent = reportContent.replace(
            "</body>",
            searchScript + "</body>"
          );
        } else {
          reportContent += searchScript;
        }

        const blob = new Blob([reportContent], {
          type: "text/html; charset=utf-8",
        });
        currentObjectUrl = URL.createObjectURL(blob);
        html += `<iframe src="${currentObjectUrl}" sandbox="allow-scripts" style="width: 100%; height: 600px; border: 1px solid #ccc; background-color: white;"></iframe>`;
        html +=
          '<p class="note" style="font-size: 0.8em; color: #666; margin-top: 5px;">Rendering as HTML.</p>';
      } else if (looksLikeXml) {
        const blob = new Blob([reportContent], {
          type: "text/xml; charset=utf-8",
        });
        currentObjectUrl = URL.createObjectURL(blob);
        html += `<iframe src="${currentObjectUrl}" sandbox="allow-scripts" style="width: 100%; height: 600px; border: 1px solid #ccc; background-color: white;"></iframe>`;
        html +=
          '<p class="note" style="font-size: 0.8em; color: #666; margin-top: 5px;">Rendering as XML.</p>';
      } else {
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

// Tab Switching
elements.tabButtons.forEach((btn) => {
  btn.addEventListener("click", () => {
    const tabName = (btn as HTMLElement).dataset.tab;
    if (tabName === "dashboard" || tabName === "reports") {
      state.currentTab = tabName;

      // Update Buttons
      elements.tabButtons.forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");

      // Update Views
      if (tabName === "dashboard") {
        elements.dashboardView.classList.remove("hidden");
        elements.reportsView.classList.add("hidden");

        if (state.selectedCompany) {
          fetchReports(state.selectedCompany);
        } else {
          state.reports = [];
          updateUIWithReports();
        }
      } else {
        elements.dashboardView.classList.add("hidden");
        elements.reportsView.classList.remove("hidden");
        fetchReports();
      }
    }
  });
});

// Company Input (Dashboard)
if (elements.companyInput) {
  elements.companyInput.addEventListener(
    "input",
    debounce((e: Event) => {
      const target = e.target as HTMLInputElement;
      fetchCompanies(target.value);
    }, 300)
  );

  elements.companyInput.addEventListener("focus", () => {
    elements.companyList.classList.remove("hidden");
    if (state.companies.length === 0) fetchCompanies();
  });

  // Click outside (Generic for both lists)
  document.addEventListener("click", (e: Event) => {
    const target = e.target as HTMLElement;
    // Dashboard List
    if (
      !elements.companyInput.contains(target) &&
      !elements.companyList.contains(target)
    ) {
      elements.companyList.classList.add("hidden");
    }
    // Reports List
    if (
      !elements.listCompanyInput.contains(target) &&
      !elements.listCompanyList.contains(target)
    ) {
      elements.listCompanyList.classList.add("hidden");
    }
  });
}

// Company Input (Reports List)
if (elements.listCompanyInput) {
  elements.listCompanyInput.addEventListener(
    "input",
    debounce((e: Event) => {
      const target = e.target as HTMLInputElement;
      fetchCompanies(target.value);
    }, 300)
  );

  elements.listCompanyInput.addEventListener("focus", () => {
    elements.listCompanyList.classList.remove("hidden");
    if (state.companies.length === 0) fetchCompanies();
  });
}

// Dashboard Filters
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
    // Re-select if still valid
    const report = state.reports.find((r) => {
      const id = getFieldValue<string | number>(
        r as Record<string, unknown>,
        "RawReportID",
        "raw_report_id",
        "rawReportId"
      );
      return String(id) === state.selectedReport;
    });
    if (report) displayReportDetails(report, null);
  }
});

// Report List Filters
[elements.dateStart, elements.dateEnd, elements.sortOrder].forEach((el) => {
  if (el) {
    el.addEventListener("change", () => {
      renderReportTable();
    });
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
    const health = await checkHealth();
    console.log("API health check:", health);
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
