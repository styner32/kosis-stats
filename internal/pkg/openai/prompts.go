package openai

const (
	systemPrompt = `너의 임무는 한국 DART 공시 원문에서 숫자/사실을 추출해
		정해진 JSON 스키마로만 출력하는 것이다.
		추론은 최소화하고, 문서에 명시된 값만 사용하라.
		출력은 반드시 유효한 JSON 한개만 반환. 설명문 금지.`

	// Securities Issuance Terms
	securitiesIssuanceTermsSchema = `
		스키마 필드: 
		doc_id, doc_type, issuer{name, areg_cik}, dates{first_filed, correction_announced}, 
		tranches[{name, seniority, amount_krw, coupon_before_pct, coupon_after_pct, coupon_delta_bp, annual_interest_delta_krw}], 
		totals{amount_krw, wac_before_pct, wac_after_pct, wac_delta_bp, annual_interest_delta_krw},
		reason_of_correction, spread_after_bp, 
		impact_score{equity_impact_0to5, credit_impact_0to5, liquidity_impact_0to5}, notes.
		단위와 포맷:
		- 금액은 정수 KRW.
		- 퍼센트는 소수점 4자리 이내.
		- bp 계산: (정정후 - 정정전)*10000. 
		- 연간 이자 증분은 amount_krw*(coupon_after - coupon_before).
	`
	additionalSecuritiesIssuanceTermsSchema = `
- "증권발행조건확정" 표에서 각 트랜치의 정정전/정정후 금리를 읽어라.
- "모집 또는 매출금액" 문단에서 각 트랜치 금액을 읽어라.
- seniority는 '선순위/후순위'로 추론 가능할 때만 표기하고, 없으면 null.
- impact_score는 다음 기준의 휴리스틱을 적용:
  - equity_impact: SPC 금리 확정 정정은 0~1.
  - credit_impact: 금리 변화가 25bp 미만이면 1, 25~75bp면 2, 그 이상 3.
  - liquidity_impact: 모집액이 1천억 미만이면 0.5, 1천억 이상이면 1.0.`

	supplySchema = `스키마 필드: {
	doc_id, corp_name, report_title, event_code, amendment{reason, prev_amount_krw, new_amount_krw, prev_ratio_to_sales, new_ratio_to_sales}, contract{name, counterparty, amount_krw, company_recent_sales_krw, counterparty_recent_sales_krw, country, term_from, term_to, progress_pct}, score{direction: string, magnitude: float64, confidence: float64, horizons: []string, rationale_short: string}}.
	단위와 포맷:
	- 금액은 정수 KRW.
	- 퍼센트는 소수점 4자리 이내.`

	reportSchema = `스키마 필드: {
		company_name: string,
		period_start_date: string (YYYY-MM-DD),
		period_end_date: string (YYYY-MM-DD),
		submission_date: string (YYYY-MM-DD),
		share_info: {
			issued_common_shares: int64,
			par_value_krw: int64,
			capital_million_krw: int64,
			outstanding_common_shares: int64
		},
		sales_breakdown_million_krw: {
			segment: string,
			export: int64,
			domestic: int64,
			total: int64
		},
		consolidated_financials_million_krw: {
			balance_sheet: {
				"period_YYYY_MM_DD": {
					total_assets: int64,
					total_liabilities: int64,
					total_equity: int64,
					equity_attributable_to_owners: int64,
					non_controlling_interests: int64,
					capital: int64
				}
			},
			income_statement: {
				"period_YYYY_H1" 또는 "period_YYYY": {
					sales: int64,
					operating_income: int64,
					net_income: int64,
					owners_net_income: int64
				}
			}
		},
		separate_financials_million_krw: {
			balance_sheet: {
				"period_YYYY_MM_DD": {
					total_assets: int64,
					total_liabilities: int64,
					total_equity: int64,
					capital: int64
				}
			},
			income_statement: {
				"period_YYYY_H1" 또는 "period_YYYY": {
					sales: int64,
					operating_income: int64,
					net_income: int64
				}
			}
		},
		fx_exposure_million_krw: {
			current_period_end: {
				assets: {
					USD: int64,
					EUR: int64,
					JPY: int64,
					CNY_etc: int64
				},
				liabilities: {
					USD: int64,
					EUR: int64,
					JPY: int64,
					CNY_etc: int64
				}
			}
		},
		fx_sensitivity_10pct_million_krw: {
			current_period_end: {
				USD: {
					profit_loss_if_up: int64,
					profit_loss_if_down: int64
				},
				EUR: {
					profit_loss_if_up: int64,
					profit_loss_if_down: int64
				},
				JPY: {
					profit_loss_if_up: int64,
					profit_loss_if_down: int64
				},
				CNY_etc: {
					profit_loss_if_up: int64,
					profit_loss_if_down: int64
				}
			}
		},
		derivatives_valuation_effects_million_krw: {
			forward_fx_loss: int64,
			cross_currency_swap_loss: int64,
			total_loss: int64
		},
		production_capacity: {
			current_half_year: {
				capacity_million_krw: int64,
				production_million_krw: int64,
				utilization_percent: float64
			},
			prior_year: {
				utilization_percent: float64
			}
		},
		rnd: {
			expenses_million_krw: {
				period_2025_H1_total: int64
			}
		},
		market_share: {
			product: string,
			period_2025_H1_percent: float64,
			period_2024_percent: float64,
			period_2023_percent: float64
		},
		capex: {
			amount_hundred_million_krw: int64,
			period: string
		},
		cash_flows_consolidated_million_krw: {
			period_2025_H1: {
				operating: int64,
				investing: int64,
				financing: int64,
				ending_cash: int64,
				beginning_cash: int64
			}
		},
		credit_ratings: [
			{
				date: string,
				agency: string,
				subject: string,
				rating: string
			}
		]
	}.
	단위와 포맷:
	- 모든 금액은 정수로 표시하며, 단위는 백만원(KRW)입니다. 단, capex의 amount_hundred_million_krw는 억원 단위입니다.
	- 날짜는 YYYY-MM-DD 형식으로 표시합니다.
	- 퍼센트는 소수점 2자리 이내의 float64로 표시합니다 (예: 85.5).
	- 문서에 명시되지 않은 필드는 null이 아닌 0 또는 빈 문자열로 표시합니다.
	- credit_ratings는 배열이며, 정보가 없으면 빈 배열 []로 표시합니다.
	- consolidated_financials_million_krw와 separate_financials_million_krw의 balance_sheet와 income_statement는 객체(map) 형태이며, 키는 기간을 나타내는 문자열입니다.
	- balance_sheet의 키 형식: "period_YYYY_MM_DD" (예: "period_2025_06_30", "period_2024_12_31")
	- income_statement의 키 형식: "period_YYYY_H1" (반기) 또는 "period_YYYY" (연간) (예: "period_2025_H1", "period_2024")
	- 각 기간별로 문서에 있는 모든 재무제표 데이터를 해당 키로 추출합니다.
	- 연결재무제표와 별도재무제표를 구분하여 정확히 추출합니다.
	- 외화노출액과 외화민감도는 각 통화별로 구분하여 추출합니다.`

	additionalReportSchema = `
	- 회사명, 보고기간, 제출일자는 문서 상단에서 확인합니다.
	- 주식정보는 발행주식수, 액면가, 자본금, 유통주식수를 정확히 추출합니다.
	- 매출구성은 사업부문별, 수출/내수 구분하여 추출합니다.
	- 재무제표는 연결/별도 구분하여 각 기간별로 추출합니다.
	- 외화노출은 자산/부채별, 통화별로 구분하여 추출합니다.
	- 외화민감도는 각 통화별로 10% 상승/하락 시 손익을 추출합니다.
	- 파생상품평가는 선물환 손실, 통화스왑 손실, 합계를 추출합니다.
	- 생산능력은 당기 반기와 전년도 가동률을 추출합니다.
	- R&D 비용은 당기 반기 합계를 추출합니다.
	- 시장점유율은 제품명과 각 기간별 점유율을 추출합니다.
	- 자본지출은 금액(억원)과 해당 기간을 추출합니다.
	- 현금흐름은 영업/투자/재무 활동별로 추출합니다.
	- 신용등급은 날짜, 평가기관, 평가대상, 등급을 추출합니다.
	- 모든 숫자는 문서에 명시된 값만 사용하며, 계산이나 추론은 하지 않습니다.`

	defaultSchema = `
	스키마 필드: {
		company_name: string,
		date: string (YYYY-MM-DD),
		type: string,
		summary: string,
	}`
	defaultAdditionalSchema = `
	- 회사명, 날짜, 유형, 요약은 문서 상단에서 확인합니다.
	- 요약은 문서 내용을 요약한 문자열로 추출합니다.
	- 유형은 문서 유형을 추출합니다.
	- 날짜는 문서 날짜를 추출합니다.
	- 회사명은 문서 회사명을 추출합니다.
	- 해당 문서에 맞는 JSON 스키마를 추출합니다.
	`
)
