package spec

import (
	"github.com/burgrp/bleriot/lib/shared/firmware"
	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/shared/puya"
)

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

var Chip = puya.PY32F003x6

var Type = inventory.DeviceType{
	Name: "fan",
	Chip: Chip,
	Firmware: firmware.Manifest{
		Package: "github.com/burgrp/hw-fan-switch/fw",
		TinyGo: firmware.TinyGoProfile{
			Scheduler:        firmware.SchedulerTasks,
			StackSizeBytes:   1024,
			GarbageCollector: firmware.GCLeaking,
			Serial:           firmware.SerialRTT,
			SizeReport:       firmware.SizeReportHTML,
			PrintAllocs:      true,
		},
		PyOCD: firmware.PyOCDProfile{
			Reclaim:                  true,
			ReclaimDelayMilliseconds: 1000,
		},
	},
	Registers: []inventory.Register{
		{
			Tag:  RegDuty,
			Name: "duty",
			Type: inventory.TypeInt,
			Metadata: map[string]string{
				"unit": "%",
			},
		},
	},
}
