package types

type S3EventMessage struct {
	Detail S3Detail `json:"detail"`
}

type S3Detail struct {
	Object S3Object `json:"object"`
}

type S3Object struct {
	Key string `json:"key"`
}

type SinchWebhookPayload struct {
	Message SinchMessage `json:"message"`
}

type SinchMessage struct {
	ContactMessage  SinchContactMessage  `json:"contact_message"`
	ChannelIdentity SinchChannelIdentity `json:"channel_identity"`
}

type SinchContactMessage struct {
	TextMessage SinchTextMessage `json:"text_message"`
}

type SinchTextMessage struct {
	Text string `json:"text"`
}

type SinchChannelIdentity struct {
	Identity string `json:"identity"`
}
