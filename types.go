package fssync

import (
	"github.com/google/uuid"
	"time"
)

type FS_PointLocation struct {
	Geometry    Geometry `db:"geometry"`
	Access      *string  `db:"accessdesc"`
	Comments    *string  `db:"comments"`
	Description *string  `db:"description"`
	GlobalID    string   `db:"globalid"`
	Habitat     *string  `db:"habitat"`
	Name        *string  `db:"name"`
	UseType     *string  `db:"usetype"`
	WaterOrigin *string  `db:"waterorigin"`
}

type FS_MosquitoInspection struct {
	Comments        *string `db:"comments"`
	Condition       *string `db:"sitecond"`
	EndDateTime     string  `db:"enddatetime"`
	PointLocationID string  `db:"pointlocid"`
}

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

func (g Geometry) asLatLong() LatLong {
	return LatLong{
		Latitude:  g.X,
		Longitude: g.Y,
	}
}

type LatLong struct {
	Latitude  float64
	Longitude float64
}

type MosquitoInspection struct {
	Comments  *string
	Condition *string
	Created   time.Time
}
type ByCreated []MosquitoInspection

func (a ByCreated) Len() int           { return len(a) }
func (a ByCreated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreated) Less(i, j int) bool { return a[i].Created.After(a[j].Created) }

type MosquitoSource struct {
	Access      *string
	Comments    *string
	Description *string
	Location    LatLong
	Habitat     *string
	// ordered by created
	Inspections []MosquitoInspection
	Name        *string
	UseType     *string
	WaterOrigin *string
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
