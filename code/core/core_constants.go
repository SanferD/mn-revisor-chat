package core

var TestStatute1 = Statute{
	Chapter: "1a", Section: "34", Title: "not a real statute",
	Subdivisions: []Subdivision{
		{Number: "1", Heading: "hello", Content: "some sample text"},
		{Number: "2a", Heading: "world", Content: "more sample text"},
	},
}

var TestStatute2 = Statute{
	Chapter: "2b", Section: "34", Title: "statute without any subdivisions",
	Subdivisions: []Subdivision{
		{Number: "", Heading: "", Content: "not really a subdivision"},
	},
}
