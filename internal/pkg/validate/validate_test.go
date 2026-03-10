package validate

import "testing"

func TestIP_Valid(t *testing.T) {
	cases := []string{"192.168.1.1", "10.0.0.1", "255.255.255.255", "0.0.0.0", "::1", "fe80::1"}
	for _, ip := range cases {
		if err := IP(ip); err != nil {
			t.Errorf("IP(%q) returned error: %v", ip, err)
		}
	}
}

func TestIP_Invalid(t *testing.T) {
	cases := []string{"", "999.999.999.999", "abc", "192.168.1", "192.168.1.1.1"}
	for _, ip := range cases {
		if err := IP(ip); err == nil {
			t.Errorf("IP(%q) expected error, got nil", ip)
		}
	}
}

func TestPort_Valid(t *testing.T) {
	cases := []int{1, 80, 443, 8080, 65535}
	for _, p := range cases {
		if err := Port(p); err != nil {
			t.Errorf("Port(%d) returned error: %v", p, err)
		}
	}
}

func TestPort_Invalid(t *testing.T) {
	cases := []int{0, -1, 65536, 100000}
	for _, p := range cases {
		if err := Port(p); err == nil {
			t.Errorf("Port(%d) expected error, got nil", p)
		}
	}
}

func TestDomain_Valid(t *testing.T) {
	cases := []string{"example.com", "sub.example.com", "a.b.c.example.org", "my-site.co.uk"}
	for _, d := range cases {
		if err := Domain(d); err != nil {
			t.Errorf("Domain(%q) returned error: %v", d, err)
		}
	}
}

func TestDomain_Invalid(t *testing.T) {
	cases := []string{"", "localhost", "-example.com", "example-.com", ".example.com", "example.c"}
	for _, d := range cases {
		if err := Domain(d); err == nil {
			t.Errorf("Domain(%q) expected error, got nil", d)
		}
	}
}

func TestRequired_NonEmpty(t *testing.T) {
	if err := Required("name", "hello"); err != nil {
		t.Errorf("Required with non-empty value returned error: %v", err)
	}
}

func TestRequired_Empty(t *testing.T) {
	err := Required("name", "")
	if err == nil {
		t.Fatal("Required with empty value expected error, got nil")
	}
	if got := err.Error(); got != "name is required" {
		t.Errorf("unexpected error message: %s", got)
	}
}
