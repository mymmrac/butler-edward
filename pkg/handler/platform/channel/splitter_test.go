package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		text      string
		maxLength int
		expected  []string
	}{
		{
			name:      "simple_split_by_sentence",
			text:      "Sentence one. Sentence two.",
			maxLength: 15,
			expected:  []string{"Sentence one.", "Sentence two."},
		},
		{
			name:      "split_by_word",
			text:      "One two three four five",
			maxLength: 10,
			expected:  []string{"One two", "three four", "five"},
		},
		{
			name:      "split_by_char",
			text:      "VeryLongWordThatMustBeSplit",
			maxLength: 5,
			expected:  []string{"VeryL", "ongWo", "rdTha", "tMust", "BeSpl", "it"},
		},
		{
			name:      "mixed_split",
			text:      "Short sentence. AVeryLongWordThatNeedsCharSplit And then words.",
			maxLength: 10,
			expected:  []string{"Short", "sentence.", "AVeryLongW", "ordThatNee", "dsCharSpli", "t", "And then", "words."},
		},
		{
			name:      "no_split_needed",
			text:      "Hello world",
			maxLength: 20,
			expected:  []string{"Hello world"},
		},
		{
			name:      "empty_text",
			text:      "",
			maxLength: 10,
			expected:  []string{""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := SplitText(tc.text, tc.maxLength)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestSplitTextBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		text      string
		maxLength int
		delimiter string
		expected  []string
	}{
		{
			name:      "simple_split",
			text:      "one two three",
			maxLength: 5,
			delimiter: " ",
			expected:  []string{"one", "two", "three"},
		},
		{
			name:      "no_split_needed",
			text:      "one two",
			maxLength: 10,
			delimiter: " ",
			expected:  []string{"one two"},
		},
		{
			name:      "split_with_longer_words",
			text:      "a b cde fgh i",
			maxLength: 3,
			delimiter: " ",
			expected:  []string{"a b", "cde", "fgh", "i"},
		},
		{
			name:      "split_by_dot_space",
			text:      "Sentence one. Sentence two. Sentence three.",
			maxLength: 15,
			delimiter: ". ",
			expected:  []string{"Sentence one.", "Sentence two.", "Sentence three."},
		},
		{
			name:      "impossible_to_split",
			text:      "toolongword",
			maxLength: 5,
			delimiter: " ",
			expected:  []string{"toolongword"},
		},
		{
			name:      "empty_text",
			text:      "",
			maxLength: 5,
			delimiter: " ",
			expected:  []string{""},
		},
		{
			name:      "empty_delimiter",
			text:      "abcd",
			maxLength: 2,
			delimiter: "",
			expected:  []string{"ab", "cd"},
		},
		{
			name:      "split_exact_max_length",
			text:      "123 456",
			maxLength: 3,
			delimiter: " ",
			expected:  []string{"123", "456"},
		},
		{
			name:      "split_with_multiple_delimiters",
			text:      "a  b",
			maxLength: 1,
			delimiter: " ",
			expected:  []string{"a", "", "b"},
		},
		{
			name:      "split_with_multiple_delimiters_max_2",
			text:      "a  b",
			maxLength: 2,
			delimiter: " ",
			expected:  []string{"a ", "b"},
		},
		{
			name:      "split_by_comma_not_whitespace",
			text:      "a,b,c",
			maxLength: 2,
			delimiter: ",",
			expected:  []string{"a,", "b,", "c"},
		},
		{
			name:      "split_by_comma_with_whitespace",
			text:      "a, b, c",
			maxLength: 2,
			delimiter: ", ",
			expected:  []string{"a,", "b,", "c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := SplitTextBy(tc.text, tc.maxLength, tc.delimiter)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
