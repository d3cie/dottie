package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/d3cie/dottie/internal/auth"
	"github.com/d3cie/dottie/internal/infra/sqlite/db"
	"github.com/google/uuid"
)

var (
	ErrAlreadyConfigured = errors.New("dottie is already configured")
	ErrInvalidLogin      = errors.New("invalid email or password")
	ErrUnauthorized      = errors.New("authentication required")
	ErrNotFound          = errors.New("resource not found")
)

const sessionDuration = 30 * 24 * time.Hour

type Service struct {
	queries *db.Queries
	now     func() time.Time
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries, now: time.Now}
}

func (s *Service) SetupRequired(ctx context.Context) (bool, error) {
	count, err := s.queries.CountUsers(ctx)
	if err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return count == 0, nil
}

func (s *Service) Setup(ctx context.Context, email, password string) (User, string, error) {
	required, err := s.SetupRequired(ctx)
	if err != nil {
		return User{}, "", err
	}
	if !required {
		return User{}, "", ErrAlreadyConfigured
	}
	return s.createUserAndSession(ctx, email, password)
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	record, err := s.queries.GetUserByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !auth.VerifyPassword(record.PasswordHash, password)) {
		return User{}, "", ErrInvalidLogin
	}
	if err != nil {
		return User{}, "", fmt.Errorf("get user: %w", err)
	}
	if !auth.VerifyPassword(record.PasswordHash, password) {
		return User{}, "", ErrInvalidLogin
	}
	token, err := s.createSession(ctx, record.ID)
	if err != nil {
		return User{}, "", err
	}
	return userFromRecord(record), token, nil
}

func (s *Service) CurrentUser(ctx context.Context, token string) (User, error) {
	if token == "" {
		return User{}, ErrUnauthorized
	}
	record, err := s.queries.GetSessionUser(ctx, db.GetSessionUserParams{
		TokenHash: auth.HashToken(token),
		ExpiresAt: s.now().UTC().Format(time.RFC3339Nano),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnauthorized
	}
	if err != nil {
		return User{}, fmt.Errorf("get session: %w", err)
	}
	return userFromRecord(record), nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.queries.DeleteSession(ctx, auth.HashToken(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, email, password string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	if _, err := s.queries.GetUserByEmail(ctx, normalizeEmail(email)); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if err := s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{PasswordHash: hash, Email: normalizeEmail(email)}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (s *Service) ListWebsites(ctx context.Context, token string) ([]Website, error) {
	if _, err := s.CurrentUser(ctx, token); err != nil {
		return nil, err
	}
	records, err := s.queries.ListWebsites(ctx)
	if err != nil {
		return nil, fmt.Errorf("list websites: %w", err)
	}
	result := make([]Website, 0, len(records))
	for _, record := range records {
		result = append(result, websiteFromRecord(record))
	}
	return result, nil
}

func (s *Service) CreateWebsite(ctx context.Context, token, name, domainName, timezone string) (Website, error) {
	if _, err := s.CurrentUser(ctx, token); err != nil {
		return Website{}, err
	}
	domainName, err := normalizeDomain(domainName)
	if err != nil {
		return Website{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Website{}, errors.New("website name is required")
	}
	if timezone == "" {
		timezone = "UTC"
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return Website{}, errors.New("timezone is invalid")
	}
	record, err := s.queries.CreateWebsite(ctx, db.CreateWebsiteParams{
		ID: uuid.NewString(), Name: name, Domain: domainName, Timezone: timezone, CreatedAt: s.now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return Website{}, fmt.Errorf("create website: %w", err)
	}
	return websiteFromRecord(record), nil
}

func (s *Service) Website(ctx context.Context, id string) (Website, error) {
	record, err := s.queries.GetWebsite(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Website{}, ErrNotFound
	}
	if err != nil {
		return Website{}, fmt.Errorf("get website: %w", err)
	}
	return websiteFromRecord(record), nil
}

func (s *Service) createUserAndSession(ctx context.Context, email, password string) (User, string, error) {
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") {
		return User{}, "", errors.New("a valid email address is required")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return User{}, "", err
	}
	record, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		ID: uuid.NewString(), Email: email, PasswordHash: hash, CreatedAt: s.now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return User{}, "", fmt.Errorf("create user: %w", err)
	}
	token, err := s.createSession(ctx, record.ID)
	if err != nil {
		return User{}, "", err
	}
	return userFromRecord(record), token, nil
}

func (s *Service) createSession(ctx context.Context, userID string) (string, error) {
	token, tokenHash, err := auth.NewToken()
	if err != nil {
		return "", err
	}
	now := s.now().UTC()
	if err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		TokenHash: tokenHash, UserID: userID, ExpiresAt: now.Add(sessionDuration).Format(time.RFC3339Nano), CreatedAt: now.Format(time.RFC3339Nano),
	}); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeDomain(value string) (string, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Hostname() == "" {
		return "", errors.New("a valid website domain is required")
	}
	return parsed.Hostname(), nil
}

func userFromRecord(record db.User) User {
	return User{ID: record.ID, Email: record.Email}
}

func websiteFromRecord(record db.Website) Website {
	createdAt, _ := time.Parse(time.RFC3339Nano, record.CreatedAt)
	return Website{ID: record.ID, Name: record.Name, Domain: record.Domain, Timezone: record.Timezone, CreatedAt: createdAt}
}
