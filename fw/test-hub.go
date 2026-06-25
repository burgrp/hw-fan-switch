//go:build !tinygo

package main

import (
	"bob/spec"

	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/site/cli"
)

func main() {
	cli.Start(inventory.Inventory{
		{
			Name: "fan",
			UID:  [12]byte{0x5A, 0x33, 0x50, 0x41, 0x12, 0x32, 0x35, 0x32, 0x29, 0x93, 0x95, 0x00},
			Key:  [16]byte{0xCD, 0x76, 0x78, 0x0D, 0x7A, 0xE4, 0x37, 0xD4, 0x4C, 0xAF, 0x34, 0xEB, 0xCB, 0x1D, 0x5B, 0xFE},
			Channel: inventory.Channel{
				Number:       37,
				Name:         "Test",
				SpreadFactor: config.SpreadFactorS8,
			},
			Type: spec.Type,
			Config: spec.Config{
				DefaultDuty: 50,
			},
		},
	})
}
