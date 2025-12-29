package db

import (
	"ATlas/models"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SQLiteConfig struct {
	Path string

	SessionExpiry    time.Duration
	InactivityExpiry time.Duration
	RequestExpiry    time.Duration
}

type SQLiteStore struct {
	db  *gorm.DB
	cfg *SQLiteConfig
}

var _ oauth.ClientAuthStore = &SQLiteStore{}

type storedSessionData struct {
	AccountDid syntax.DID              `gorm:"primaryKey"`
	SessionID  string                  `gorm:"primaryKey"`
	Data       oauth.ClientSessionData `gorm:"serializer:json"`
	CreatedAt  time.Time               `gorm:"index"`
	UpdatedAt  time.Time               `gorm:"index"`
}

type storedAuthRequest struct {
	State     string                `gorm:"primaryKey"`
	Data      oauth.AuthRequestData `gorm:"serializer:json"`
	CreatedAt time.Time             `gorm:"index"`
}

type storedPin struct {
	Did         string `gorm:"primaryKey"`
	Longitude   float64
	Latitude    float64
	Name        string
	Handle      string
	Description string
	Website     string
	Avatar      string
}

func NewSQLiteStore(cfg *SQLiteConfig) (*SQLiteStore, error) {
	if cfg == nil {
		return nil, fmt.Errorf("missing cfg")
	}
	if cfg.Path == "" {
		return nil, fmt.Errorf("missing DatabasePath")
	}
	if cfg.SessionExpiry == 0 {
		return nil, fmt.Errorf("missing SessionExpiryDuration")
	}
	if cfg.InactivityExpiry == 0 {
		return nil, fmt.Errorf("missing SessionInactivityDuration")
	}
	if cfg.RequestExpiry == 0 {
		return nil, fmt.Errorf("missing AuthRequestExpiryDuration")
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed opening db: %w", err)
	}

	db.AutoMigrate(&storedSessionData{})
	db.AutoMigrate(&storedAuthRequest{})
	db.AutoMigrate(&storedPin{})

	return &SQLiteStore{db, cfg}, nil
}

// Default page load behaviour
// TODO: add ability to load KNN for given lat/long
func (m *SQLiteStore) GetPins(ctx context.Context) ([]models.Pin, error) {
	var storedPins []storedPin
	res := m.db.WithContext(ctx).Limit(10000).Find(&storedPins) // Add ordering/increase limit/etc...

	pins := make([]models.Pin, len(storedPins))
	for i, sp := range storedPins {
		pins[i] = sp.toPin()
	}

	if res.Error != nil {
		return nil, res.Error
	}

	slog.Info("test", "r", storedPins)
	return pins, nil
}

func (m *SQLiteStore) SavePin(ctx context.Context, pin models.Pin) error {
	res := m.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&storedPin{
		Did:         pin.Did,
		Longitude:   pin.Longitude,
		Latitude:    pin.Latitude,
		Name:        pin.Name,
		Handle:      pin.Handle,
		Description: pin.Description,
		Website:     pin.Website,
		Avatar:      pin.Avatar,
	})

	slog.Info("pin persist", "err", res.Error)
	return res.Error
}

func (m *SQLiteStore) DeletePin(ctx context.Context, did string) error {
	res := m.db.WithContext(ctx).Where("did = ?", did).Delete(&storedPin{})

	slog.Info("pin delete", "err", res.Error)
	return res.Error
}

func (m *SQLiteStore) GetSession(ctx context.Context, did syntax.DID, sessionID string) (*oauth.ClientSessionData, error) {
	// Expire sessions that are expired or inactive
	expiry_threshold := time.Now().Add(-m.cfg.SessionExpiry)
	inactive_threshold := time.Now().Add(-m.cfg.InactivityExpiry)
	m.db.WithContext(ctx).Where(
		"created_at < ? OR updated_at < ?", expiry_threshold, inactive_threshold,
	).Delete(&storedSessionData{})

	var row storedSessionData
	res := m.db.WithContext(ctx).Where(&storedSessionData{
		AccountDid: did,
		SessionID:  sessionID,
	}).First(&row)
	if res.Error != nil {
		return nil, res.Error
	}
	return &row.Data, nil
}

func (m *SQLiteStore) SaveSession(ctx context.Context, sess oauth.ClientSessionData) error {
	res := m.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&storedSessionData{
		AccountDid: sess.AccountDID,
		SessionID:  sess.SessionID,
		Data:       sess,
	})
	return res.Error
}

func (m *SQLiteStore) DeleteSession(ctx context.Context, did syntax.DID, sessionID string) error {
	res := m.db.WithContext(ctx).Delete(&storedSessionData{
		AccountDid: did,
		SessionID:  sessionID,
	})
	return res.Error
}

func (m *SQLiteStore) GetAuthRequestInfo(ctx context.Context, state string) (*oauth.AuthRequestData, error) {
	// Delete expired auth requests
	threshold := time.Now().Add(-m.cfg.RequestExpiry)
	m.db.WithContext(ctx).Where("created_at < ?", threshold).Delete(&storedAuthRequest{})

	var row storedAuthRequest
	res := m.db.WithContext(ctx).Where(&storedAuthRequest{State: state}).First(&row)
	if res.Error != nil {
		return nil, res.Error
	}
	return &row.Data, nil
}

func (m *SQLiteStore) SaveAuthRequestInfo(ctx context.Context, info oauth.AuthRequestData) error {
	res := m.db.WithContext(ctx).Create(&storedAuthRequest{
		State: info.State,
		Data:  info,
	})
	return res.Error
}

func (m *SQLiteStore) DeleteAuthRequestInfo(ctx context.Context, state string) error {
	res := m.db.WithContext(ctx).Delete(&storedAuthRequest{State: state})
	return res.Error
}

func (sp storedPin) toPin() models.Pin {
	return models.Pin{
		Did:         sp.Did,
		Longitude:   sp.Longitude,
		Latitude:    sp.Latitude,
		Name:        sp.Name,
		Handle:      sp.Handle,
		Description: sp.Description,
		Website:     sp.Website,
		Avatar:      sp.Avatar,
	}
}
