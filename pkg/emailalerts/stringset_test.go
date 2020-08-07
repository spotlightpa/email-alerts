package emailalerts

import (
	"strings"
	"testing"
)

func TestSymDiff(t *testing.T) {
	tests := map[string]struct {
		Old, New, Added, Removed string
	}{
		"empty":       {"", "", "", ""},
		"none":        {"a,b,c", "a,b,c", "", ""},
		"adding":      {"", "a", "a", ""},
		"removing":    {"a", "", "", "a"},
		"swap":        {"a", "b", "b", "a"},
		"some_remain": {"a,b,c", "b,c,d", "d", "a"},
		"doubles":     {"a,a,b,c,c", "b,b,c,d,d", "d", "a"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			old := strings.Split(tc.Old, ",")
			new := strings.Split(tc.New, ",")
			wantadded := strings.Split(tc.Added, ",")
			wantremoved := strings.Split(tc.Removed, ",")
			gotadded, gotremoved := symDiff(old, new)
			// quick and dirty test, can't rely on ordering!
			if strings.Join(wantadded, ",") != strings.Join(gotadded, ",") {
				t.Errorf("%v != %v", wantadded, gotadded)
			}
			if strings.Join(wantremoved, ",") != strings.Join(gotremoved, ",") {
				t.Errorf("%v != %v", wantremoved, gotremoved)
			}
		})
	}
}
