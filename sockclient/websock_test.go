package sockclient

import (
	"fmt"
	"testing"
)

func TestAuthorize(t *testing.T) {
	client, err := NewSockClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(client.Buffer))

	const (
		auth      = `{ "type":"auth", "access_token":"%s" }`
		config    = `{ "type":"get_config", "id":%d }`
		states    = `{ "type":"get_states", "id":%d }`
		subscribe = `{ "type":"subscribe_events", "event_type":"state_changed", "id":%d }`
	)

	// authorize
	cmd := fmt.Sprintf(auth, Token)

	err = client.WriteCommand(cmd)
	if err != nil {
		t.Fatal(err)
	}

	var buf []byte
	buf, err = client.Read()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("FIRST TIME", string(buf))

	err = client.WriteCommand(cmd)
	if err != nil {
		t.Fatal(err)
	}
	buf, err = client.Read()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("second TIME", string(buf))
}
