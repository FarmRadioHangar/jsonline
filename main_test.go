package main

import "testing"

func TestConfig(t *testing.T) {

	//default config
	cfg := defaultConfig()
	if cfg.Out == nil {
		t.Error("expected default output to be os.Stdout")
	}
	tagSample := []struct {
		src    string
		expect string
		ok     bool
	}{
		{"not_a_tag", "", false},
	}
	for _, v := range tagSample {
		e, ok := cfg.IsTag(v.src)
		if e != v.expect {
			t.Errorf("expected %s got %s", v.expect, e)
		}
		if ok != v.ok {
			t.Errorf("expected %v got %v", v.ok, ok)
		}
	}
}
