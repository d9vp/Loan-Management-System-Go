package statistics

type Statistics struct {
	ActiveCustomersCount       int     `json:"active_customers_count"`
	AverageCustomerSessionTime float64 `json:"average_customer_session_time"`
	TotalLoanApplications      int     `json:"total_loan_applications"`
	NPACount                   int     `json:"npa_count"`
	SchemeStats                []LoanSchemeStats
	OfficerStats               []LoanOfficerStats
}

type LoanSchemeStats struct {
	SchemeID                   uint `json:"scheme_id"`
	ApplicationCount           uint `json:"application_count"`
	NonPerformingAssetInScheme uint `json:"non_performing_asset_in_scheme"`
}

type LoanOfficerStats struct {
	OfficerID                     uint `json:"officer_id"`
	ApplicationCount              uint `json:"application_count"`
	NonPerformingAssetFromOfficer uint `json:"non_performing_asset_from_officer"`
}
