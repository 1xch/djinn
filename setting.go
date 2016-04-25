package djinn

type settings struct {
	cached bool
}

func defaultSettings() *settings {
	return &settings{
		cached: true,
	}
}
