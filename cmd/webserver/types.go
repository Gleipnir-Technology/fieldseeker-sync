package main

import (
	"net/http"

	"github.com/go-chi/render"

	"gleipnir.technology/fieldseeker-sync"
)

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Error          error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
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

type ResponseMosquitoInspection struct {
	Comments  *string `json:"comments"`
	Condition *string `json:"condition"`
	Created   string  `json:"created"`
}

func (rtd ResponseMosquitoInspection) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
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
	Address  *string `json:"address"`
	City     *string `json:"city"`
	Lat      float64 `json:"lat"`
	Long     float64 `json:"long"`
	Priority *string `json:"priority"`
	Source   *string `json:"source"`
	Status   *string `json:"status"`
	Target   *string `json:"target"`
	Zip      *string `json:"zip"`
}

func (srr ResponseServiceRequest) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewServiceRequest(sr *fssync.ServiceRequest) ResponseServiceRequest {
	return ResponseServiceRequest{
		Address:  sr.Address,
		City:     sr.City,
		Lat:      sr.Geometry.Y,
		Long:     sr.Geometry.X,
		Priority: sr.Priority,
		Status:   sr.Status,
		Source:   sr.Source,
		Target:   sr.Target,
		Zip:      sr.Zip,
	}
}

type ResponseTrapData struct {
	Description *string `json:"description"`
	Lat         float64 `json:"lat"`
	Long        float64 `json:"long"`
	Name        *string `json:"name"`
}

func (rtd ResponseTrapData) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewTrapData(td *fssync.TrapData) ResponseTrapData {
	return ResponseTrapData{
		Description: td.Description,
		Lat:         td.Geometry.Y,
		Long:        td.Geometry.X,
		Name:        td.Name,
	}
}

func NewLocation(l fssync.LatLong) ResponseLocation {
	return ResponseLocation{
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
	}
}

func NewMosquitoInspection(i fssync.MosquitoInspection) ResponseMosquitoInspection {
	return ResponseMosquitoInspection{
		Comments:  i.Comments,
		Condition: i.Condition,
		Created:   i.Created.String(),
	}
}
func NewMosquitoInspections(inspections []fssync.MosquitoInspection) []ResponseMosquitoInspection {
	results := make([]ResponseMosquitoInspection, 0)
	for _, i := range inspections {
		results = append(results, NewMosquitoInspection(i))
	}
	return results
}

func NewMosquitoSource(ms *fssync.MosquitoSource) ResponseMosquitoSource {

	return ResponseMosquitoSource{
		Access:      ms.Access,
		Comments:    ms.Comments,
		Description: ms.Description,
		Location:    NewLocation(ms.Location),
		Habitat:     ms.Habitat,
		Inspections: NewMosquitoInspections(ms.Inspections),
		Name:        ms.Name,
		Treatments:  NewMosquitoTreatments(ms.Treatments),
		UseType:     ms.UseType,
		WaterOrigin: ms.WaterOrigin,
	}
}

func NewMosquitoTreatment(i fssync.MosquitoTreatment) ResponseMosquitoTreatment {
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
func NewMosquitoTreatments(treatments []fssync.MosquitoTreatment) []ResponseMosquitoTreatment {
	results := make([]ResponseMosquitoTreatment, 0)
	for _, i := range treatments {
		results = append(results, NewMosquitoTreatment(i))
	}
	return results
}

func NewNote(n fssync.Note) ResponseNote {
	return ResponseNote{
		CategoryName: n.Category,
		Content:      n.Content,
		ID:           n.ID.String(),
		Location:     NewLocation(n.Location),
		Timestamp:    n.Created.Format("2006-01-02T15:04:05.000Z"),
	}
}
