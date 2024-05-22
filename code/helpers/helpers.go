package helpers

import (
	"code/core"
	"encoding/base64"
	"strings"
)

func IsLocalhostURL(url string) bool {
	return strings.HasPrefix(url, "localhost") || strings.HasPrefix(url, "http://localhost") || strings.HasPrefix(url, "https://localhost")

}

func Base64Encode(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func Statute2SubdivisionChunks(statute core.Statute) []core.Chunk {
	var chunks []core.Chunk = make([]core.Chunk, 0)
	id := statute.Chapter + "." + statute.Section
	for _, subdivision := range statute.Subdivisions {
		var idSubdiv string = id
		if len(subdivision.Number) > 0 {
			idSubdiv = idSubdiv + "." + subdivision.Number
		}
		var builder strings.Builder
		builder.WriteString(idSubdiv)
		builder.WriteString(": ")
		builder.WriteString(statute.Title)
		if len(subdivision.Heading) > 0 {
			builder.WriteString(" -- ")
			builder.WriteString(subdivision.Heading)
		}
		builder.WriteString("\n")
		builder.WriteString(subdivision.Content)
		if !strings.HasSuffix(subdivision.Content, "\n") {
			builder.WriteString("\n")
		}

		chunk := core.Chunk{ID: idSubdiv, Body: builder.String()}
		chunks = append(chunks, chunk)
	}
	return chunks

}
