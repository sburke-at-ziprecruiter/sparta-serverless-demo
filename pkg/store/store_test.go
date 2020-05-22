package store

import "testing"

func TestGet(t *testing.T) {
	phone := "828-555-1249"

	sto, err := Get(phone)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Store: %v", sto)
	if sto.Phone != phone {
		t.Errorf("Expecting phone %q, got %q", phone, sto.Phone)
	}
}
