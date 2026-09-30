package contextband

import (
	"strconv"
	"strings"
	"testing"
)

func line(typ string, sidechain bool, in, read, create int) string {
	return `{"type":"` + typ + `","isSidechain":` + map[bool]string{true: "true", false: "false"}[sidechain] +
		`,"message":{"usage":{"input_tokens":` + strconv.Itoa(in) + `,"cache_read_input_tokens":` + strconv.Itoa(read) +
		`,"cache_creation_input_tokens":` + strconv.Itoa(create) + `,"output_tokens":5}}}`
}

func TestTokens(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  int
	}{
		{"sums the three inputs", []string{line("assistant", false, 8, 100000, 1500)}, 101508},
		{"last main line, not the largest", []string{line("assistant", false, 0, 200000, 0), line("assistant", false, 0, 20000, 0)}, 20000},
		{"sidechain ignored", []string{line("assistant", false, 0, 1000, 0), line("assistant", true, 0, 9000, 0)}, 1000},
		{"zero usage skipped", []string{line("assistant", false, 0, 1000, 0), line("assistant", false, 0, 0, 0)}, 1000},
		{"other lines skipped", []string{line("assistant", false, 0, 1000, 0), `{"type":"user"}`, `not json`, `{"type":"summary"}`}, 1000},
		{"no assistant line", []string{`{"type":"user"}`}, 0},
		{"no trailing newline", []string{line("assistant", false, 0, 7, 0) + "\n" + line("assistant", false, 0, 9, 0)}, 9},
	}
	for _, c := range cases {
		if got := tokens(strings.NewReader(strings.Join(c.lines, "\n"))); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

func TestBand(t *testing.T) {
	bands := []int{125000, 150000, 200000}
	cases := []struct{ tokens, want, index int }{
		{0, 0, -1}, {124999, 0, -1}, {125000, 125000, 0}, {149999, 125000, 0}, {160000, 150000, 1}, {999999, 200000, 2},
	}
	for _, c := range cases {
		if got, i := band(bands, c.tokens); got != c.want || i != c.index {
			t.Errorf("band(%d) = %d, %d; want %d, %d", c.tokens, got, i, c.want, c.index)
		}
	}
	if got, _ := band([]int{200000, 125000}, 130000); got != 125000 {
		t.Errorf("unsorted bands: got %d", got)
	}
}

func TestLabel(t *testing.T) {
	for n, want := range map[int]string{125000: "125K", 50000: "50K", 1500: "1.5K"} {
		if got := label(n); got != want {
			t.Errorf("label(%d) = %q, want %q", n, got, want)
		}
	}
}
