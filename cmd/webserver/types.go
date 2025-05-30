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
func NewResponseClientIos(sources []*fssync.MosquitoSource, requests []*fssync.ServiceRequest, traps []*fssync.TrapData) ResponseClientIos {
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
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
	}
}

type ResponseMosquitoInspection struct {
	Comments  *string `json:"comments"`
	Condition *string `json:"condition"`
	Created   string  `json:"created"`
}

func (rtd ResponseMosquitoInspection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewResponseMosquitoInspection(i fssync.MosquitoInspection) ResponseMosquitoInspection {
	return ResponseMosquitoInspection{
		Comments:  i.Comments,
		Condition: i.Condition,
		Created:   i.Created.String(),
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
	Access      *string                      `json:"access"`
	Comments    *string                      `json:"comments"`
	Description *string                      `json:"description"`
	Location    ResponseLocation             `json:"location"`
	Habitat     *string                      `json:"habitat"`
	Inspections []ResponseMosquitoInspection `json:"inspections"`
	Name        *string                      `json:"name"`
	Treatments  []ResponseMosquitoTreatment  `json:"treatments"`
	UseType     *string                      `json:"usetype"`
	WaterOrigin *string                      `json:"waterorigin"`
}

func (rtd ResponseMosquitoSource) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewResponseMosquitoSource(ms *fssync.MosquitoSource) ResponseMosquitoSource {

	return ResponseMosquitoSource{
		Access:      ms.Access,
		Comments:    ms.Comments,
		Description: ms.Description,
		Location:    NewResponseLocation(ms.Location),
		Habitat:     ms.Habitat,
		Inspections: NewResponseMosquitoInspections(ms.Inspections),
		Name:        ms.Name,
		Treatments:  NewResponseMosquitoTreatments(ms.Treatments),
		UseType:     ms.UseType,
		WaterOrigin: ms.WaterOrigin,
	}
}
func NewResponseMosquitoSources(sources []*fssync.MosquitoSource) []ResponseMosquitoSource {
	results := make([]ResponseMosquitoSource, 0)
	for _, i := range sources {
		results = append(results, NewResponseMosquitoSource(i))
	}
	return results
}

type ResponseMosquitoTreatment struct {
	Comments      *string  `json:"comments"`
	Created       string   `json:"created"`
	Habitat       *string  `json:"habitat"`
	Product       *string  `json:"product"`
	Quantity      float64  `json:"quantity"`
	QuantityUnit  *string  `json:"quantity_unit"`
	SiteCondition *string  `json:"site_condition"`
	TreatAcres    *float64 `json:"treat_acres"`
	TreatHectares *float64 `json:"treat_hectares"`
}

func (rtd ResponseMosquitoTreatment) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewResponseMosquitoTreatment(i fssync.MosquitoTreatment) ResponseMosquitoTreatment {
	return ResponseMosquitoTreatment{
		Comments:      i.Comments,
		Created:       i.Created.String(),
		Habitat:       i.Habitat,
		Product:       i.Product,
		Quantity:      i.Quantity,
		QuantityUnit:  i.QuantityUnit,
		SiteCondition: i.SiteCondition,
		TreatAcres:    i.TreatAcres,
		TreatHectares: i.TreatHectares,
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
	CategoryName string           `json:"categoryName"`
	Content      string           `json:"content"`
	ID           string           `json:"id"`
	Location     ResponseLocation `json:"location"`
	Timestamp    string           `json:"timestamp"`
}

func (rtd ResponseNote) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ResponseServiceRequest struct {
	Address  *string          `json:"address"`
	City     *string          `json:"city"`
	Location ResponseLocation `json:"location"`
	Priority *string          `json:"priority"`
	Source   *string          `json:"source"`
	Status   *string          `json:"status"`
	Target   *string          `json:"target"`
	Zip      *string          `json:"zip"`
}

func (srr ResponseServiceRequest) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewResponseServiceRequest(sr *fssync.ServiceRequest) ResponseServiceRequest {
	return ResponseServiceRequest{
		Address:  sr.Address,
		City:     sr.City,
		Location: NewResponseLocation(sr.Geometry),
		Priority: sr.Priority,
		Status:   sr.Status,
		Source:   sr.Source,
		Target:   sr.Target,
		Zip:      sr.Zip,
	}
}
func NewResponseServiceRequests(requests []*fssync.ServiceRequest) []ResponseServiceRequest {
	results := make([]ResponseServiceRequest, 0)
	for _, i := range requests {
		results = append(results, NewResponseServiceRequest(i))
	}
	return results
}

type ResponseTrapData struct {
	Description *string          `json:"description"`
	Location    ResponseLocation `json:"location"`
	Name        *string          `json:"name"`
}

func (rtd ResponseTrapData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewResponseTrapDatum(td *fssync.TrapData) ResponseTrapData {
	return ResponseTrapData{
		Description: td.Description,
		Lat:         td.Geometry.Y,
		Long:        td.Geometry.X,
		Name:        td.Name,
	}
}
func NewResponseTrapData(data []*fssync.TrapData) []ResponseTrapData {
	results := make([]ResponseTrapData, 0)
	for _, i := range data {
		results = append(results, NewResponseTrapDatum(i))
	}
	return results
}

func NewNote(n fssync.Note) ResponseNote {
	return ResponseNote{
		CategoryName: n.Category,
		Content:      n.Content,
		ID:           n.ID.String(),
		Location:     NewResponseLocation(n.Location),
		Timestamp:    n.Created.Format("2006-01-02T15:04:05.000Z"),
	}
}
