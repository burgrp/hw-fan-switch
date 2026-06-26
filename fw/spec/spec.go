package spec

import "github.com/burgrp/bleriot/lib/shared/inventory"

type Config struct {
	DefaultDuty uint32
	// LowDutyThreshold is the duty cycle (0-100) below which the fan is treated
	// as stopped and the PWM output is forced to 0. Zero disables the threshold.
	LowDutyThreshold uint32
	// LowDutyKickstart is the duty cycle (0-100) briefly applied to spin up the
	// fan when a non-zero duty below this value is requested. Zero disables the
	// kickstart pulse.
	LowDutyKickstart uint32
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
			Metadata: map[string]string{
				"unit": "%",
			},
		},
	},
}
