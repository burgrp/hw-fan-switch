package fanswitch

func normalizedDuty(value int32, null bool) int32 {
	if null {
		return 0
	}
	return clipValue(value)
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
