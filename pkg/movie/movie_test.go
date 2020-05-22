package movie

import "testing"

func TestGet(t *testing.T) {
	mov, err := Get(2013, "Rush")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Movie: %v", mov)
	if mov.Year != 2013 {
		t.Errorf("Expecting year 2013, got %d", mov.Year)
	}
}
