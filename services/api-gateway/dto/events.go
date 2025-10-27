package dto

type CreateEventRequest struct {
	Name        string `json:"name"`
	Detail      string `json:"detail"`
	Location    string `json:"location"`
	Date        string `json:"date"`
	JoiningCode string `json:"joining_code"`
	OrganizerId string `json:"organizer_id"`
}

type GetEventResponse struct {
	EventID     uint   `json:"event_id"`
	Name        string `json:"name"`
	Detail      string `json:"detail"`
	Location    string `json:"location"`
	Date        string `json:"date"`
	JoiningCode string `json:"joining_code"`
	OrganizerId string `json:"organizer_id"`
}

type JoinEventRequest struct {
	EventID     uint   `json:"event_id"`
	JoiningCode string `json:"joining_code"`
}
