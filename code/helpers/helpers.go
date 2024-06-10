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
		var builder strings.Builder
		var idSubdiv string = id
		builder.WriteString("§ ")
		builder.WriteString(id)
		if len(subdivision.Number) > 0 {
			idSubdiv = idSubdiv + "." + subdivision.Number
			builder.WriteString(", subd. ")
			builder.WriteString(subdivision.Number)
		}
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

func ChunkObjectKeyToID(chunkObjectKey string) string {
	chunkFileNameParts := strings.Split(chunkObjectKey, "/")
	chunkFileName := chunkFileNameParts[len(chunkFileNameParts)-1]
	chunkIDParts := strings.Split(chunkFileName, ".")
	chunkID := strings.Join(chunkIDParts[:len(chunkIDParts)-1], ".")
	return chunkID
}
