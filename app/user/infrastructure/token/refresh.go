package token

// RefreshVerifier adapts a JWTIssuer into the application-level
// RefreshTokenVerifier port; the wrapper exists so the application
// service depends on a narrow interface (just VerifyRefresh) rather
// than the full issuer surface.
type RefreshVerifier struct {
	issuer *JWTIssuer
}

// NewRefreshVerifier returns a verifier backed by the supplied issuer.
func NewRefreshVerifier(issuer *JWTIssuer) *RefreshVerifier {
	return &RefreshVerifier{issuer: issuer}
}

// VerifyRefresh implements appUser.RefreshTokenVerifier.
func (v *RefreshVerifier) VerifyRefresh(token string) (int64, error) {
	if v == nil || v.issuer == nil {
		return 0, nil
	}
	return v.issuer.Verify(token, KindRefresh)
}
