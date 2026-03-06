package models

type Ticket struct {
	ID        	int 	`json:"id"`
	Total     	int 	`json:"total"`
	Available 	int 	`json:"available"`
}