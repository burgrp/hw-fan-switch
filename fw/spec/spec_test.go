package spec

import (
	"testing"

	"github.com/burgrp/bleriot/lib/shared/firmware"
)

func TestType(t *testing.T) {
	if err := Type.Validate(); err != nil {
		t.Fatalf("Type.Validate: %v", err)
	}
	if Type.Firmware.Package != "github.com/burgrp/hw-fan-switch/fw" {
		t.Fatalf("firmware package = %q", Type.Firmware.Package)
	}
	if Type.Firmware.TinyGo.Scheduler != firmware.SchedulerTasks {
		t.Fatalf("scheduler = %q", Type.Firmware.TinyGo.Scheduler)
	}
	if Type.Firmware.TinyGo.StackSizeBytes != 1024 {
		t.Fatalf("stack size = %d", Type.Firmware.TinyGo.StackSizeBytes)
	}
}
