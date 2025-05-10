package runtime

import (
	"encoding/json"
	"testing"
)

func TestConfigRead(t *testing.T) {
	cfg := NewConfig()
	err := cfg.Read("testdata/config.json")
	if err != nil {
		t.Fatal(err)
	}
	buf, err := json.MarshalIndent(cfg, "", "  ")
	t.Log(string(buf))
}
