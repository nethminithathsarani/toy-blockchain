package node

import (
	"flag"
	"strings"
)

type Config struct {
	Address          string
	AdvertiseAddress string
	Peers            []string
}

func ParseConfig(args []string) (Config, error) {
	flags := flag.NewFlagSet("node", flag.ContinueOnError)
	address := flags.String(
		"address",
		"localhost:8001",
		"HTTP address used by this node",
	)
	advertiseAddress := flags.String(
		"advertise",
		"",
		"address advertised to other nodes",
	)

	peersText := flags.String(
		"peers",
		"",
		"comma-separated peer addresses",
	)
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	advertised := strings.TrimSpace(*advertiseAddress)

	if advertised == "" {
		advertised = *address
	}
	peers := []string{}

	for _, peer := range strings.Split(*peersText, ",") {
		peer = strings.TrimSpace(peer)

		if peer != "" {
			peers = append(peers, peer)
		}
	}
	return Config{
		Address:          *address,
		AdvertiseAddress: advertised,
		Peers:            peers,
	}, nil
}
