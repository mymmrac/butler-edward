package channel

import "strings"

// SplitText splits text into chunks of specified length.
func SplitText(text string, maxLength int) []string {
	chunks := make([]string, 0, (len(text)/maxLength)+1)
	sentenceChunks := SplitTextBy(text, maxLength, ". ")
	for _, sentenceChunk := range sentenceChunks {
		if len(sentenceChunk) > maxLength {
			wordChunks := SplitTextBy(sentenceChunk, maxLength, " ")
			for _, wordChunk := range wordChunks {
				if len(wordChunk) > maxLength {
					charChunks := SplitTextBy(wordChunk, maxLength, "")
					chunks = append(chunks, charChunks...)
				} else {
					chunks = append(chunks, wordChunk)
				}
			}
		} else {
			chunks = append(chunks, sentenceChunk)
		}
	}
	return chunks
}

// SplitTextBy splits text into chunks of specified length by delimiter. Can output longer chunks than max length if
// it's impossible to split by delimiter to the desired chunk size.
func SplitTextBy(text string, maxLength int, delimiter string) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	trimmedDelimiter := strings.TrimSpace(delimiter)
	isWhitespace := trimmedDelimiter == ""

	parts := strings.Split(text, delimiter)
	chunks := make([]string, 0, (len(text)/maxLength)+1)

	hasChunk := false
	chunk := &strings.Builder{}
	for i, part := range parts {
		if hasChunk {
			if chunk.Len()+len(delimiter)+len(part) > maxLength {
				if !isWhitespace {
					_, _ = chunk.WriteString(trimmedDelimiter)
				}
				chunks = append(chunks, chunk.String())
				chunk.Reset()
				hasChunk = false
			}
		}

		if !hasChunk && len(part) > maxLength {
			if !isWhitespace && i < len(parts)-1 {
				part += trimmedDelimiter
			}
			chunks = append(chunks, part)
			continue
		}

		if hasChunk {
			_, _ = chunk.WriteString(delimiter)
		}

		_, _ = chunk.WriteString(part)
		hasChunk = true
	}

	if hasChunk {
		chunks = append(chunks, chunk.String())
	}

	return chunks
}
