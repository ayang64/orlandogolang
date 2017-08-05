package meetup

import (
	"testing"
)

func TestGetEvents(t *testing.T) {
	tests := []struct {
		Group      string
		ShouldFail bool
	}{
		{Group: "OrlandoGophers", ShouldFail: false},
		{Group: "OrlandoGolang", ShouldFail: false},
		{Group: "XXXXXXXXYYYYAAA-NO-SUCH-GROUP", ShouldFail: true},
	}

	for _, test := range tests {
		t.Run(test.Group, func(t *testing.T) {
			em, err := GetEvents(test.Group)

			if err != nil {
				if test.ShouldFail {
					t.Logf("GetEvents() returned error %s as expected", err)
					return
				} else {
					t.Errorf("GetEvents() returned error: %s", err)
				}
			}

			// we only bother to loop over the results if we're in verbose mode.
			if testing.Verbose() {
				t.Logf("Received the following events for %s.", test.Group)
				for i := range em {
					t.Logf("%v", em[i])
				}
			}
		})
	}
}
