package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

type Config struct {
	DefaultDuty uint32
}

const (
	RegDuty = 1 // PWM duty cycle, 0-100
)

var Chip = inventory.PY32F003x6

var Type = inventory.DeviceType{
	Name: "fan",
	Chip: Chip,
	Registers: []inventory.Register{
		{
			Tag:        RegDuty,
			Name:       "duty",
			Type:       inventory.TypeInt,
			Multiplier: 1,
			Divider:    1,
		},
	},
}
