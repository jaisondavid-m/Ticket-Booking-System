package models

type Ticket struct {
	ID        int `json:"id"`
	Total     int `json:"total"`
	Available int `json:"available"`
}

type Booking struct {
	ID   	int 	`json:"id"`
	Name 	string	`json:"name"`
}

type BookRequest struct {
	Name string `json:"name"`
}