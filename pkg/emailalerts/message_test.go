package emailalerts_test

import (
	"reflect"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/email-alerts/pkg/emailalerts"
)

// TestMessageEncodeDecode tests the Encode and Decode methods of the Message struct
func TestMessageEncodeDecode(t *testing.T) {
	testCases := []any{
		"hello world",
		42,
		3.14159,
		true,
		[]int{1, 2, 3, 4, 5},
		map[string]int{"one": 1, "two": 2, "three": 3},
		struct {
			Name string
			Age  int
		}{"Alice", 30},
	}

	for _, tc := range testCases {
		tt := assert.Continues(t)
		var msg emailalerts.Message

		tt.NilError(msg.Encode(tc))

		tt.Truthy(msg.Body)

		objptrval := reflect.New(reflect.TypeOf(tc))
		tt.NilError(msg.Decode(objptrval.Interface()))

		tt.DeepEqual(tc, objptrval.Elem().Interface())
	}
}
