package tansultant

// Branch represents a dance studio branch.
type Branch struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	ScheduleURL string `json:"schedule_url"`
}

// Price represents a subscription pass with its price and description.
type Price struct {
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description"`
}
