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
