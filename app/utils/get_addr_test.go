package utils

import "testing"

func TestGetAddr(t *testing.T) {
	addr := GetAddr("127.0.0.1", 3000)
	if addr != "127.0.0.1:3000" {
		t.Error("The result is inconsistent")
	}
}
