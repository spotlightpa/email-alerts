package emailalerts

import (
	"testing"
	"time"

	"github.com/carlmjohnson/be"
)

func TestVerifyToken(t *testing.T) {
	app := &appEnv{
		signingSecret: "test-secret-key",
	}

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		sinceSigning time.Duration
		token        string
		valid        bool
		setupFunc    func() string
	}{
		{
			name:         "valid token",
			sinceSigning: 1,
			valid:        true,
		},
		{
			name:         "expired token",
			sinceSigning: validityWindow + 1,
			valid:        false,
		},
		{
			name:         "future token beyond window",
			sinceSigning: -1,
			valid:        false,
		},
		{
			name:         "malformed token - no separator",
			sinceSigning: 1 * time.Minute,
			token:        "invalidtokenwithoutdot",
			valid:        false,
		},
		{
			name:         "malformed token - invalid base64",
			sinceSigning: 1 * time.Minute,
			token:        "invalid!base64.stillinvalid",
			valid:        false,
		},
		{
			name:         "wrong signature",
			sinceSigning: 1 * time.Minute,
			valid:        false,
			setupFunc: func() string {
				// Create token with different secret
				wrongApp := &appEnv{signingSecret: "wrong-secret"}
				return wrongApp.createToken(baseTime)
			},
		},
		{
			name:         "empty token",
			sinceSigning: 1 * time.Minute,
			setupFunc: func() string {
				return ""
			},
			valid: false,
		},
		{
			name:         "Hardcoded valid",
			sinceSigning: 1 * time.Minute,
			token:        "AQAAAA7dJKBAAAAAAP__.yooXIW4p5Tlh_U42Ft85BgFwN2MzkhsM2uG2rOL_tNM=",
			valid:        true,
		},
		{
			name:         "Hardcoded expired",
			sinceSigning: 15 * time.Minute,
			token:        "AQAAAA7dJKBAAAAAAP__.yooXIW4p5Tlh_U42Ft85BgFwN2MzkhsM2uG2rOL_tNM=",
			valid:        false,
		},
		{
			name:         "Hardcoded too soon",
			sinceSigning: -1 * time.Minute,
			token:        "AQAAAA7dJKBAAAAAAP__.yooXIW4p5Tlh_U42Ft85BgFwN2MzkhsM2uG2rOL_tNM=",
			valid:        false,
		},
		{
			name:         "Signature is decodable but wrong",
			sinceSigning: 1 * time.Minute,
			token:        "AQAAAA7dJKBAAAAAP__.yooXIW4p5Tlh_U42Ft85BgFwN2MzkhsM2uG2rOL_tNM=",
			valid:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.token
			if tt.setupFunc != nil {
				token = tt.setupFunc()
			} else if token == "" {
				token = app.createToken(baseTime)
			}
			now := baseTime.Add(tt.sinceSigning)
			got := app.verifyToken(now, token)
			be.Equal(t, tt.valid, got)
		})
	}
}
