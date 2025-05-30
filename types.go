package fssync

import (
	"github.com/google/uuid"
	"time"
)

type FS_Geometry struct {
	X float64 `db:"X"`
	Y float64 `db:"Y"`
}

func (geo FS_Geometry) Latitude() float64 {
	return geo.X
}
func (geo FS_Geometry) Longitude() float64 {
	return geo.Y
}

type FS_InspectionSample struct {
	Geometry     FS_Geometry `db:"geometry"`
	CreationDate string      `db:"creationdate"`
	Creator      string      `db:"creator"`
	EditDate     string      `db:"editdate"`
	Editor       string      `db:"editor"`
	IDByTech     string      `db:"idbytech"`
	InspectionID string      `db:"insp_id"`
	Processed    int         `db:"processed"`
	SampleID     string      `db:"sampleid"`
}

type FS_MosquitoInspection struct {
	Comments        *string `db:"comments"`
	Condition       *string `db:"sitecond"`
	EndDateTime     string  `db:"enddatetime"`
	PointLocationID string  `db:"pointlocid"`
}

type FS_PointLocation struct {
	Geometry    FS_Geometry `db:"geometry"`
	Access      *string     `db:"accessdesc"`
	Comments    *string     `db:"comments"`
	Description *string     `db:"description"`
	GlobalID    string      `db:"globalid"`
	Habitat     *string     `db:"habitat"`
	Name        *string     `db:"name"`
	UseType     *string     `db:"usetype"`
	WaterOrigin *string     `db:"waterorigin"`
}

type FS_ServiceRequest struct {
	Geometry FS_Geometry `db:"geometry"`
	Address  *string     `db:"reqaddr1"`
	City     *string     `db:"reqcity"`
	Priority *string     `db:"priority"`
	Source   *string     `db:"source"`
	Status   *string     `db:"status"`
	Target   *string     `db:"reqtarget"`
	Zip      *string     `db:"reqzip"`
}

type FS_TrapLocation struct {
	Access      *string     `db:"accessdesc"`
	Description *string     `db:"description"`
	Geometry    FS_Geometry `db:"geometry"`
	GlobalID    *string     `db:"globalid"`
	ObjectID    int         `db:"objectid"`
	Name        *string     `db:"name"`
}

type FS_Treatment struct {
	Comments        *string  `db:"comments"`
	EndDateTime     string   `db:"enddatetime"`
	Habitat         *string  `db:"habitat"`
	PointLocationID string   `db:"pointlocid"`
	Product         *string  `db:"product"`
	Quantity        float64  `db:"qty"`
	QuantityUnit    *string  `db:"qtyunit"`
	SiteCondition   *string  `db:"sitecond"`
	TreatAcres      *float64 `db:"treatacres"`
	TreatHectares   *float64 `db:"treathectares"`
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

type MosquitoInspection struct {
	Comments  *string
	Condition *string
	Created   time.Time
}
type MosquitoInspectionByCreated []MosquitoInspection

func (a MosquitoInspectionByCreated) Len() int           { return len(a) }
func (a MosquitoInspectionByCreated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a MosquitoInspectionByCreated) Less(i, j int) bool { return a[i].Created.After(a[j].Created) }

type MosquitoSource struct {
	Access      *string
	Comments    *string
	Description *string
	Location    LatLong
	Habitat     *string
	// ordered by created
	Inspections []MosquitoInspection
	Name        *string
	// ordered by created
	Treatments  []MosquitoTreatment
	UseType     *string
	WaterOrigin *string
}

type MosquitoTreatment struct {
	Comments      *string
	Created       time.Time
	Habitat       *string
	Product       *string
	Quantity      float64
	QuantityUnit  *string
	SiteCondition *string
	TreatAcres    *float64
	TreatHectares *float64
}
type MosquitoTreatmentByCreated []MosquitoTreatment

func (a MosquitoTreatmentByCreated) Len() int           { return len(a) }
func (a MosquitoTreatmentByCreated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a MosquitoTreatmentByCreated) Less(i, j int) bool { return a[i].Created.After(a[j].Created) }

type Note struct {
	Category string
	Created  time.Time
	Content  string
	ID       uuid.UUID
	Location LatLong
}

type TrapData struct {
	Access      *string
	Comments    *string
	Condition   *string
	Data        []TrapData
	Description *string
	Geometry    FS_Geometry
	End         *string
	FieldTech   *string
	Name        *string
	Species     *string
	Type        *string
}

type User struct {
	DisplayName      string `db:"display_name"`
	PasswordHashType string `db:"password_hash_type"`
	PasswordHash     string `db:"password_hash"`
	Username         string `db:"username"`
}

type LatLong interface {
	Latitude() float64
	Longitude() float64
}

type ServiceRequest interface {
	LatLong() LatLong
	Address() string
	City() string
	Description() string
	ID() string
	Habitat() string
	Name() string
	UseType() string
	WaterOrigin() string
}
