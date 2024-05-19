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

type Subdivision struct {
	Number  string
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
	SendURL(context.Context, string) error
	Queue
}

type RawEventsQueue interface {
	DeleteEvent(context.Context, string) error
	Queue
}

type Queue interface {
	Clear(context.Context) error
	SendMessage(context.Context, QueueMessage) error
	ReceiveMessage(context.Context) (QueueMessage, error)
	DeleteMessage(context.Context, QueueMessage) error
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
