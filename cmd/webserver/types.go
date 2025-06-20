package main

import (
	"net/http"

	"github.com/go-chi/render"

	"gleipnir.technology/fieldseeker-sync"
)

// ResponseErr renderer type for handling all sorts of errors.
type ResponseClientIos struct {
	MosquitoSources []ResponseMosquitoSource `json:"sources"`
	ServiceRequests []ResponseServiceRequest `json:"requests"`
	TrapData        []ResponseTrapData       `json:"traps"`
}

func (i ResponseClientIos) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewResponseClientIos(sources []fssync.MosquitoSource, requests []fssync.ServiceRequest, traps []fssync.TrapData) ResponseClientIos {
	return ResponseClientIos{
		MosquitoSources: NewResponseMosquitoSources(sources),
		ServiceRequests: NewResponseServiceRequests(requests),
		TrapData:        NewResponseTrapData(traps),
	}
}

// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ResponseErr struct {
	Error          error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ResponseErr) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

type ResponseLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (rtd ResponseLocation) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewResponseLocation(l fssync.LatLong) ResponseLocation {
	return ResponseLocation{
		Latitude:  l.Latitude(),
		Longitude: l.Longitude(),
	}
}

type ResponseMosquitoInspection struct {
	ActionTaken   string `json:"action_taken"`
	Comments      string `json:"comments"`
	Condition     string `json:"condition"`
	Created       string `json:"created"`
	EndDateTime   string `json:"end_date_time"`
	FieldTech     string `json:"field_tech"`
	ID            string `json:"id"`
	LocationName  string `json:"location_name"`
	SiteCondition string `json:"site_condition"`
}

func (rtd ResponseMosquitoInspection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewResponseMosquitoInspection(i fssync.MosquitoInspection) ResponseMosquitoInspection {
	return ResponseMosquitoInspection{
		ActionTaken:   i.ActionTaken(),
		Comments:      i.Comments(),
		Condition:     i.Condition(),
		Created:       i.Created().Format("2006-01-02T15:04:05.000Z"),
		ID:            i.ID(),
		LocationName:  i.LocationName(),
		SiteCondition: i.SiteCondition(),
	}
}
func NewResponseMosquitoInspections(inspections []fssync.MosquitoInspection) []ResponseMosquitoInspection {
	results := make([]ResponseMosquitoInspection, 0)
	for _, i := range inspections {
		results = append(results, NewResponseMosquitoInspection(i))
	}
	return results
}

type ResponseMosquitoSource struct {
	Access                  string                       `json:"access"`
	Active                  *bool                        `json:"active"`
	Comments                string                       `json:"comments"`
	Created                 string                       `json:"created"`
	Description             string                       `json:"description"`
	ID                      string                       `json:"id"`
	LastInspectionDate      string                       `json:"last_inspection_date"`
	Location                ResponseLocation             `json:"location"`
	Habitat                 string                       `json:"habitat"`
	Inspections             []ResponseMosquitoInspection `json:"inspections"`
	Name                    string                       `json:"name"`
	NextActionDateScheduled string                       `json:"next_action_date_scheduled"`
	Treatments              []ResponseMosquitoTreatment  `json:"treatments"`
	UseType                 string                       `json:"use_type"`
	WaterOrigin             string                       `json:"water_origin"`
	Zone                    string                       `json:"zone"`
}

func (rtd ResponseMosquitoSource) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewResponseMosquitoSource(ms fssync.MosquitoSource) ResponseMosquitoSource {

	return ResponseMosquitoSource{
		Active:                  ms.Active(),
		Access:                  ms.Access(),
		Comments:                ms.Comments(),
		Created:                 ms.Created().Format("2006-01-02T15:04:05.000Z"),
		Description:             ms.Description(),
		ID:                      ms.ID().String(),
		LastInspectionDate:      ms.LastInspectionDate().Format("2006-01-02T15:04:05.000Z"),
		Location:                NewResponseLocation(ms.Location()),
		Habitat:                 ms.Habitat(),
		Inspections:             NewResponseMosquitoInspections(ms.Inspections),
		Name:                    ms.Name(),
		NextActionDateScheduled: ms.NextActionDateScheduled().Format("2006-01-02T15:04:05.000Z"),
		Treatments:              NewResponseMosquitoTreatments(ms.Treatments),
		UseType:                 ms.UseType(),
		WaterOrigin:             ms.WaterOrigin(),
		Zone:                    ms.Zone(),
	}
}
func NewResponseMosquitoSources(sources []fssync.MosquitoSource) []ResponseMosquitoSource {
	results := make([]ResponseMosquitoSource, 0)
	for _, i := range sources {
		results = append(results, NewResponseMosquitoSource(i))
	}
	return results
}

type ResponseMosquitoTreatment struct {
	Comments      string  `json:"comments"`
	Created       string  `json:"created"`
	EndDateTime   string  `json:"end_date_time"`
	FieldTech     string  `json:"field_tech"`
	Habitat       string  `json:"habitat"`
	ID            string  `json:"id"`
	Product       string  `json:"product"`
	Quantity      float64 `json:"quantity"`
	QuantityUnit  string  `json:"quantity_unit"`
	SiteCondition string  `json:"site_condition"`
	TreatAcres    float64 `json:"treat_acres"`
	TreatHectares float64 `json:"treat_hectares"`
}

func (rtd ResponseMosquitoTreatment) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewResponseMosquitoTreatment(i fssync.MosquitoTreatment) ResponseMosquitoTreatment {
	return ResponseMosquitoTreatment{
		Comments:      i.Comments(),
		Created:       i.Created().Format("2006-01-02T15:04:05.000Z"),
		FieldTech:     i.FieldTech(),
		Habitat:       i.Habitat(),
		ID:            i.ID(),
		Product:       i.Product(),
		Quantity:      i.Quantity(),
		QuantityUnit:  i.QuantityUnit(),
		SiteCondition: i.SiteCondition(),
		TreatAcres:    i.TreatAcres(),
		TreatHectares: i.TreatHectares(),
	}
}
func NewResponseMosquitoTreatments(treatments []fssync.MosquitoTreatment) []ResponseMosquitoTreatment {
	results := make([]ResponseMosquitoTreatment, 0)
	for _, i := range treatments {
		results = append(results, NewResponseMosquitoTreatment(i))
	}
	return results
}

type ResponseNote struct {
	CategoryName string `json:"categoryName"`
	Content      string `json:"content"`

	ID        string           `json:"id"`
	Location  ResponseLocation `json:"location"`
	Timestamp string           `json:"timestamp"`
}

func (rtd ResponseNote) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ResponseServiceRequest struct {
	Address           string           `json:"address"`
	AssignedTech      string           `json:"assigned_tech"`
	City              string           `json:"city"`
	Created           string           `json:"created"`
	HasDog            *bool            `json:"has_dog"`
	HasSpanishSpeaker *bool            `json:"has_spanish_speaker"`
	ID                string           `json:"id"`
	Location          ResponseLocation `json:"location"`
	Priority          string           `json:"priority"`
	RecordedDate      string           `json:"recorded_date"`
	Source            string           `json:"source"`
	Status            string           `json:"status"`
	Target            string           `json:"target"`
	Zip               string           `json:"zip"`
}

func (srr ResponseServiceRequest) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewResponseServiceRequest(sr fssync.ServiceRequest) ResponseServiceRequest {
	return ResponseServiceRequest{
		Address:           sr.Address(),
		AssignedTech:      sr.AssignedTech(),
		City:              sr.City(),
		Created:           sr.Created().Format("2006-01-02T15:04:05.000Z"),
		HasDog:            sr.HasDog(),
		HasSpanishSpeaker: sr.HasSpanishSpeaker(),
		ID:                sr.ID().String(),
		Location:          NewResponseLocation(sr.Location()),
		Priority:          sr.Priority(),
		Status:            sr.Status(),
		Source:            sr.Source(),
		Target:            sr.Target(),
		Zip:               sr.Zip(),
	}
}
func NewResponseServiceRequests(requests []fssync.ServiceRequest) []ResponseServiceRequest {
	results := make([]ResponseServiceRequest, 0)
	for _, i := range requests {
		results = append(results, NewResponseServiceRequest(i))
	}
	return results
}

type ResponseTrapData struct {
	Created     string           `json:"created"`
	Description string           `json:"description"`
	ID          string           `json:"id"`
	Location    ResponseLocation `json:"location"`
	Name        string           `json:"name"`
}

func (rtd ResponseTrapData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewResponseTrapDatum(td fssync.TrapData) ResponseTrapData {
	return ResponseTrapData{
		Created:     td.Created().Format("2006-01-02T15:04:05.000Z"),
		Description: td.Description(),
		ID:          td.ID().String(),
		Location:    NewResponseLocation(td.Location()),
		Name:        td.Name(),
	}
}
func NewResponseTrapData(data []fssync.TrapData) []ResponseTrapData {
	results := make([]ResponseTrapData, 0)
	for _, i := range data {
		results = append(results, NewResponseTrapDatum(i))
	}
	return results
}
