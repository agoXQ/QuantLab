package tests

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	appUser "github.com/agoXQ/QuantLab/app/user/application/user"
	domevent "github.com/agoXQ/QuantLab/app/user/domain/event"
	infraEvent "github.com/agoXQ/QuantLab/app/user/infrastructure/event"
	infraMemory "github.com/agoXQ/QuantLab/app/user/infrastructure/repository/memory"
)

// recordingPublisher captures every event the application service
// emits so tests can assert on the event flow without spinning up
// Kafka.
type recordingPublisher struct {
	mu     sync.Mutex
	events []domevent.Event
}

func newRecordingPublisher() *recordingPublisher { return &recordingPublisher{} }

func (p *recordingPublisher) Publish(_ context.Context, e domevent.Event) error {
	p.mu.Lock()
	p.events = append(p.events, e)
	p.mu.Unlock()
	return nil
}

func (p *recordingPublisher) Snapshot() []domevent.Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]domevent.Event, len(p.events))
	copy(out, p.events)
	return out
}

// fakeHasher is a deterministic hasher used by the unit tests so the
// password matching path stays fast and predictable. The hash is the
// SHA-256 hex of the plaintext; verify is a constant-time string
// compare.
type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:]), nil
}

func (fakeHasher) Verify(hash, plain string) error {
	expected, _ := fakeHasher{}.Hash(plain)
	if expected != hash {
		// Fall back to the platform error so the application service
		// path that translates ErrInvalidCredentials still runs.
		return userInvalidCredentials()
	}
	return nil
}

// fakeIssuer issues opaque tokens so we can assert the application
// service called the issuer without involving a real JWT library.
// The refresh-token format is "refresh-<userID>" so the matching
// verifier can decode it back without sharing state.
type fakeIssuer struct{}

func (fakeIssuer) Issue(userID int64) (appUser.TokenPair, error) {
	tag := strconv.FormatInt(userID, 10)
	return appUser.TokenPair{
		AccessToken:  "access-" + tag,
		RefreshToken: "refresh-" + tag,
		ExpiresIn:    3600,
	}, nil
}

// fakeRefreshVerifier decodes the deterministic refresh tokens minted
// above. Empty / malformed tokens map to the canonical
// ErrTokenInvalid so the application service surfaces the same error
// production sees.
type fakeRefreshVerifier struct{}

func (fakeRefreshVerifier) VerifyRefresh(token string) (int64, error) {
	const prefix = "refresh-"
	if !strings.HasPrefix(token, prefix) {
		return 0, userInvalidToken()
	}
	id, err := strconv.ParseInt(token[len(prefix):], 10, 64)
	if err != nil || id <= 0 {
		return 0, userInvalidToken()
	}
	return id, nil
}

// fixture wires the application service against the in-memory
// repositories with the deterministic hasher / issuer so the unit
// tests stay sandbox-friendly.
type fixture struct {
	svc       appUser.Service
	publisher *recordingPublisher
	clock     time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	clock := time.Date(2024, 6, 28, 12, 0, 0, 0, time.UTC)
	now := func() time.Time { return clock }
	publisher := newRecordingPublisher()
	svc := appUser.NewService(appUser.Dependencies{
		Users:           infraMemory.NewUserRepository(),
		Follows:         infraMemory.NewFollowRepository(),
		Hasher:          fakeHasher{},
		Tokens:          fakeIssuer{},
		RefreshVerifier: fakeRefreshVerifier{},
		Publisher:       publisher,
		Clock:           now,
	})
	return &fixture{svc: svc, publisher: publisher, clock: clock}
}

// silence: the Noop publisher is referenced from servicecontext.go,
// but we keep an import-side reference here so refactors that drop
// the package surface fail loudly during the tests.
var _ domevent.Publisher = infraEvent.Noop{}
