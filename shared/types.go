package shared

import (
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type H3Cell uint64

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
	ActionTaken     *string `db:"actiontaken"`
	Comments        *string `db:"comments"`
	Condition       *string `db:"sitecond"`
	EndDateTime     string  `db:"enddatetime"`
	FieldTech       *string `db:"fieldtech"`
	GlobalID        string  `db:"globalid"`
	LocationName    *string `db:"locationname"`
	PointLocationID string  `db:"pointlocid"`
	SiteCond        *string `db:"sitecond"`
	Zone            *string `db:"zone"`
}

type FS_PointLocation struct {
	Access                  *string     `db:"accessdesc"`
	Active                  *int        `db:"active"`
	Comments                *string     `db:"comments"`
	CreationDate            *int64      `db:"creationdate"`
	Description             *string     `db:"description"`
	Geometry                FS_Geometry `db:"geometry"`
	GlobalID                string      `db:"globalid"`
	Habitat                 *string     `db:"habitat"`
	Inspections             MosquitoInspectionSlice
	LastInspectDate         *int64  `db:"lastinspectdate"`
	Name                    *string `db:"name"`
	NextActionDateScheduled *int64  `db:"nextactiondatescheduled"`
	Treatments              []MosquitoTreatment
	UseType                 *string `db:"usetype"`
	WaterOrigin             *string `db:"waterorigin"`
	Zone                    *string `db:"zone"`
}

type FS_ServiceRequest struct {
	AssignedTech *string     `db:"assignedtech"`
	CreationDate *int64      `db:"creationdate"`
	City         *string     `db:"reqcity"`
	Dog          *int        `db:"dog"`
	Geometry     FS_Geometry `db:"geometry"`
	GlobalID     string      `db:"globalid"`
	Priority     *string     `db:"priority"`
	RecDateTime  *int64      `db:"recdatetime"`
	ReqAddr1     *string     `db:"reqaddr1"`
	ReqTarget    *string     `db:"reqtarget"`
	ReqZip       *string     `db:"reqzip"`
	Source       *string     `db:"source"`
	Spanish      *int        `db:"spanish"`
	Status       *string     `db:"status"`
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
	EndDateTime     *int64   `db:"enddatetime"`
	FieldTech       *string  `db:"fieldtech"`
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

func (mi MosquitoInspection) ActionTaken() string {
	if mi.data.ActionTaken == nil {
		return ""
	}
	return *mi.data.ActionTaken
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

func (mi MosquitoInspection) Created() time.Time {
	return parseTime(mi.data.EndDateTime)
}

func (mi MosquitoInspection) FieldTechnician() string {
	if mi.data.FieldTech == nil {
		return ""
	}
	return *mi.data.FieldTech
}

func (mi MosquitoInspection) ID() string {
	return mi.data.GlobalID
}

func (mi MosquitoInspection) LocationName() string {
	if mi.data.LocationName == nil {
		return ""
	}
	return *mi.data.LocationName
}

func (mi MosquitoInspection) SiteCondition() string {
	if mi.data.SiteCond == nil {
		return ""
	}
	return *mi.data.SiteCond
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

func (s MosquitoSource) Active() *bool {
	var result bool
	if s.location.Active == nil {
		return nil
	} else if *s.location.Active == 0 {
		result = false
	} else {
		result = true
	}
	return &result
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

func (s MosquitoSource) ID() uuid.UUID {
	return uuid.MustParse(s.location.GlobalID)
}
func (s MosquitoSource) Habitat() string {
	if s.location.Habitat == nil {
		return ""
	}
	return *s.location.Habitat
}

func (s MosquitoSource) LastInspectionDate() time.Time {
	if s.location.LastInspectDate == nil {
		return time.UnixMilli(0)
	}
	return time.UnixMilli(*s.location.LastInspectDate)
}

func (s MosquitoSource) Location() LatLong {
	return s.location.Geometry
}

func (s MosquitoSource) Name() string {
	if s.location.Name == nil {
		return ""
	}
	return *s.location.Name
}

func (s MosquitoSource) NextActionDateScheduled() time.Time {
	if s.location.NextActionDateScheduled == nil {
		return time.UnixMilli(0)
	}
	return time.UnixMilli(*s.location.NextActionDateScheduled)
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
func (s MosquitoSource) Zone() string {
	if s.location.Zone == nil {
		return ""
	}
	return *s.location.Zone
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
	if t.data.EndDateTime == nil {
		return time.UnixMilli(0)
	}
	return time.UnixMilli(*t.data.EndDateTime)
}
func (t MosquitoTreatment) FieldTechnician() string {
	if t.data.FieldTech == nil {
		return ""
	}
	return *t.data.FieldTech
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
	if sr.data.ReqAddr1 == nil {
		return ""
	}
	return *sr.data.ReqAddr1
}
func (sr ServiceRequest) AssignedTechnician() string {
	if sr.data.AssignedTech == nil {
		return ""
	}
	return *sr.data.AssignedTech
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
func (sr ServiceRequest) HasDog() *bool {
	var result bool
	if sr.data.Dog == nil {
		return nil
	} else if *sr.data.Dog == 0 {
		result = false
	} else {
		result = true
	}
	return &result
}
func (sr ServiceRequest) HasSpanishSpeaker() *bool {
	var result bool
	if sr.data.Spanish == nil {
		return nil
	} else if *sr.data.Spanish == 0 {
		result = false
	} else {
		result = true
	}
	return &result
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
func (sr ServiceRequest) RecDateTime() time.Time {
	if sr.data.RecDateTime == nil {
		return time.UnixMilli(0)
	}
	return time.UnixMilli(*sr.data.RecDateTime)
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
	if sr.data.ReqTarget == nil {
		return ""
	}
	return *sr.data.ReqTarget
}
func (sr ServiceRequest) UseType() string {
	return ""
}
func (sr ServiceRequest) WaterOrigin() string {
	return ""
}
func (sr ServiceRequest) Zip() string {
	if sr.data.ReqZip == nil {
		return ""
	}
	return *sr.data.ReqZip
}
func NewServiceRequest(data *FS_ServiceRequest) ServiceRequest {
	return ServiceRequest{data: data}
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
func NewTrapData(data *FS_TrapLocation) TrapData {
	return TrapData{data: data}
}

type Location struct {
	Latitude  float64
	Longitude float64
}

type NoteImagePayload struct {
	UUID    string    `json:"uuid"`
	Cell    H3Cell    `json:"cell"`
	Created time.Time `json:"created"`
}

type NoteAudioPayload struct {
	UUID          string                       `json:"uuid"`
	Breadcrumbs   []NoteAudioBreadcrumbPayload `json:"breadcrumbs"`
	Created       time.Time                    `json:"created"`
	Duration      int                          `json:"duration"`
	Transcription *string                      `json:"transcription"`
}

type NoteAudioBreadcrumbPayload struct {
	Cell    H3Cell    `json:"cell"`
	Created time.Time `json:"created"`
}

type NidusNotePayload struct {
	UUID      string    `json:"uuid"`
	Timestamp time.Time `json:"timestamp"`
	Images    []string  `json:"images"`
	Location  Location  `json:"location"`
	Text      string    `json:"text"`
}
