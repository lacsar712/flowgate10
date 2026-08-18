package deadband

import "testing"

func TestAccept(t *testing.T) {
	if err := Enforce("n", "eu=100 deadband=1", []string{"n"}); err != nil {
		t.Fatal(err)
	}
}

func TestReject(t *testing.T) {
	if err := Enforce("n", "eu=100 deadband=20", []string{"n"}); err == nil {
		t.Fatal("expected reject")
	}
}
