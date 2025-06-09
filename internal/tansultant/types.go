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
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	Lifetime      int    `json:"lifetime"`
	Hours         int    `json:"hours"`
	FreezeAllowed int    `json:"freeze_allowed"`
	GuestVisits   int    `json:"guest_visits"`
}
