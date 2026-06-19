// Package token also exposes a KeySet so the issuer can rotate signing
// secrets without invalidating live tokens. Each key carries a stable
// id (kid) so verify can pick the right secret per token, and the
// active key drives every freshly minted token. The MVP keeps the
// data structure in process memory; a future deployment can drop in a
// JWKS-backed loader behind the same shape.
package token

import (
	"errors"
	"fmt"
	"strings"
)

// MinSecretBytes is the platform-wide floor for HS256 signing keys.
// HS256 reduces to HMAC-SHA-256 internally, which mixes key material
// into a 256-bit state; anything shorter is detected as weak at boot
// so a misconfigured deployment fails closed instead of issuing
// guessable tokens.
const MinSecretBytes = 32

// ErrSecretTooShort is returned when a key fails the length floor.
var ErrSecretTooShort = errors.New("token: secret must be at least 32 bytes")

// ErrEmptyKeySet is returned when the issuer is wired without any
// keys at all. Production deployments must supply at least one.
var ErrEmptyKeySet = errors.New("token: at least one signing key required")

// SigningKey carries one HS256 secret + its stable id. The id ends up
// on every issued token's `kid` header so verify can pick the right
// secret per token; rotating a key amounts to flipping which entry
// is marked active.
type SigningKey struct {
	ID     string
	Secret string
}

// validate enforces the platform floor and trims surface area. The
// caller normalises the id to a non-empty trimmed string; an empty id
// becomes "default" so old configs without a kid still work.
func (k SigningKey) validate() (SigningKey, error) {
	id := strings.TrimSpace(k.ID)
	if id == "" {
		id = "default"
	}
	secret := k.Secret
	if len(secret) < MinSecretBytes {
		return SigningKey{}, fmt.Errorf("%w: kid=%s", ErrSecretTooShort, id)
	}
	return SigningKey{ID: id, Secret: secret}, nil
}

// KeySet wires the active signing key plus the read-only verification
// keys. The active key drives Issue; Verify walks every key that
// matches the token's kid header, falling back to the active key when
// the token carries no kid.
type KeySet struct {
	active   SigningKey
	verify   []SigningKey
	byID     map[string]string
}

// NewKeySet builds a KeySet from the supplied keys. The first key is
// the active signing key by default; pass activeID = "" to keep that
// behaviour, or set activeID to flip the active slot to a previously
// added key. Production deployments rotate by appending the new
// secret, advancing activeID, and removing the retired secret on a
// later deployment once every issued token has expired.
func NewKeySet(activeID string, keys []SigningKey) (*KeySet, error) {
	if len(keys) == 0 {
		return nil, ErrEmptyKeySet
	}
	cleaned := make([]SigningKey, 0, len(keys))
	byID := make(map[string]string, len(keys))
	for _, k := range keys {
		v, err := k.validate()
		if err != nil {
			return nil, err
		}
		if _, exists := byID[v.ID]; exists {
			return nil, fmt.Errorf("token: duplicate key id %q", v.ID)
		}
		byID[v.ID] = v.Secret
		cleaned = append(cleaned, v)
	}
	active := cleaned[0]
	if id := strings.TrimSpace(activeID); id != "" {
		secret, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("token: active key id %q not in key set", id)
		}
		active = SigningKey{ID: id, Secret: secret}
	}
	return &KeySet{active: active, verify: cleaned, byID: byID}, nil
}

// SingleKeySet is a convenience helper for the common "one signing
// key" case; it enforces the same length floor as NewKeySet.
func SingleKeySet(id, secret string) (*KeySet, error) {
	return NewKeySet(id, []SigningKey{{ID: id, Secret: secret}})
}

// Active returns the key new tokens are signed with.
func (s *KeySet) Active() SigningKey { return s.active }

// Lookup returns the secret for a given kid, falling back to the
// active key when kid is empty. ok = false signals "no key matches"
// so the caller can map it to an invalid-token error.
func (s *KeySet) Lookup(kid string) (secret string, id string, ok bool) {
	if s == nil {
		return "", "", false
	}
	if kid == "" {
		return s.active.Secret, s.active.ID, true
	}
	if secret, exists := s.byID[kid]; exists {
		return secret, kid, true
	}
	return "", "", false
}

// IDs returns every kid in the set. Useful for diagnostics; the
// returned slice is a copy so callers cannot mutate internal state.
func (s *KeySet) IDs() []string {
	if s == nil {
		return nil
	}
	out := make([]string, 0, len(s.verify))
	for _, k := range s.verify {
		out = append(out, k.ID)
	}
	return out
}
