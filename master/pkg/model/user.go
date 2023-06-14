package model

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/guregu/null.v3"

	"github.com/uptrace/bun"

	"github.com/determined-ai/determined/proto/pkg/userv1"
)

var (
	// EmptyPassword is the empty password (i.e., the empty string).
	EmptyPassword = null.NewString("", false)

	// NoPasswordLogin is a password that prevents the user from logging in
	// directly. They can still login via external authentication methods like
	// OAuth.
	NoPasswordLogin = null.NewString("", true)
)

// BCryptCost is a stopgap until we implement sane master-configuration.
const BCryptCost = 15

// UserID is the type for user IDs.
type UserID int

// SessionID is the type for user session IDs.
type SessionID int

// User corresponds to a row in the "users" DB table.
type User struct {
	bun.BaseModel `bun:"table:users"`
	ID            UserID      `db:"id" bun:"id,pk" json:"id"`
	Username      string      `db:"username" json:"username"`
	PasswordHash  null.String `db:"password_hash" json:"-"`
	DisplayName   null.String `db:"display_name" json:"display_name"`
	Admin         bool        `db:"admin" json:"admin"`
	Active        bool        `db:"active" json:"active"`
	ModifiedAt    time.Time   `db:"modified_at" json:"modified_at"`
	Remote        bool        `db:"remote" json:"remote"`
}

// UserSession corresponds to a row in the "user_sessions" DB table.
type UserSession struct {
	bun.BaseModel `bun:"table:user_sessions"`
	ID            SessionID `db:"id" json:"id"`
	UserID        UserID    `db:"user_id" json:"user_id"`
	Expiry        time.Time `db:"expiry" json:"expiry"`
}

// A FullUser is a User joined with any other user relations.
type FullUser struct {
	ID          UserID      `db:"id" json:"id"`
	DisplayName null.String `db:"display_name" json:"display_name"`
	Username    string      `db:"username" json:"username"`
	Name        string      `db:"name" json:"name"`
	Admin       bool        `db:"admin" json:"admin"`
	Active      bool        `db:"active" json:"active"`
	ModifiedAt  time.Time   `db:"modified_at" json:"modified_at"`
	Remote      bool        `db:"remote" json:"remote"`

	AgentUID   null.Int    `db:"agent_uid" json:"agent_uid"`
	AgentGID   null.Int    `db:"agent_gid" json:"agent_gid"`
	AgentUser  null.String `db:"agent_user" json:"agent_user"`
	AgentGroup null.String `db:"agent_group" json:"agent_group"`
}

// ToUser converts a FullUser model to just a User model.
func (u FullUser) ToUser() User {
	return User{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: null.String{},
		DisplayName:  u.DisplayName,
		Admin:        u.Admin,
		Active:       u.Active,
		ModifiedAt:   u.ModifiedAt,
		Remote:       u.Remote,
	}
}

// ValidatePassword checks that the supplied password is correct.
func (user User) ValidatePassword(password string) bool {
	// If an empty password was posted, we need to check that the
	// user is a password-less user.
	if password == "" {
		return !user.PasswordHash.Valid
	}

	// If the model's password is empty, then
	// supplied password must be incorrect
	if !user.PasswordHash.Valid {
		return false
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash.ValueOrZero()),
		[]byte(password))

	return err == nil
}

// UpdatePasswordHash updates the model's password hash employing necessary cryptographic
// techniques.
func (user *User) UpdatePasswordHash(password string) error {
	if password == "" {
		user.PasswordHash = EmptyPassword
	} else {
		passwordHash, err := HashPassword(password)
		if err != nil {
			return errors.Wrap(err, "error updating user password")
		}

		user.PasswordHash = null.StringFrom(passwordHash)
	}
	return nil
}

// Proto converts a user to its protobuf representation.
func (user *User) Proto() *userv1.User {
	return &userv1.User{
		Id:          int32(user.ID),
		Username:    user.Username,
		DisplayName: user.DisplayName.ValueOrZero(),
		Admin:       user.Admin,
		Active:      user.Active,
		ModifiedAt:  timestamppb.New(user.ModifiedAt),
		Remote:      user.Remote,
	}
}

// Users is a slice of User objects—primarily useful for its methods.
type Users []User

// Proto converts a slice of users to its protobuf representation.
func (users Users) Proto() []*userv1.User {
	out := make([]*userv1.User, len(users))
	for i, u := range users {
		out[i] = u.Proto()
	}
	return out
}

// ExternalSessions provides an integration point for an external service to issue JWTs to control
// access to the cluster.
type ExternalSessions struct {
	LoginURI        string    `json:"login_uri"`
	LogoutURI       string    `json:"logout_uri"`
	InvalidationURI string    `json:"invalidation_uri"`
	JwtKey          string    `json:"jwt_key"`
	OrgID           OrgID     `json:"org_id"`
	ClusterID       ClusterID `json:"cluster_id"`
	Invalidations   *InvalidationMap
}

// invalsLock synchronizes reading and updating ExternalSessions.Invalidations.
// OK to use a single lock since StartInvalidationPoll() is guarded by sync.Once.
var invalsLock sync.RWMutex

// Enabled returns whether or not external sessions are enabled.
func (e *ExternalSessions) Enabled() bool {
	return len(e.LoginURI) > 1
}

// Validate throws an error if the provided JWT is invalidated.
func (e *ExternalSessions) Validate(claims *JWT) error {
	invalsLock.RLock()
	defer invalsLock.RUnlock()
	if e.Invalidations == nil {
		return nil
	}
	d := time.Unix(claims.IssuedAt, 0)
	v := e.Invalidations.ValidFrom(claims.UserID)
	if d.Before(v) {
		return errors.New("token has been invalidated")
	}
	return nil
}

func (e *ExternalSessions) fetchInvalidations(cert *tls.Certificate) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{*cert},
				MinVersion:   tls.VersionTLS12,
			},
		},
	}

	req, err := http.NewRequest("GET", e.InvalidationURI, nil)
	if err != nil {
		log.WithError(err).Errorf("error fetching token invalidations")
		return
	}
	if e.Invalidations != nil {
		func() {
			invalsLock.RLock()
			defer invalsLock.RUnlock()
			req.Header.Set("If-Modified-Since", e.Invalidations.LastUpdated.Format(time.RFC1123))
		}()
	}

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Errorf("error fetching token invalidations")
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return
	}

	var im InvalidationMap
	err = json.NewDecoder(resp.Body).Decode(&im)
	if err != nil {
		log.WithError(err).Errorf("error parsing received token invalidations")
		return
	}

	func() {
		invalsLock.Lock()
		defer invalsLock.Unlock()
		e.Invalidations = &im
	}()
}

// StartInvalidationPoll polls for new invalidations every minute.
func (e *ExternalSessions) StartInvalidationPoll(cert *tls.Certificate) {
	t := time.NewTicker(1 * time.Minute)
	go func() {
		for range t.C {
			e.fetchInvalidations(cert)
		}
	}()
}

// InvalidationMap tracks times before which users should be considered invalid.
type InvalidationMap struct {
	DefaultTime time.Time            `json:"defaultTime"`
	LastUpdated time.Time            `json:"lastUpdated"`
	Overrides   map[string]time.Time `json:"overrides"`
}

// ValidFrom returns the time from which tokens for the specified user are valid.
func (im *InvalidationMap) ValidFrom(id string) time.Time {
	ts, ok := im.Overrides[id]
	if ok {
		return ts
	}
	return im.DefaultTime
}

// UserWebSetting is a record of user web setting.
type UserWebSetting struct {
	UserID      UserID
	Key         string
	Value       string
	StoragePath string
}

// HashPassword hashes the user's password.
func HashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		BCryptCost,
	)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}
