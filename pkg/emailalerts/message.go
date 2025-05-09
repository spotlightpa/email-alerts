package emailalerts

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"strings"
	"time"

	"github.com/spotlightpa/email-alerts/pkg/must"
)

const validityWindow time.Duration = 5 * time.Minute

type Message struct {
	Body      string
	CreatedAt time.Time
}

func (msg Message) ValidAt(t time.Time) bool {
	return msg.CreatedAt.Before(t) && msg.CreatedAt.Add(validityWindow).After(t)
}

func (msg Message) ValidNow() bool {
	return msg.ValidAt(time.Now())
}

func (msg *Message) Encode(obj any) error {
	var buf strings.Builder
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(obj); err != nil {
		return err
	}
	msg.Body = buf.String()
	return nil
}

func (msg Message) Decode(obj any) error {
	return gob.NewDecoder(strings.NewReader(msg.Body)).Decode(obj)
}

func (app *appEnv) signMessage(msg Message) string {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	must.Do(enc.Encode(msg))
	payloadGob := buf.Bytes()
	mac := hmac.New(sha256.New, []byte(app.signingSecret))
	mac.Write(payloadGob)
	rawSig := mac.Sum(nil)
	b := make([]byte, 0, base64.URLEncoding.EncodedLen(len(rawSig))+
		len(".")+
		base64.URLEncoding.EncodedLen(len(payloadGob)))
	b = base64.URLEncoding.AppendEncode(b, rawSig)
	b = append(b, '.')
	b = base64.URLEncoding.AppendEncode(b, payloadGob)
	return string(b)
}

func (app *appEnv) unpackMessage(signedMsg string) *Message {
	// Split on the dot
	b64Sig, b64Obj, ok := strings.Cut(signedMsg, ".")
	if !ok {
		return nil
	}
	encoding := base64.URLEncoding
	rawSig, err := encoding.DecodeString(b64Sig)
	if err != nil {
		return nil
	}
	gobObj, err := encoding.DecodeString(b64Obj)
	if err != nil {
		return nil
	}
	// Check that the signature matches the encoded gob
	mac := hmac.New(sha256.New, []byte(app.signingSecret))
	_ = must.Get(mac.Write(gobObj))
	expectedSig := mac.Sum(nil)
	if !hmac.Equal(rawSig, expectedSig) {
		return nil
	}
	// Decode the gob
	dec := gob.NewDecoder(bytes.NewBuffer(gobObj))
	var payload Message
	if err = dec.Decode(&payload); err != nil {
		// This should not happen because it passed the signing check
		app.logErr(context.Background(), err)
		return nil
	}
	return &payload
}
