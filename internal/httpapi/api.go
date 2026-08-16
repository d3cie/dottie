package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/d3cie/dottie/internal/domain"
	duckstore "github.com/d3cie/dottie/internal/infra/duckdb"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mileusna/useragent"
)

const sessionCookie = "dottie_session"

type Dependencies struct {
	Domain        *domain.Service
	Analytics     *duckstore.Client
	Logger        *slog.Logger
	SecureCookies bool
	Version       string
}

type Server struct {
	Router chi.Router
	API    huma.API
	deps   Dependencies
}

func New(deps Dependencies) *Server {
	router := chi.NewRouter()
	router.Use(recoverer(deps.Logger), requestLogger(deps.Logger), securityHeaders)
	router.Get("/health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	})
	router.Options("/api/v1/collect", collectorOptions)

	config := huma.DefaultConfig("Dottie API", deps.Version)
	config.Info.Description = "The local API used by the Dottie dashboard and browser tracker."
	config.DocsPath = "/api/docs"
	config.OpenAPIPath = "/openapi"
	api := humachi.New(router, config)
	server := &Server{Router: router, API: api, deps: deps}
	server.register()
	return server
}

func (s *Server) register() {
	huma.Register(s.API, operation(http.MethodGet, "/api/v1/system/bootstrap", "get-bootstrap-status", "System"), s.bootstrapStatus)
	huma.Register(s.API, operation(http.MethodPost, "/api/v1/auth/setup", "setup-admin", "Auth"), s.setup)
	huma.Register(s.API, operation(http.MethodPost, "/api/v1/auth/login", "login", "Auth"), s.login)
	huma.Register(s.API, operation(http.MethodPost, "/api/v1/auth/logout", "logout", "Auth"), s.logout)
	huma.Register(s.API, operation(http.MethodGet, "/api/v1/auth/me", "get-current-user", "Auth"), s.me)
	huma.Register(s.API, operation(http.MethodGet, "/api/v1/websites", "list-websites", "Websites"), s.listWebsites)
	huma.Register(s.API, operation(http.MethodPost, "/api/v1/websites", "create-website", "Websites"), s.createWebsite)
	huma.Register(s.API, operation(http.MethodGet, "/api/v1/websites/{website_id}/dashboard", "get-dashboard", "Analytics"), s.dashboard)
	huma.Register(s.API, operation(http.MethodGet, "/api/v1/websites/{website_id}/breakdowns", "get-breakdowns", "Analytics"), s.breakdowns)
	huma.Register(s.API, operation(http.MethodGet, "/api/v1/websites/{website_id}/visitors", "list-visitors", "Analytics"), s.visitors)
	huma.Register(s.API, operation(http.MethodPost, "/api/v1/collect", "collect-event", "Collector"), s.collect)
}

func operation(method, path, id, tag string) huma.Operation {
	return huma.Operation{Method: method, Path: path, OperationID: id, Tags: []string{tag}}
}

type bootstrapOutput struct {
	Body struct {
		SetupRequired bool `json:"setup_required"`
	}
}

func (s *Server) bootstrapStatus(ctx context.Context, _ *struct{}) (*bootstrapOutput, error) {
	required, err := s.deps.Domain.SetupRequired(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	response := &bootstrapOutput{}
	response.Body.SetupRequired = required
	return response, nil
}

type credentialsInput struct {
	Body struct {
		Email    string `json:"email" format:"email" maxLength:"320"`
		Password string `json:"password" minLength:"12" maxLength:"1024"`
	}
}

type authOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      domain.User
}

func (s *Server) setup(ctx context.Context, input *credentialsInput) (*authOutput, error) {
	user, token, err := s.deps.Domain.Setup(ctx, input.Body.Email, input.Body.Password)
	if errors.Is(err, domain.ErrAlreadyConfigured) {
		return nil, huma.Error409Conflict(err.Error())
	}
	if err != nil {
		return nil, badRequest(err)
	}
	return s.authResponse(user, token), nil
}

func (s *Server) login(ctx context.Context, input *credentialsInput) (*authOutput, error) {
	user, token, err := s.deps.Domain.Login(ctx, input.Body.Email, input.Body.Password)
	if errors.Is(err, domain.ErrInvalidLogin) {
		return nil, huma.Error401Unauthorized(err.Error())
	}
	if err != nil {
		return nil, internalError(err)
	}
	return s.authResponse(user, token), nil
}

type sessionInput struct {
	Session string `cookie:"dottie_session"`
}

type userOutput struct {
	Body domain.User
}

func (s *Server) me(ctx context.Context, input *sessionInput) (*userOutput, error) {
	user, err := s.deps.Domain.CurrentUser(ctx, input.Session)
	if err != nil {
		return nil, authError(err)
	}
	return &userOutput{Body: user}, nil
}

type logoutOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      struct {
		OK bool `json:"ok"`
	}
}

func (s *Server) logout(ctx context.Context, input *sessionInput) (*logoutOutput, error) {
	if err := s.deps.Domain.Logout(ctx, input.Session); err != nil {
		return nil, internalError(err)
	}
	response := &logoutOutput{SetCookie: s.expiredCookie()}
	response.Body.OK = true
	return response, nil
}

type websitesOutput struct {
	Body struct {
		Websites []domain.Website `json:"websites"`
	}
}

func (s *Server) listWebsites(ctx context.Context, input *sessionInput) (*websitesOutput, error) {
	websites, err := s.deps.Domain.ListWebsites(ctx, input.Session)
	if err != nil {
		return nil, authOrInternal(err)
	}
	response := &websitesOutput{}
	response.Body.Websites = websites
	return response, nil
}

type createWebsiteInput struct {
	Session string `cookie:"dottie_session"`
	Body    struct {
		Name     string `json:"name" minLength:"1" maxLength:"100"`
		Domain   string `json:"domain" minLength:"1" maxLength:"255"`
		Timezone string `json:"timezone" default:"UTC" maxLength:"100"`
	}
}

type websiteOutput struct {
	Body domain.Website
}

func (s *Server) createWebsite(ctx context.Context, input *createWebsiteInput) (*websiteOutput, error) {
	website, err := s.deps.Domain.CreateWebsite(ctx, input.Session, input.Body.Name, input.Body.Domain, input.Body.Timezone)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, authError(err)
		}
		return nil, badRequest(err)
	}
	return &websiteOutput{Body: website}, nil
}

type analyticsInput struct {
	Session   string `cookie:"dottie_session"`
	WebsiteID string `path:"website_id" format:"uuid"`
	Period    string `query:"period" enum:"7d,30d,90d" default:"30d"`
}

type dashboardOutput struct {
	Body domain.Dashboard
}

func (s *Server) dashboard(ctx context.Context, input *analyticsInput) (*dashboardOutput, error) {
	if _, err := s.deps.Domain.CurrentUser(ctx, input.Session); err != nil {
		return nil, authError(err)
	}
	if _, err := s.deps.Domain.Website(ctx, input.WebsiteID); err != nil {
		return nil, notFoundOrInternal(err)
	}
	result, err := s.deps.Analytics.Dashboard(ctx, input.WebsiteID, input.Period, time.Now())
	if err != nil {
		return nil, internalError(err)
	}
	return &dashboardOutput{Body: result}, nil
}

type breakdownInput struct {
	Session   string `cookie:"dottie_session"`
	WebsiteID string `path:"website_id" format:"uuid"`
	Period    string `query:"period" enum:"7d,30d,90d" default:"30d"`
	Dimension string `query:"dimension" enum:"pages,referrers,countries,cities,devices,browsers,os,events" default:"pages"`
	Search    string `query:"search" maxLength:"200"`
	Limit     int    `query:"limit" minimum:"1" maximum:"100" default:"9"`
	Offset    int    `query:"offset" minimum:"0" default:"0"`
}

type breakdownOutput struct {
	Body struct {
		Items []domain.BreakdownItem `json:"items"`
		Total int64                  `json:"total"`
	}
}

func (s *Server) breakdowns(ctx context.Context, input *breakdownInput) (*breakdownOutput, error) {
	if _, err := s.deps.Domain.CurrentUser(ctx, input.Session); err != nil {
		return nil, authError(err)
	}
	items, total, err := s.deps.Analytics.Breakdowns(ctx, input.WebsiteID, input.Period, input.Dimension, input.Search, input.Limit, input.Offset, time.Now())
	if err != nil {
		return nil, internalError(err)
	}
	response := &breakdownOutput{}
	response.Body.Items = items
	response.Body.Total = total
	return response, nil
}

type visitorsInput struct {
	Session   string `cookie:"dottie_session"`
	WebsiteID string `path:"website_id" format:"uuid"`
	Search    string `query:"search" maxLength:"200"`
	Limit     int    `query:"limit" minimum:"1" maximum:"100" default:"15"`
	Offset    int    `query:"offset" minimum:"0" default:"0"`
}

type visitorsOutput struct {
	Body struct {
		Visitors []domain.Visitor `json:"visitors"`
		Total    int64            `json:"total"`
	}
}

func (s *Server) visitors(ctx context.Context, input *visitorsInput) (*visitorsOutput, error) {
	if _, err := s.deps.Domain.CurrentUser(ctx, input.Session); err != nil {
		return nil, authError(err)
	}
	items, total, err := s.deps.Analytics.Visitors(ctx, input.WebsiteID, input.Search, input.Limit, input.Offset)
	if err != nil {
		return nil, internalError(err)
	}
	response := &visitorsOutput{}
	response.Body.Visitors = items
	response.Body.Total = total
	return response, nil
}

type collectInput struct {
	Origin    string `header:"Origin"`
	UserAgent string `header:"User-Agent"`
	DNT       string `header:"DNT"`
	Body      struct {
		WebsiteID string    `json:"website_id" format:"uuid"`
		VisitorID string    `json:"visitor_id" minLength:"8" maxLength:"100"`
		SessionID string    `json:"session_id" minLength:"8" maxLength:"100"`
		Name      string    `json:"name" enum:"pageview,event" default:"pageview"`
		Path      string    `json:"path" maxLength:"2048"`
		Hostname  string    `json:"hostname" maxLength:"255"`
		Title     string    `json:"title" maxLength:"500"`
		Referrer  string    `json:"referrer" maxLength:"2048"`
		Event     string    `json:"event,omitempty" maxLength:"100"`
		Timestamp time.Time `json:"timestamp,omitempty"`
	}
}

type collectOutput struct {
	AccessControlAllowOrigin string `header:"Access-Control-Allow-Origin"`
	Vary                     string `header:"Vary"`
	Body                     struct {
		Accepted bool `json:"accepted"`
	}
}

func (s *Server) collect(ctx context.Context, input *collectInput) (*collectOutput, error) {
	response := &collectOutput{AccessControlAllowOrigin: input.Origin, Vary: "Origin"}
	if input.DNT == "1" {
		response.Body.Accepted = false
		return response, nil
	}
	website, err := s.deps.Domain.Website(ctx, input.Body.WebsiteID)
	if err != nil {
		return nil, huma.Error404NotFound("Website not found")
	}
	if !allowedOrigin(website.Domain, input.Origin, input.Body.Hostname) {
		return nil, huma.Error403Forbidden("This origin is not allowed for the website")
	}
	ua := useragent.Parse(input.UserAgent)
	device := "desktop"
	if ua.Mobile {
		device = "mobile"
	} else if ua.Tablet {
		device = "tablet"
	}
	name := input.Body.Name
	if name == "event" && input.Body.Event != "" {
		name = input.Body.Event
	}
	timestamp := input.Body.Timestamp
	if timestamp.IsZero() || timestamp.Before(time.Now().Add(-24*time.Hour)) || timestamp.After(time.Now().Add(5*time.Minute)) {
		timestamp = time.Now()
	}
	err = s.deps.Analytics.InsertEvent(ctx, domain.Event{
		ID: uuid.NewString(), WebsiteID: website.ID, VisitorID: input.Body.VisitorID, SessionID: input.Body.SessionID,
		Name: name, Path: input.Body.Path, Hostname: input.Body.Hostname, Title: input.Body.Title, Referrer: input.Body.Referrer,
		Country: "unknown", City: "unknown", Device: device, Browser: valueOrUnknown(ua.Name), OS: valueOrUnknown(ua.OS), OccurredAt: timestamp,
	})
	if err != nil {
		return nil, internalError(err)
	}
	response.Body.Accepted = true
	return response, nil
}

func (s *Server) authResponse(user domain.User, token string) *authOutput {
	return &authOutput{Body: user, SetCookie: http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, Secure: s.deps.SecureCookies, SameSite: http.SameSiteLaxMode, MaxAge: int((30 * 24 * time.Hour).Seconds()),
	}}
}

func (s *Server) expiredCookie() http.Cookie {
	return http.Cookie{Name: sessionCookie, Path: "/", HttpOnly: true, Secure: s.deps.SecureCookies, SameSite: http.SameSiteLaxMode, MaxAge: -1}
}

func allowedOrigin(domainName, origin, hostname string) bool {
	if origin == "" {
		return strings.EqualFold(domainName, hostname)
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(domainName, parsed.Hostname()) && strings.EqualFold(domainName, hostname)
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func collectorOptions(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", request.Header.Get("Origin"))
	writer.Header().Set("Access-Control-Allow-Methods", http.MethodPost+", "+http.MethodOptions)
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, DNT")
	writer.Header().Set("Access-Control-Max-Age", "86400")
	writer.Header().Set("Vary", "Origin")
	writer.WriteHeader(http.StatusNoContent)
}

func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("request panic", "method", request.Method, "path", request.URL.Path, "error", fmt.Sprint(recovered))
					http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(writer, request)
		})
	}
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			started := time.Now()
			next.ServeHTTP(writer, request)
			if request.URL.Path != "/health" {
				logger.Debug("request completed", "method", request.Method, "path", request.URL.Path, "duration", time.Since(started))
			}
		})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-Frame-Options", "DENY")
		writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(writer, request)
	})
}

func badRequest(err error) error { return huma.Error400BadRequest(err.Error()) }
func internalError(err error) error {
	return huma.Error500InternalServerError("An unexpected error occurred", err)
}
func authError(err error) error {
	if errors.Is(err, domain.ErrUnauthorized) {
		return huma.Error401Unauthorized("Authentication required")
	}
	return internalError(err)
}
func authOrInternal(err error) error { return authError(err) }
func notFoundOrInternal(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return huma.Error404NotFound("Resource not found")
	}
	return internalError(err)
}
