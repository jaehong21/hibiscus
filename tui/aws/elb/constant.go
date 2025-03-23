package elb

// TAB_NAME
const (
	ELB_LOADBALANCER_TAB = iota
	ELB_LISTENER_TAB
	ELB_RULE_TAB
)

// LOADING_MSG
const (
	FETCHING_LOADBALANCERS_MSG = "fetching load balancers..."
	FETCHING_LISTENERS_MSG     = "fetching listeners..."
	FETCHING_RULES_MSG         = "fetching rules..."

	FILTERING_LOADBALANCERS_MSG = "searching load balancers..."
	FILTERING_LISTENERS_MSG     = "searching listeners..."
	FILTERING_RULES_MSG         = "searching rules..."
)
