package settings

import "testing"

type listenerFunc func() error

func (f listenerFunc) OnSettingsChanged() error { return f() }

func TestForceTLS12_DefaultFalse(t *testing.T) {
	svc := NewService(setupRepo(t))
	if svc.ForceTLS12() {
		t.Error("expected ForceTLS12 default false")
	}
}

func TestSetForceTLS12_PersistsAndNotifies(t *testing.T) {
	svc := NewService(setupRepo(t))
	var notified bool
	svc.SetChangeListener(listenerFunc(func() error { notified = true; return nil }))

	if err := svc.SetForceTLS12(true); err != nil {
		t.Fatalf("SetForceTLS12: %v", err)
	}
	if !svc.ForceTLS12() {
		t.Error("expected ForceTLS12 true after set")
	}
	if !notified {
		t.Error("expected change listener to be notified")
	}
}

func TestSetForceTLS12_NoListener(t *testing.T) {
	svc := NewService(setupRepo(t))
	if err := svc.SetForceTLS12(true); err != nil {
		t.Fatalf("SetForceTLS12 without listener: %v", err)
	}
	if !svc.ForceTLS12() {
		t.Error("expected ForceTLS12 true after set with no listener")
	}
}
