package db

type Point struct {
	Longitude float64 `db:"longitude"`
	Latitude  float64 `db:"latitude"`
}
