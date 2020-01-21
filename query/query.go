package query

import (
	"context"
	"fmt"
	"net"
	"strings"

	"go.txtdirect.org/txtdirect/config"
	"go.txtdirect.org/txtdirect/variables"
)

// query checks the given zone using net.LookupTXT to
// find TXT records in that zone
func Query(zone string, ctx context.Context, c config.Config) ([]string, error) {
	var txts []string
	var err error
	if c.Resolver != "" {
		net := CustomResolver(c)
		txts, err = net.LookupTXT(ctx, absoluteZone(zone))
	} else {
		txts, err = net.LookupTXT(absoluteZone(zone))
	}
	if err != nil {
		return nil, fmt.Errorf("could not get TXT record: %s", err)
	}
	if txts[0] == "" {
		return nil, fmt.Errorf("TXT record doesn't exist or is empty")
	}
	return txts, nil
}

// CustomResolver returns a net.Resolver instance based
// on the given txtdirect config to use a custom DNS resolver.
func CustomResolver(c config.Config) net.Resolver {
	return net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, c.Resolver)
		},
	}
}

func absoluteZone(zone string) string {
	// Removes port from zone
	if strings.Contains(zone, ":") {
		zoneSlice := strings.Split(zone, ":")
		zone = zoneSlice[0]
	}

	if !strings.HasPrefix(zone, variables.Basezone) {
		zone = strings.Join([]string{variables.Basezone, zone}, ".")
	}

	if strings.HasSuffix(zone, ".") {
		return zone
	}

	return strings.Join([]string{zone, "."}, "")
}
