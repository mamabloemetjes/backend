package health

import "testing"

func TestIsAuthorizedMonitoringRequest(t *testing.T) {
	const token = "12345678901234567890123456789012"
	tests := []struct {
		name          string
		authorization string
		want          bool
	}{
		{name: "valid bearer token", authorization: "Bearer " + token, want: true},
		{name: "wrong token", authorization: "Bearer wrong", want: false},
		{name: "missing scheme", authorization: token, want: false},
		{name: "empty configured token", authorization: "Bearer " + token, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configuredToken := token
			if test.name == "empty configured token" {
				configuredToken = ""
			}
			if got := isAuthorizedMonitoringRequest(test.authorization, configuredToken); got != test.want {
				t.Fatalf("authorization result = %t, want %t", got, test.want)
			}
		})
	}
}
