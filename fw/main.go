//go:build tinygo

package main

import (
	"device/py32"
	"encoding/binary"
	"machine"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"bob/spec"

	"github.com/burgrp/bleriot/lib/node"
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/bleriot/lib/shared/protocol"

	"github.com/burgrp/tinygo-drivers/bb/spi"
	"github.com/burgrp/tinygo-drivers/pan211x"
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

func main() {
	println("fan-switch starting...")

	device := &Device{}

	pinLed.Configure(machine.PinConfig{Mode: machine.PinOutput})
	device.setupFanPWM()

	pinLed.High()
	time.Sleep(500 * time.Millisecond)
	pinLed.Low()

	pageData := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(spec.Chip.PageAddr))), spec.Chip.PageBytes)
	header, cfgBytes, err := config.Decode(pageData)
	if err != nil {
		if config.IsUnprovisioned(err) {
			haltBlink("unprovisioned", 1000*time.Millisecond)
		}
		haltBlink("bad page: "+err.Error(), 100*time.Millisecond)
	}
	cfg := spec.Config{
		DefaultDuty: binary.LittleEndian.Uint32(cfgBytes[0:4]),
	}

	println("Provisioned: channel", int(header.Channel), "spreadFactor", int(header.SpreadFactor))

	radio := pan211x.NewDriverBLELongRange(
		pan211x.NewRegistersSPI(spi.NewMaster(pinSpiSck, pinSpiData), pinSpiCsn))
	must(radio.Init(pan211x.ConfigBLELongRange{
		PayloadLen:      protocol.PacketLen,
		SerialInterface: pan211x.SerialInterfaceSPI3W,
		SpreadFactor:    pan211x.SpreadFactor(header.SpreadFactor),
	}))
	must(radio.SetChannelRF(header.Channel, header.Channel))
	must(radio.EnableRxAddress(0, header.Address))
	println("Radio initialized")

	println("Device config: defaultDuty", cfg.DefaultDuty)

	device.duty.Store(int32(cfg.DefaultDuty))
	device.setFanDuty(int32(cfg.DefaultDuty))

	node, err := node.New(radio, header.Address, header.Key, device)
	must(err)
	device.node = node

	//go memstat()

	for {
		node.Poll()
		runtime.Gosched()
	}

}

func memstat() {
	for {
		mem := runtime.MemStats{}
		runtime.ReadMemStats(&mem)
		println("mem: alloc", mem.Alloc, "sys", mem.Sys, "alloc", mem.HeapAlloc)
		time.Sleep(1 * time.Second)
	}
}

type Device struct {
	node      *node.Node
	duty      atomic.Int32
	pwmPeriod uint32
}

func (d *Device) Read(tag uint16) (value int32, null bool) {

	switch tag {
	case spec.RegDuty:
		return d.duty.Load(), false
	default:
		// unknown tag: report null
	}

	return 0, true
}

func (d *Device) Write(tag uint16, value int32, null bool) {

	switch tag {
	case spec.RegDuty:
		d.duty.Store(value)
		d.setFanDuty(value)
		d.node.Notify(spec.RegDuty, value, null)
	default:
		// unknown tag: ignore
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
	if duty < 0 {
		duty = 0
	}
	if duty > 100 {
		duty = 100
	}
	py32.TIM14.SetCCR1(uint32(duty) * d.pwmPeriod / 100)
}

func (d *Device) ledLoop(pin machine.Pin, period *atomic.Int32) {
	for {
		v := period.Load()
		switch {
		case v == 0:
			pin.Low()
			time.Sleep(100 * time.Millisecond)
		case v == 1:
			pin.High()
			time.Sleep(100 * time.Millisecond)
		default:
			p := time.Duration(v) * time.Millisecond
			pin.High()
			time.Sleep(p)
			pin.Low()
			time.Sleep(p)
		}
	}
}

func must(err error) {
	if err == nil {
		return
	}
	haltBlink("fatal: "+err.Error(), 100*time.Millisecond)
}

func haltBlink(msg string, period time.Duration) {
	println(msg)
	for {
		pinLed.High()
		time.Sleep(period)
		pinLed.Low()
		time.Sleep(period)
	}
}
