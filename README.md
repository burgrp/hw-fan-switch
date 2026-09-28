# hw-fan-switch

A small, wirelessly-controlled 12 V DC fan switch. It drives a fan with a
variable PWM duty cycle and is controlled over a BLE long-range radio link,
making it part of the [Bleriot](https://github.com/burgrp/bleriot) node network.

![PCB](pcb.png)

The board is built around a **Puya PY32F003** Cortex-M0+ microcontroller with a
**PAN211x** radio, and the firmware is written in Go and compiled with
[TinyGo](https://tinygo.org/).

## Features

- Variable-speed fan control via hardware PWM (TIM14_CH1 on `PB1`).
- 50 kHz PWM carrier, above the audible range to avoid switching noise.
- Configurable low-speed handling:
  - **Low-duty threshold** — below this duty the fan is treated as stopped and
    the output is forced to 0.
  - **Low-duty kickstart** — a brief higher-duty pulse to overcome fan stiction
    when starting at low speed.
- Wireless control and telemetry over the PAN211x BLE long-range link.
- Device identity, radio settings, and configuration are owned by the site
  inventory and baked into each firmware image at build time.

## Repository layout

| Path | Description |
| --- | --- |
| `board/` | KiCad hardware design (schematic, PCB, BOM, production files). |
| `fw/` | Importable TinyGo firmware and board-owned build profile. |
| `fw/spec/` | Device specification: registers and runtime `Config`. |
| `fw/cmd/dev/` | Optional local development inventory. |
| `sub/hw-kicad/` | Shared KiCad symbol/footprint library (git submodule). |

## Hardware

| Function | Pin |
| --- | --- |
| Status LED | `PB0` |
| Fan PWM output (TIM14_CH1, AF0) | `PB1` |
| Radio SPI clock | `PA2` |
| Radio SPI data | `PA1` |
| Radio SPI chip select | `PA4` |

The KiCad project lives in `board/`, with fabrication outputs (Gerbers, BOM,
positions) under `board/production/`.

## Firmware

The firmware is the importable `fanswitch` package. Its TinyGo implementation
exports `Run`, while `fw/spec` exposes `Config`, the register table, chip, and a
typed build/flash profile. BleRiot generates a private external `package main`
inside the deployment module and calls `Run` with that instance's baked identity
and configuration. Nothing is generated in this repository for a deployment
build.

### Prerequisites

- [TinyGo](https://tinygo.org/getting-started/install/)
- [pyOCD](https://pyocd.io/) for flashing and RTT logging
- `arm-none-eabi-objdump` (optional, for disassembly)

Clone with submodules:

```sh
git clone --recurse-submodules https://github.com/burgrp/hw-fan-switch.git
```

### Build and flash

Build deployment firmware from the authoritative site inventory. For the AGC
checkout and its `basement.fan` instance:

```sh
cd /home/paul/git/agc/bleriot

# Inspect the generated entry point without writing it
go run . node gen --name basement.fan

# One-time: install the target CMSIS pack
go run . node install-pack --name basement.fan

# Generate the private entry point and build image.elf
go run . node build --name basement.fan

# Generate, build, flash, and open RTT
go run . node build --name basement.fan --flash --rtt

# Open RTT without flashing
go run . node rtt --name basement.fan
```

The `bleriot node` commands resolve the selected inventory instance. `node
build` generates its identity and `Config` under that deployment's private
`.bleriot` directory, then builds this module using the profile carried by
`spec.Type`. The deployment pins the firmware version in `go.mod`; it does not
carry fan-specific build flags. This repository's Makefile contains optional
aliases for `fw/cmd/dev`.

## Configuration

The inventory `Config` (see `fw/spec/spec.go`) controls runtime behaviour:

| Field | Meaning |
| --- | --- |
| `DefaultDuty` | Fan duty cycle (0–100) applied at startup. |
| `LowDutyThreshold` | Duty below which the fan is forced to 0. `0` disables the threshold. |
| `LowDutyKickstart` | Duty briefly applied to spin up the fan from low speed. `0` disables the kickstart. |

## Control interface

The device exposes a single register over the Bleriot network:

| Register | Tag | Type | Unit | Description |
| --- | --- | --- | --- | --- |
| `duty` | 1 | int | % | PWM duty cycle, 0–100. |

Writing `duty` updates the fan speed; the hub's next scheduled GET observes the
new value. A NULL write stops the fan. Reading returns the current duty.
