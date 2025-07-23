package messaging

type SubscribeEvent struct {
	Email string `json:"email"`
	Token string `json:"token"`
}
