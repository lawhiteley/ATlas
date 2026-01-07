package models

type Pin struct {
	Did         string
	Uri         string
	PlacedAt    string
	Longitude   float64 `validate:"longitude"`
	Latitude    float64 `validate:"latitude"`
	Name        string
	Handle      string
	Description string `validate:"min=1,max=280"`
	Website     string `validate:"omitempty,url"`
	Avatar      string
}
