package token

type Options func(*JWTTokenManager)

func WithIssuer(issuer string) Options {
	return func(tm *JWTTokenManager) {
		tm.Issuer = issuer
	}
}

func WithAudience(audiences ...string) Options {
	return func(tm *JWTTokenManager) {
		tm.Audiance = make(map[string]struct{})
		for _, audience := range audiences {
			tm.Audiance[audience] = struct{}{}
		}
	}
}
