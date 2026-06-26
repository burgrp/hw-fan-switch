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
			UID:  [12]byte{0x4C, 0x34, 0x50, 0x41, 0x08, 0x39, 0x32, 0x37, 0x44, 0xE2, 0xEA, 0x00},
			Key:  [16]byte{0xF7, 0xBC, 0x67, 0x20, 0xB8, 0x4E, 0x5E, 0x73, 0xB6, 0x43, 0x55, 0xA3, 0xA4, 0x91, 0x57, 0xA5},
			Channel: inventory.Channel{
				Number:       37,
				Name:         "Test",
				SpreadFactor: config.SpreadFactorS8,
			},
			Type: spec.Type,
			Config: spec.Config{
				DefaultDuty:      30,
				LowDutyThreshold: 5,
				LowDutyKickstart: 30,
			},
		},
	})
}
