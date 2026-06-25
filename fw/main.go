//go:build tinygo

package main

import (
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

	// PAN211x over 3-wire SPI.
	pinSpiSck  = machine.PA9  // SCK  → PAN211x pin 2
	pinSpiData = machine.PA7  // DATA → PAN211x pin 3, bidirectional
	pinSpiCsn  = machine.PA10 // CSN  → PAN211x pin 1, active-low
)

func main() {
	println("fan-switch starting...")

	pinLed.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinFan.Configure(machine.PinConfig{Mode: machine.PinOutput})

	pinFan.Low()
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

	device := &Device{}

	device.duty.Store(int32(cfg.DefaultDuty))

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
	node *node.Node
	duty atomic.Int32
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
		d.node.Notify(spec.RegDuty, value, null)
	default:
		// unknown tag: ignore
	}
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
