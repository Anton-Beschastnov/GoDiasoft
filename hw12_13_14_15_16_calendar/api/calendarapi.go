package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	Day   ListEventsParamsPeriod = "day"
	Month ListEventsParamsPeriod = "month"
	Week  ListEventsParamsPeriod = "week"
)

type ListEventsParams struct {
	UserId string `form:"user_id" json:"user_id"`

	StartDate time.Time `form:"start_date" json:"start_date"`

	Period *ListEventsParamsPeriod `form:"period,omitempty" json:"period,omitempty"`
}

type ListEventsParamsPeriod string

type Event struct {
	Description  string    `json:"description"`
	EndTime      time.Time `json:"end_time"`
	ID           string    `json:"id"`
	NotifyBefore int64     `json:"notify_before"`
	StartTime    time.Time `json:"start_time"`
	Title        string    `json:"title"`
	UserID       string    `json:"user_id"`
}

type CreateEventRequest struct {
	Description  *string   `json:"description,omitempty"`
	EndTime      time.Time `json:"end_time"`
	NotifyBefore *int64    `json:"notify_before,omitempty"`
	StartTime    time.Time `json:"start_time"`
	Title        string    `json:"title"`
	UserID       string    `json:"user_id"`
}

type UpdateEventRequest struct {
	Description  *string    `json:"description,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	NotifyBefore *int64     `json:"notify_before,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	Title        *string    `json:"title,omitempty"`
}

type ServerInterface interface {
	ListEvents(w http.ResponseWriter, r *http.Request, params ListEventsParams)
	CreateEvent(w http.ResponseWriter, r *http.Request)
	DeleteEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)
	GetEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)
	UpdateEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)
}

type Unimplemented struct{}

func (_ Unimplemented) ListEvents(w http.ResponseWriter, r *http.Request, params ListEventsParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (_ Unimplemented) CreateEvent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (_ Unimplemented) DeleteEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (_ Unimplemented) GetEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (_ Unimplemented) UpdateEvent(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	w.WriteHeader(http.StatusNotImplemented)
}

type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

func (siw *ServerInterfaceWrapper) ListEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	var params ListEventsParams

	if paramValue := r.URL.Query().Get("user_id"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "user_id"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "user_id", r.URL.Query(), &params.UserId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "user_id", Err: err})
		return
	}

	if paramValue := r.URL.Query().Get("start_date"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "start_date"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "start_date", r.URL.Query(), &params.StartDate)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "start_date", Err: err})
		return
	}

	err = runtime.BindQueryParameter("form", true, false, "period", r.URL.Query(), &params.Period)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "period", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.ListEvents(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

func (siw *ServerInterfaceWrapper) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateEvent(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

func (siw *ServerInterfaceWrapper) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	var id openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeleteEvent(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

func (siw *ServerInterfaceWrapper) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	var id openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetEvent(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

func (siw *ServerInterfaceWrapper) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	var id openapi_types.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UpdateEvent(w, r, id)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/events", wrapper.ListEvents)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/events", wrapper.CreateEvent)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/events/{id}", wrapper.DeleteEvent)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/events/{id}", wrapper.GetEvent)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/events/{id}", wrapper.UpdateEvent)
	})

	return r
}

var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xY3W7jVBB+FevApdlmVyu05A7trlAlLtACV6haufFJ6yWx3ePjXVVVpPwgllUrIhAS",
	"V4AQL+DNxjRtEvcVZt4IzZy0+bGTNKUXsOpNlcT2nJn55vvmc49EJfCjuC4jUf5GOGFY8yqO9gJ/60UU",
	"+GLHFq6ser5HP0WifCSevpS+pg+hCkKptCf5Z1dGFeWFdFvuq4Df4QIG2IIExjCA1MIWZPAWj7ENA+wK",
	"W+jDUIqyiLTy/D3RsIX03efaq8uCYL9Agm1ILBhY0MMmpDDCrgUZnEMGY3xtTsFu/pRqoOqOFmXhOlp+",
	"xOELjvbcgkP/4szPIYEhnsAYj+HMggH0IYUxx//OXMY2ZNhccXQce27RqX6gverh811ZDVRR1b9CQlGp",
	"RsrgHDILxlxqhi1I4Rw7MIY+teQYW3gMQ0oGTyzsQA9S6EMGIxhSvgTBbEqerz9+OM3J87Xck4qSirSj",
	"9IY4jCFhDIYm4ZsgoD1dW9IDeEcNgAx6BPg1BimOpHpeBOn2EwsuJt085YBUTgpD7NoW9Ch/hneIJ/h9",
	"US0LRzVsoeRB7CnpEpUMylzIXB9nRnua3M5VtGD3haxojrYfRNSpWlBxavS5/Kj0qCRs4fnVIF/Os6df",
	"fmV9+sW2BX0qgdJNsQk9wzvLjC6j34cEm6YC02fx2KlJ33UUPS9s8VKqyAS9f690r0RdDELpO6EnykLY",
	"InT0PpN+S5IY8Mc9qQsA+xkyOOXhSPANJJBimxIzYrCIHzFqknuH0z3l8R5DRqAvBcui+xYeMfS84A4M",
	"IIO+4BIUK9u2K8ricy/ST032VJBy6lJLRSK4wZwwGKIsDmKpDoUtfIeIcoXq7EBoFUtbRJV9WXeoUfmZ",
	"X/hhOd1mCTZbIyTLCFaUpBlJumtlntckbEHyVSeumQfp1IVi/pimfQn6pMkdfG00akHA4YypE9eJXCbm",
	"Kym/FbaoB77eJwoVlRlK5QUExbSkxdRWB11X6w51LwoDPzKL8EGpVECEP5fO/FxqnpZ1jvKhklVRFh9s",
	"zWzfLbN6pyk4SjmHLBUPCw/9jdYCrSMW53NsMzFS6nRClIQRERKbeMwcggts8jJJOGYU1+uOOjRoXSIz",
	"MCultaoc7eyxmZiIw07DFiFLWUFXSB5Ij37CtsVMJ1lPc2rFl+ZtA6+weU4/VtLR0nQpR2oejt3AnZmN",
	"mfufyYNYRnolFd5bw3NnPd4v6/EvXEdeze4v4e0VDS3ssB6k+ANNhEmWWU1f5+RtlaptrmL9q12f4xcH",
	"+6TIkFwNC3bgb4rCjgG75Nst6GMTO/AOBjBaEBsYLYriVLxo1GfEK6dSOUFs2JfWaevIcxsmz5rUsvDF",
	"o88vHd1L8zTX+lbe+Iys7Sc5aXzC4ZdIY244c/PG4km2byqea/zN2vedtWv04caD1zGtYp3JzAysDwJj",
	"8yeBM/M2R4/OIz2BYLr81sBrb+aGbwToZ1L/r9AsXQeIWQw20o1bgbnA5cxFuIDMIFFkcGJd6ALeTnRh",
	"yl/s4BtIsYVt6GEHf6Rv663N16Hr/Kf5axc6rJm07xzWncO6ZYfVuI5/KW26RujaJWtvokW36mFuRdim",
	"MnTdDUYBQhW4cWX5/4W5JebyvtYh//TK2aPhKosH90qi8U8AAAD//xBLvuxcFgAA",
}

func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
