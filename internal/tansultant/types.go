package tansultant

// Branch represents a dance studio branch.
// Only fields listed here are parsed from the API response.
type Branch struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	ScheduleLink string `json:"schedule_public_link"`
}

// Price represents a subscription pass with its price and description.
// Only fields listed here are parsed from the API response.
type Price struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         string `json:"price"`
	Lifetime      string `json:"lifetime"`
	Hours         string `json:"hours"`
	FreezeAllowed string `json:"freeze_allowed"`
	GuestVisits   string `json:"guest_visits"`
}
