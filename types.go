package fssync

import (
	"time"
	"github.com/google/uuid"
)

type Bounds struct {
	East  float64
	North float64
	South float64
	West  float64
}

func NewBounds() Bounds {
	return Bounds{
		East:  180,
		North: 180,
		South: -180,
		West:  -180,
	}
}

type Geometry struct {
	X float64 `db:"X"`
	Y float64 `db:"Y"`
}

type LatLong struct {
	Latitude  float64
	Longitude float64
}

type MosquitoInspection struct {
	Comments  string
	Condition string
	Created   time.Time
}

type MosquitoSource struct {
	AccessDescription string
	Comments          string
	Description       string
	Name              string
	Habitat           string
	Inspections       []MosquitoInspection
	UseType           string
	WaterOrigin       string
}

type Note struct {
	Category string
	Created  time.Time
	Content  string
	ID       uuid.UUID
	Location LatLong
}

type ServiceRequest struct {
	Geometry Geometry `db:"geometry"`
	Address  *string  `db:"reqaddr1"`
	City     *string  `db:"reqcity"`
	Priority *string  `db:"priority"`
	Source   *string  `db:"source"`
	Status   *string  `db:"status"`
	Target   *string  `db:"reqtarget"`
	Zip      *string  `db:"reqzip"`
}

type TrapData struct {
	Access      *string
	Comments    *string
	Condition   *string
	Data        []TrapData
	Description *string
	Geometry    Geometry
	End         *string
	FieldTech   *string
	Name        *string
	Species     *string
	Type        *string
}

type FS_TrapLocation struct {
	Access      *string  `db:"accessdesc"`
	Description *string  `db:"description"`
	Geometry    Geometry `db:"geometry"`
	GlobalID    *string  `db:"globalid"`
	ObjectID    int      `db:"objectid"`
	Name        *string  `db:"name"`
}

type User struct {
	DisplayName      string `db:"display_name"`
	PasswordHashType string `db:"password_hash_type"`
	PasswordHash     string `db:"password_hash"`
	Username         string `db:"username"`
}
