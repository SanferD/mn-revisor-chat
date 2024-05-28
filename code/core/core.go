package core

import (
	"context"
	"io"
)

type QueueMessage struct {
	Body    string
	Handle  string
	IsEmpty bool
}

type MNRevisorPageKind int

type Chunk struct {
	ID   string
	Body string
}

type Subdivision struct {
	Number  string
	Heading string
	Content string
}

type VectorDocument struct {
	ID     string
	Vector []float64
}

const (
	MNRevisorPageKindError MNRevisorPageKind = -1
	StatutesChaptersTable  MNRevisorPageKind = iota
	StatutesChaptersShortTable
	StatutesSectionsTable
	Statutes
)

type Statute struct {
	Chapter      string
	Section      string
	Title        string
	Subdivisions []Subdivision
}

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Debug(string, ...any)
	Error(string, ...any)
	Fatal(string, ...any)
}

type InterruptWatcher interface {
	StartBackgroundWatcher()
	IsInterrupted() bool
}

type URLQueue interface {
	SendURL(context.Context, string) error
	Queue
}

type RawEventsQueue interface {
	Queue
}

type Queue interface {
	Clear(context.Context) error
	SendMessage(context.Context, QueueMessage) error
	ReceiveMessage(context.Context) (QueueMessage, error)
	DeleteMessage(context.Context, QueueMessage) error
	DeleteMessageByHandle(context.Context, string) error
}

type SeenURLStore interface {
	PutURL(context.Context, string) error
	HasURL(context.Context, string) (bool, error)
	DeleteAll(context.Context) error
}

type RawDataStore interface {
	GetTextFile(context.Context, string) (string, error)
	PutTextFile(context.Context, string, io.Reader) error
	DeleteTextFile(context.Context, string) error
}

type ChunksDataStore interface {
	PutChunk(context.Context, Chunk) error
	GetChunk(context.Context, string) (Chunk, error)
}

type WebClient interface {
	GetHTML(context.Context, string) ([]byte, error)
}

type MNRevisorStatutesScraper interface {
	GetPageKind(io.Reader) (MNRevisorPageKind, error)
	ExtractURLs(io.Reader, MNRevisorPageKind) ([]string, error)
	ExtractStatute(io.Reader) (Statute, error)
}

type Invoker interface {
	InvokeTriggerCrawler(context.Context) error
	IsTriggerCrawlerAlreadyRunning(context.Context) (bool, error)
}

type Agent interface {
	Ask(context.Context, string, string) (string, error)
}

type Vectorizer interface {
	VectorizeChunk(context.Context, Chunk) (VectorDocument, error)
}
