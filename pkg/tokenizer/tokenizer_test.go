package tokenizer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"testing"
)

// TestEmbeddedVocabularyHash pins the embedded vocabulary to the file tiktoken
// publishes, so it cannot drift or be truncated unnoticed.
func TestEmbeddedVocabularyHash(t *testing.T) {
	const want = "223921b76ee99bde995b7ff738513eef100fb51d18c93597a113bcffe865b2a7"
	sum := sha256.Sum256(cl100kBase)
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("embedded cl100k_base.tiktoken sha256 = %s, want %s", got, want)
	}
}

// TestEncodeMatchesTiktoken compares against token IDs produced by
// tiktoken-go's own cl100k_base encoding (downloaded vocabulary).
func TestEncodeMatchesTiktoken(t *testing.T) {
	tok, err := NewTokenizer()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		text string
		want []int
	}{
		{"hello world", []int{15339, 1917}},
		{"tiktoken is great!", []int{83, 1609, 5963, 374, 2294, 0}},
		{"Ünïcödé 文字 and emoji-free text", []int{53591, 77, 38672, 66, 3029, 67, 978, 54140, 19113, 323, 43465, 12862, 1495}},
		{"  leading spaces\n\ttabs\r\nCRLF", []int{220, 6522, 12908, 198, 3324, 3518, 319, 34, 81758}},
		{"<|endoftext|> is not special here", []int{27, 91, 8862, 728, 428, 91, 29, 374, 539, 3361, 1618}},
		{"func main() { fmt.Println(\"x\") }", []int{2900, 1925, 368, 314, 9055, 12701, 446, 87, 909, 335}},
	}
	for _, tt := range tests {
		if got := tok.Encode(tt.text); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Encode(%q) = %v, want %v", tt.text, got, tt.want)
		}
		if got := tok.Count(tt.text); got != len(tt.want) {
			t.Errorf("Count(%q) = %d, want %d", tt.text, got, len(tt.want))
		}
	}
}

// TestNewTokenizerOffline verifies that no download is attempted: with an
// empty tiktoken cache and an unreachable proxy, a download would fail.
func TestNewTokenizerOffline(t *testing.T) {
	t.Setenv("TIKTOKEN_CACHE_DIR", t.TempDir())
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")

	tok, err := NewTokenizer()
	if err != nil {
		t.Fatalf("NewTokenizer() without network: %v", err)
	}
	if got := tok.Count("hello world"); got != 2 {
		t.Errorf("Count() = %d, want 2", got)
	}
}

// TestParseRanksCRLF verifies that a vocabulary with CRLF line endings parses
// the same as one with LF endings.
func TestParseRanksCRLF(t *testing.T) {
	lf, err := parseRanks(cl100kBase)
	if err != nil {
		t.Fatal(err)
	}
	crlf, err := parseRanks(bytes.ReplaceAll(cl100kBase, []byte("\n"), []byte("\r\n")))
	if err != nil {
		t.Fatalf("parseRanks() with CRLF line endings: %v", err)
	}
	if !reflect.DeepEqual(lf, crlf) {
		t.Error("CRLF and LF vocabularies parsed differently")
	}
}
