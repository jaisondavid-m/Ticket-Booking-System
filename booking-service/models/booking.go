package models
type Booking struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateBookingRequest struct {
	Name string `json:"name"`
}