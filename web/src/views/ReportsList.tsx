import { useState, useEffect } from "react";
import { getFieldValue } from "../types";
import type { Company, AnalysisRecord } from "../types";
import { apiRequest } from "../api";
import { CompanySelect } from "../components/CompanySelect";
import { ReportDetail } from "../components/ReportDetail";

export function ReportsList() {
  const [reports, setReports] = useState<AnalysisRecord[]>([]);
  const [loading, setLoading] = useState(false);

  // Filters
  const [selectedCompany, setSelectedCompany] = useState<Company | null>(null);
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  // Expanded Row State
  const [expandedRowIndex, setExpandedRowIndex] = useState<number | null>(null);

  // Fetch Logic
  const fetchAll = async () => {
    setLoading(true);
    try {
      // Build query - note: backend logic in original code handles corpCode optional
      // Here we can use the same logic or just filter client side if we fetch all?
      // Original app.ts logic: fetch `/reports` (all) or `/reports/:code`.
      let endpoint = "/reports?limit=100";
      if (selectedCompany) {
        const code = getFieldValue<string>(
          selectedCompany as unknown as Record<string, unknown>,
          "corp_code"
        );
        if (code) {
          endpoint += `&corp_code=${code}`;
        }
      }

      const data = await apiRequest<
        { reports: AnalysisRecord[] } | AnalysisRecord[]
      >(endpoint);
      setReports(Array.isArray(data) ? data : data.reports || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAll();
    // Reset expansion on refetch
    setExpandedRowIndex(null);
  }, [selectedCompany]);

  // Client-side Filtering & Sorting
  const processedReports = reports
    .filter((r) => {
      const d = getFieldValue<string>(
        r as unknown as Record<string, unknown>,
        "created_at"
      );
      if (!d) return true; // Keep if no date? Or filter out? Legacy kept them unless range specified.
      const date = new Date(d);

      if (startDate && date < new Date(startDate)) return false;
      if (endDate) {
        const e = new Date(endDate);
        e.setHours(23, 59, 59);
        if (date > e) return false;
      }
      return true;
    })
    .sort((a, b) => {
      const d1 = getFieldValue<string>(
        a as unknown as Record<string, unknown>,
        "created_at"
      );
      const d2 = getFieldValue<string>(
        b as unknown as Record<string, unknown>,
        "created_at"
      );
      if (!d1) return 1;
      if (!d2) return -1;
      return sortOrder === "asc" ? d1.localeCompare(d2) : d2.localeCompare(d1);
    });

  const toggleRow = (index: number) => {
    setExpandedRowIndex(expandedRowIndex === index ? null : index);
  };

  return (
    <div className="tab-content active">
      <div className="filters">
        <div className="filter-group">
          <label>Company:</label>
          <CompanySelect
            selectedCompany={selectedCompany}
            onSelect={setSelectedCompany}
            placeholder="Search company (optional)..."
          />
        </div>
        <div className="filter-group">
          <label>Start Date:</label>
          <input
            type="date"
            value={startDate}
            onChange={(e) => setStartDate(e.target.value)}
          />
        </div>
        <div className="filter-group">
          <label>End Date:</label>
          <input
            type="date"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
          />
        </div>
        <div className="filter-group">
          <label>Sort:</label>
          <select
            value={sortOrder}
            onChange={(e) => setSortOrder(e.target.value as "asc" | "desc")}
          >
            <option value="desc">Newest First</option>
            <option value="asc">Oldest First</option>
          </select>
        </div>
      </div>

      <div className="results">
        {loading ? (
          <div className="loading">
            <div className="spinner"></div>Loading...
          </div>
        ) : processedReports.length === 0 ? (
          <div className="empty-state">No matching reports</div>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th>Corp Code</th>
                <th>Corp Name</th>
                <th>Name</th>
                <th>Date</th>
                <th>Receipt #</th>
              </tr>
            </thead>
            <tbody>
              {processedReports.map((r, i) => {
                const code =
                  getFieldValue<string>(
                    r as unknown as Record<string, unknown>,
                    "corp_code"
                  ) || "-";
                const corpName =
                  getFieldValue<string>(
                    r as unknown as Record<string, unknown>,
                    "corp_name"
                  ) || "-";
                const name =
                  getFieldValue<string>(
                    r as unknown as Record<string, unknown>,
                    "report_name"
                  ) || "-";
                const date = getFieldValue<string>(
                  r as unknown as Record<string, unknown>,
                  "receipt_date"
                );
                const receipt =
                  getFieldValue<string>(
                    r as unknown as Record<string, unknown>,
                    "receipt_number"
                  ) ||
                  getFieldValue<string | number>(
                    r as unknown as Record<string, unknown>,
                    "raw_report_id"
                  ) ||
                  "-";

                return (
                  <>
                    <tr
                      key={i}
                      onClick={() => toggleRow(i)}
                      style={{
                        cursor: "pointer",
                        background:
                          expandedRowIndex === i ? "#f1f5f9" : "inherit",
                      }}
                    >
                      <td>{code}</td>
                      <td>{corpName}</td>
                      <td>{name}</td>
                      <td>{date}</td>
                      <td>{receipt}</td>
                    </tr>
                    {expandedRowIndex === i && (
                      <tr className="detail-row">
                        <td colSpan={5}>
                          <div className="detail-content">
                            <ReportDetail report={r} />
                          </div>
                        </td>
                      </tr>
                    )}
                  </>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
