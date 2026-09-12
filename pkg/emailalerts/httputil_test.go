package emailalerts

import (
	"testing"
	"time"

	"github.com/earthboundkid/assert"
)

func TestVerifyToken(t *testing.T) {
	app := &appEnv{
		signingSecret: "test-secret-key",
	}

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	type testcase struct {
		sinceSigning time.Duration
		token        string
		valid        bool
		setupFunc    func() string
	}
	assert.Run(t, map[string]testcase{
		"valid token": {
			sinceSigning: 1,
			valid:        true,
		},
		"expired token": {
			sinceSigning: validityWindow + 1,
			valid:        false,
		},
		"future token beyond window": {
			sinceSigning: -1,
			valid:        false,
		},
		"malformed token - no separator": {
			sinceSigning: 1 * time.Minute,
			token:        "invalidtokenwithoutdot",
			valid:        false,
		},
		"malformed token - invalid base64": {
			sinceSigning: 1 * time.Minute,
			token:        "invalid!base64.stillinvalid",
			valid:        false,
		},
		"wrong signature": {
			sinceSigning: 1 * time.Minute,
			valid:        false,
			setupFunc: func() string {
				// Create token with different secret
				wrongApp := &appEnv{signingSecret: "wrong-secret"}
				return wrongApp.createToken(baseTime)
			},
		},
		"empty token": {
			sinceSigning: 1 * time.Minute,
			setupFunc: func() string {
				return ""
			},
			valid: false,
		},
		"Hardcoded valid": {
			sinceSigning: 1 * time.Minute,
			token:        "B1-pzvvUsrttRA53aYqe7C4fIGYK-LKCNtra5RTVg2A=.LH8DAQEHTWVzc2FnZQH_gAABAgEEQm9keQEMAAEJQ3JlYXRlZEF0Af-CAAAAEP-BBQEBBFRpbWUB_4IAAAAU_4ACDwEAAAAO3SSgQAAAAAD__wA=",
			valid:        true,
		},
		"Hardcoded expired": {
			sinceSigning: 15 * time.Minute,
			token:        "B1-pzvvUsrttRA53aYqe7C4fIGYK-LKCNtra5RTVg2A=.LH8DAQEHTWVzc2FnZQH_gAABAgEEQm9keQEMAAEJQ3JlYXRlZEF0Af-CAAAAEP-BBQEBBFRpbWUB_4IAAAAU_4ACDwEAAAAO3SSgQAAAAAD__wA=",
			valid:        false,
		},
		"Hardcoded too soon": {
			sinceSigning: -1 * time.Minute,
			token:        "B1-pzvvUsrttRA53aYqe7C4fIGYK-LKCNtra5RTVg2A=.LH8DAQEHTWVzc2FnZQH_gAABAgEEQm9keQEMAAEJQ3JlYXRlZEF0Af-CAAAAEP-BBQEBBFRpbWUB_4IAAAAU_4ACDwEAAAAO3SSgQAAAAAD__wA=",
			valid:        false,
		},
	}, func(be assert.TB, tt testcase) {
		token := tt.token
		if tt.setupFunc != nil {
			token = tt.setupFunc()
		} else if token == "" {
			token = app.createToken(baseTime)
		}
		now := baseTime.Add(tt.sinceSigning)
		got := app.verifyToken(now, token)
		be.Equal(got, tt.valid)
	})

}
