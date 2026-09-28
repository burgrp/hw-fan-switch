package fanswitch

func statusLEDOn(online, heartbeat bool) bool {
	return online || heartbeat
}
