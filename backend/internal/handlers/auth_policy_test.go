package handlers

import "testing"

func TestPasswordPolicyError(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantOK   bool
	}{
		{"good", "juan@example.com", "s3cretPhrase!", true},
		{"too short", "juan@example.com", "abc123", false},
		{"common password", "juan@example.com", "password123", false},
		{"common uppercase", "juan@example.com", "PASSWORD", false},
		{"contains email local", "juanito@example.com", "myjuanito99", false},
		{"short local not checked", "ab@example.com", "abzz1234", true},
		{"exactly 8 ok", "x@example.com", "abcdEF12", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := passwordPolicyError(tc.email, tc.password)
			if ok != tc.wantOK {
				t.Errorf("passwordPolicyError(%q,%q) ok=%v, want %v", tc.email, tc.password, ok, tc.wantOK)
			}
		})
	}
}
