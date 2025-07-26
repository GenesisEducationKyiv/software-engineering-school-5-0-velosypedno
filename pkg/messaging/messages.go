package messaging

type SubscribeEvent struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type Weather struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

type WeatherNotifyCommand struct {
	Email   string  `json:"email"`
	Token   string  `json:"token"`
	Weather Weather `json:"weather"`
}
