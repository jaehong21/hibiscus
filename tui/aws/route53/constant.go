package route53

// TAB_NAME
const (
	ROUTE53_HOSTED_ZONE_TAB = iota
	ROUTE53_RECORD_TAB
)

// LOADING_MSG
const (
	FETCHING_HOSTED_ZONES_MSG = "fetching hosted zones..."
	FETCHING_RECORDS_MSG      = "fetching records..."

	FILTERING_HOSTED_ZONES_MSG = "searching hosted zones..."
	FILTERING_RECORDS_MSG      = "searching records..."
)
