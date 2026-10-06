package node

import "testing"

func TestParseConfig(t *testing.T) {
	args := []string{
		"-address", "localhost:8002",
		"-peers", "localhost:8001, localhost:8003",
	}

	config, err := ParseConfig(args)

	if err != nil {
		t.Fatalf("expected configuration to parse: %v", err)
	}
	if config.Address != "localhost:8002" {
		t.Fatalf(
			"expected address localhost:8002, got %s",
			config.Address,
		)

	}
	if config.AdvertiseAddress != "localhost:8002" {
		t.Fatalf(
			"expected advertised address to fall back to localhost:8002, got %s",
			config.AdvertiseAddress,
		)
	}
	if len(config.Peers) != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			len(config.Peers),
		)
	}

	if config.Peers[0] != "localhost:8001" {
		t.Errorf("unexpected first peer: %s", config.Peers[0])
	}

	if config.Peers[1] != "localhost:8003" {
		t.Errorf("unexpected second peer: %s", config.Peers[1])
	}
}
func TestParseConfigWithAdvertiseAddress(t *testing.T) {
	args := []string{
		"-address", "0.0.0.0:8001",
		"-advertise", "node-a:8001",
		"-peers", "node-b:8002",
	}

	config, err := ParseConfig(args)

	if err != nil {
		t.Fatalf(
			"expected configuration to parse: %v",
			err,
		)
	}

	if config.Address != "0.0.0.0:8001" {
		t.Fatalf(
			"expected listen address 0.0.0.0:8001, got %s",
			config.Address,
		)
	}

	if config.AdvertiseAddress != "node-a:8001" {
		t.Fatalf(
			"expected advertised address node-a:8001, got %s",
			config.AdvertiseAddress,
		)
	}

	if len(config.Peers) != 1 ||
		config.Peers[0] != "node-b:8002" {

		t.Fatalf(
			"expected peer node-b:8002, got %v",
			config.Peers,
		)
	}
}
