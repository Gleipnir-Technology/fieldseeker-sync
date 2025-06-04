package fssync

import (
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type hasCreated interface {
	getCreated() string
}

func parseTime(x string) time.Time {
	created_epoch, err := strconv.ParseInt(x, 10, 64)
	if err != nil {
		log.Println("Unable to convert inspection timestamp", x, err)
	}
	created := time.UnixMilli(created_epoch)
	return created
}

type FS_Geometry struct {
	X float64 `db:"X"`
	Y float64 `db:"Y"`
}

func (geo FS_Geometry) Latitude() float64 {
	return geo.Y
}
func (geo FS_Geometry) Longitude() float64 {
	return geo.X
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
	GlobalID        string  `db:"globalid"`
	PointLocationID string  `db:"pointlocid"`
}

type FS_PointLocation struct {
	Access       *string     `db:"accessdesc"`
	Comments     *string     `db:"comments"`
	CreationDate *int64      `db:"creationdate"`
	Description  *string     `db:"description"`
	Geometry     FS_Geometry `db:"geometry"`
	GlobalID     string      `db:"globalid"`
	Habitat      *string     `db:"habitat"`
	Inspections  MosquitoInspectionSlice
	Name         *string `db:"name"`
	Treatments   []MosquitoTreatment
	UseType      *string `db:"usetype"`
	WaterOrigin  *string `db:"waterorigin"`
}

type FS_ServiceRequest struct {
	Address      *string     `db:"reqaddr1"`
	CreationDate *int64      `db:"creationdate"`
	City         *string     `db:"reqcity"`
	Geometry     FS_Geometry `db:"geometry"`
	GlobalID     string      `db:"globalid"`
	Priority     *string     `db:"priority"`
	Source       *string     `db:"source"`
	Status       *string     `db:"status"`
	Target       *string     `db:"reqtarget"`
	Zip          *string     `db:"reqzip"`
}
type FS_TrapLocation struct {
	Access       *string     `db:"accessdesc"`
	CreationDate *int64      `db:"creationdate"`
	Description  *string     `db:"description"`
	Geometry     FS_Geometry `db:"geometry"`
	GlobalID     string      `db:"globalid"`
	ObjectID     int         `db:"objectid"`
	Name         *string     `db:"name"`
}

type FS_Treatment struct {
	Comments        *string  `db:"comments"`
	EndDateTime     string   `db:"enddatetime"`
	GlobalID        string   `db:"globalid"`
	Habitat         *string  `db:"habitat"`
	PointLocationID string   `db:"pointlocid"`
	Product         *string  `db:"product"`
	Quantity        float64  `db:"qty"`
	QuantityUnit    *string  `db:"qtyunit"`
	SiteCondition   *string  `db:"sitecond"`
	TreatAcres      *float64 `db:"treatacres"`
	TreatHectares   *float64 `db:"treathectares"`
}

type User struct {
	DisplayName      string `db:"display_name"`
	PasswordHashType string `db:"password_hash_type"`
	PasswordHash     string `db:"password_hash"`
	Username         string `db:"username"`
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
	data *FS_MosquitoInspection
}

func (mi MosquitoInspection) Comments() string {
	if mi.data.Comments == nil {
		return ""
	}
	return *mi.data.Comments
}

func (mi MosquitoInspection) Condition() string {
	if mi.data.Condition == nil {
		return ""
	}
	return *mi.data.Condition
}
func (mi MosquitoInspection) ID() string {
	return mi.data.GlobalID
}

func (mi MosquitoInspection) Created() time.Time {
	return parseTime(mi.data.EndDateTime)
}
func NewMosquitoInspections(inspections []*FS_MosquitoInspection) []MosquitoInspection {
	results := make([]MosquitoInspection, 0)
	for _, t := range inspections {
		results = append(results, MosquitoInspection{data: t})
	}
	MosquitoInspectionSlice(results).Sort()

	return results
}

type MosquitoInspectionSlice []MosquitoInspection
type ByCreatedMI []MosquitoInspection

func (a ByCreatedMI) Len() int           { return len(a) }
func (a ByCreatedMI) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedMI) Less(i, j int) bool { return a[i].Created().After(a[j].Created()) }

func (inspections MosquitoInspectionSlice) Sort() {
	sort.Sort(ByCreatedMI(inspections))
}

type MosquitoSource struct {
	location    *FS_PointLocation
	Inspections []MosquitoInspection
	Treatments  []MosquitoTreatment
}

func (s MosquitoSource) Access() string {
	if s.location.Access == nil {
		return ""
	}
	return *s.location.Access
}

func (s MosquitoSource) Comments() string {
	if s.location.Comments == nil {
		return ""
	}
	return *s.location.Comments
}

func (s MosquitoSource) Created() time.Time {
	if s.location.CreationDate == nil {
		return time.UnixMilli(0)
	}
	return time.UnixMilli(*s.location.CreationDate)
}

func (s MosquitoSource) Description() string {
	if s.location.Description == nil {
		return ""
	}
	return *s.location.Description
}

func (s MosquitoSource) Location() LatLong {
	return s.location.Geometry
}

func (s MosquitoSource) ID() uuid.UUID {
	return uuid.MustParse(s.location.GlobalID)
}
func (s MosquitoSource) Habitat() string {
	if s.location.Habitat == nil {
		return ""
	}
	return *s.location.Habitat
}

func (s MosquitoSource) Name() string {
	if s.location.Name == nil {
		return ""
	}
	return *s.location.Name
}

func (s MosquitoSource) UseType() string {
	if s.location.UseType == nil {
		return ""
	}
	return *s.location.UseType
}
func (s MosquitoSource) WaterOrigin() string {
	if s.location.WaterOrigin == nil {
		return ""
	}
	return *s.location.WaterOrigin
}
func NewMosquitoSource(location *FS_PointLocation, inspections []*FS_MosquitoInspection, treatments []*FS_Treatment) MosquitoSource {
	return MosquitoSource{
		location:    location,
		Inspections: NewMosquitoInspections(inspections),
		Treatments:  NewMosquitoTreatments(treatments),
	}
}

type MosquitoTreatment struct {
	data *FS_Treatment
}

func (t MosquitoTreatment) Comments() string {
	if t.data.Comments == nil {
		return ""
	}
	return *t.data.Comments
}
func (t MosquitoTreatment) Created() time.Time {
	return parseTime(t.data.EndDateTime)
}
func (mi MosquitoTreatment) ID() string {
	return mi.data.GlobalID
}
func (t MosquitoTreatment) Habitat() string {
	if t.data.Habitat == nil {
		return ""
	}
	return *t.data.Habitat
}
func (t MosquitoTreatment) Product() string {
	if t.data.Product == nil {
		return ""
	}
	return *t.data.Product
}
func (t MosquitoTreatment) Quantity() float64 {
	return t.data.Quantity
}
func (t MosquitoTreatment) QuantityUnit() string {
	if t.data.QuantityUnit == nil {
		return ""
	}
	return *t.data.QuantityUnit
}
func (t MosquitoTreatment) SiteCondition() string {
	if t.data.SiteCondition == nil {
		return ""
	}
	return *t.data.SiteCondition
}
func (t MosquitoTreatment) TreatAcres() float64 {
	if t.data.TreatAcres == nil {
		return 0
	}
	return *t.data.TreatAcres
}
func (t MosquitoTreatment) TreatHectares() float64 {
	if t.data.TreatHectares == nil {
		return 0
	}
	return *t.data.TreatHectares
}
func NewMosquitoTreatments(treatments []*FS_Treatment) []MosquitoTreatment {
	results := make([]MosquitoTreatment, 0)
	for _, t := range treatments {
		results = append(results, MosquitoTreatment{data: t})
	}
	MosquitoTreatmentSlice(results).Sort()
	return results
}

type MosquitoTreatmentSlice []MosquitoTreatment
type ByCreatedMT []MosquitoTreatment

func (a ByCreatedMT) Len() int           { return len(a) }
func (a ByCreatedMT) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedMT) Less(i, j int) bool { return a[i].Created().After(a[j].Created()) }

func (inspections MosquitoTreatmentSlice) Sort() {
	sort.Sort(ByCreatedMT(inspections))
}

type LatLong interface {
	Latitude() float64
	Longitude() float64
}

type ServiceRequest struct {
	data *FS_ServiceRequest
}

func (sr ServiceRequest) Address() string {
	if sr.data.Address == nil {
		return ""
	}
	return *sr.data.Address
}
func (sr ServiceRequest) City() string {
	if sr.data.City == nil {
		return ""
	}
	return *sr.data.City
}
func (sr ServiceRequest) Created() time.Time {
	if sr.data.CreationDate == nil {
		return time.UnixMilli(0)
	}
	return time.UnixMilli(*sr.data.CreationDate)
}
func (sr ServiceRequest) ID() uuid.UUID {
	return uuid.MustParse(sr.data.GlobalID)
}
func (sr ServiceRequest) Location() LatLong {
	return sr.data.Geometry
}
func (sr ServiceRequest) Priority() string {
	if sr.data.Priority == nil {
		return ""
	}
	return *sr.data.Priority
}
func (sr ServiceRequest) Status() string {
	if sr.data.Status == nil {
		return ""
	}
	return *sr.data.Status
}
func (sr ServiceRequest) Source() string {
	if sr.data.Source == nil {
		return ""
	}
	return *sr.data.Source
}
func (sr ServiceRequest) Target() string {
	if sr.data.Target == nil {
		return ""
	}
	return *sr.data.Target
}
func (sr ServiceRequest) UseType() string {
	return ""
}
func (sr ServiceRequest) WaterOrigin() string {
	return ""
}
func (sr ServiceRequest) Zip() string {
	if sr.data.Zip == nil {
		return ""
	}
	return *sr.data.Zip
}

type TrapData struct {
	data *FS_TrapLocation
}

func (tl TrapData) Access() string {
	if tl.data.Access == nil {
		return ""
	}
	return *tl.data.Access
}
func (tl TrapData) Created() time.Time {
	if tl.data.CreationDate == nil {
		return time.UnixMilli(0)
	}
	return time.UnixMilli(*tl.data.CreationDate)
}
func (tl TrapData) Description() string {
	if tl.data.Description == nil {
		return ""
	}
	return *tl.data.Description
}
func (tl TrapData) ID() uuid.UUID {
	return uuid.MustParse(tl.data.GlobalID)
}
func (tl TrapData) Location() LatLong {
	return tl.data.Geometry
}
func (tl TrapData) Name() string {
	if tl.data.Name == nil {
		return ""
	}
	return *tl.data.Name
}
