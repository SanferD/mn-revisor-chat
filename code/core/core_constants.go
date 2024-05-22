package core

var TestStatute1 = Statute{
	Chapter: "1a", Section: "34", Title: "not a real statute",
	Subdivisions: []Subdivision{
		{Number: "1", Heading: "hello", Content: "some sample text"},
		{Number: "2a", Heading: "world", Content: "more sample text"},
	},
}

var Chunk11 = Chunk{ID: "1a.34.1", Body: "1a.34.1: not a real statute -- hello\nsome sample text\n"}
var Chunk12 = Chunk{ID: "1a.34.2a", Body: "1a.34.2a: not a real statute -- world\nmore sample text\n"}

var TestStatute2 = Statute{
	Chapter: "2b", Section: "34", Title: "statute without any subdivisions",
	Subdivisions: []Subdivision{
		{Number: "", Heading: "", Content: "not really a subdivision"},
	},
}

var Chunk21 = Chunk{ID: "2b.34", Body: "2b.34: statute without any subdivisions\nnot really a subdivision\n"}
