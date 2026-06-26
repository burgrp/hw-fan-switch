//go:build tinygo

package main

import (
	"device/py32"
	"encoding/binary"
	"machine"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/burgrp/hw-fan-switch/fw/spec"

	"github.com/burgrp/bleriot/lib/node"

	"github.com/burgrp/bleriot/lib/node/pan211x"
)

const (
	pinLed = machine.PB0
	pinFan = machine.PB1

	pinSpiSck  = machine.PA2
	pinSpiData = machine.PA1
	pinSpiCsn  = machine.PA4
)

const (
	// PB1 alternate function 0 maps to TIM14_CH1 (PY32F003 datasheet Table 3-7).
	fanPwmAltFunc = 0
	// Fan PWM carrier frequency. 50 kHz keeps the carrier above the audible
	// range to avoid switching noise.
	fanPwmFreqHz = 50_000
)

type Device struct {
	node      *node.Node
	duty      atomic.Int32
	pwmPeriod uint32
	// lowDutyThreshold: duty below this is forced to 0 (0 disables the threshold).
	lowDutyThreshold int32
	// lowDutyKickstart: duty briefly applied to spin up the fan (0 disables it).
	lowDutyKickstart int32
}

func main() {
	println("fan-switch starting...")

	device := &Device{}

	pinLed.High()
	time.Sleep(500 * time.Millisecond)
	pinLed.Low()

	pinLed.Configure(machine.PinConfig{Mode: machine.PinOutput})
	device.setupFanPWM()

	node, cfgBytes, err := pan211x.StartNode(&spec.Chip, pinSpiSck, pinSpiData, pinSpiCsn, device)
	if err != nil {
		panic("failed to start node: " + err.Error())
	}
	device.node = node

	cfg := spec.Config{
		DefaultDuty:      cfgU32(cfgBytes, 0),
		LowDutyThreshold: cfgU32(cfgBytes, 4),
		LowDutyKickstart: cfgU32(cfgBytes, 8),
	}

	device.lowDutyThreshold = clipValue(int32(cfg.LowDutyThreshold))
	device.lowDutyKickstart = clipValue(int32(cfg.LowDutyKickstart))

	println("Device config: defaultDuty", cfg.DefaultDuty,
		"lowDutyThreshold", cfg.LowDutyThreshold,
		"lowDutyKickstart", cfg.LowDutyKickstart)

	device.duty.Store(clipValue(int32(cfg.DefaultDuty)))
	device.setFanDuty(int32(device.duty.Load()))

	// report free RAM with at least one poll
	node.Poll()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	println("Free RAM:", mem.Sys-mem.HeapAlloc, "bytes")

	for {
		node.Poll()
		runtime.Gosched()
	}

}

func (d *Device) Read(tag uint16) (value int32, null bool) {
	switch tag {
	case spec.RegDuty:
		return d.duty.Load(), false
	default:
		return 0, true
	}
}

func (d *Device) Write(tag uint16, value int32, null bool) {
	switch tag {
	case spec.RegDuty:
		value = clipValue(value)
		d.duty.Store(value)
		d.setFanDuty(value)
		pinLed.High()
		d.node.Notify(spec.RegDuty, value, null)
	}
}

// setupFanPWM configures TIM14 channel 1 to drive the fan on PB1 (TIM14_CH1).
func (d *Device) setupFanPWM() {
	// The timer clock equals the CPU/APB clock (APB prescaler = 1 after reset).
	period := machine.CPUFrequency() / fanPwmFreqHz
	d.pwmPeriod = period

	// Enable the TIM14 peripheral clock.
	py32.RCC.SetAPBENR2_TIM14EN(1)

	// Route PB1 to TIM14_CH1 (AF0) in alternate-function mode.
	pinFan.Configure(machine.PinConfig{Mode: machine.PinAlternate})
	pinFan.SetAltFunc(fanPwmAltFunc)

	tim := py32.TIM14
	tim.SetPSC(0)          // count at the full timer clock
	tim.SetARR(period - 1) // PWM period
	tim.SetCCR1(0)         // start at 0% duty

	// Channel 1 in PWM mode 1 (OC1M = 110) with output-compare preload.
	tim.SetCCMR1_Output_OC1M(0b110)
	tim.SetCCMR1_Output_OC1PE(1)

	// Enable auto-reload preload and the channel output (active high).
	tim.SetCR1_ARPE(1)
	tim.SetCCER_CC1E(1)

	// Load the preloaded registers via an update event, then start the counter.
	tim.SetEGR_UG(1)
	tim.SetCR1_CEN(1)
}

// setFanDuty applies a duty cycle in percent (0-100) to the fan PWM output.
func (d *Device) setFanDuty(duty int32) {

	if duty < d.lowDutyThreshold {
		duty = 0
	}

	if duty > 0 && duty < d.lowDutyKickstart {
		// Briefly drive at the kickstart duty to overcome the fan's stiction,
		// then settle to the requested low duty.
		py32.TIM14.SetCCR1(uint32(d.lowDutyKickstart) * d.pwmPeriod / 100)
		time.Sleep(100 * time.Millisecond)
	}

	py32.TIM14.SetCCR1(uint32(duty) * d.pwmPeriod / 100)
}

// cfgU32 reads a little-endian uint32 from the config bytes at the given offset,
// returning 0 when the field is absent (e.g. a shorter, older provisioning page).
func cfgU32(b []byte, off int) uint32 {
	if len(b) < off+4 {
		return 0
	}
	return binary.LittleEndian.Uint32(b[off : off+4])
}

func clipValue(value int32) int32 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
