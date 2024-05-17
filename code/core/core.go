package core

import (
	"context"
	"io"
)

type URLQueueMessage struct {
	QueueMessage
}

type QueueMessage struct {
	ID     string
	Body   string
	Handle string
}

type MNRevisorPageKind int

type Subdivision struct {
	Number  int
	Heading string
	Content string
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
	Clear(context.Context) error
	SendURL(context.Context, string) error
	ReceiveURLQueueMessage(context.Context) (*URLQueueMessage, error)
	DeleteURLQueueMessage(context.Context, *URLQueueMessage) error
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

type StatutesDataStore interface {
	GetStatute(context.Context, string) (Statute, error)
	PutStatute(context.Context, Statute) error
	DeleteStatute(context.Context, Statute) error
}

type WebClient interface {
	GetHTML(context.Context, string) ([]byte, error)
}

type MNRevisorStatutesScraper interface {
	GetPageKind(io.Reader) (MNRevisorPageKind, error)
	ExtractURLs(io.Reader, MNRevisorPageKind) ([]string, error)
	ExtractStatute(io.Reader) (Statute, error)
}
