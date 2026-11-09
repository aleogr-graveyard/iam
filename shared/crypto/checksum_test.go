package crypto

import (
	"encoding/hex"
	"testing"
)

// The vector is SHA-256("abc") from FIPS 180-2, appendix B.1.
func TestChecksum_IsSHA256(t *testing.T) {
	got := Checksum([]byte("abc"))
	if hex.EncodeToString(got) != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("Checksum(abc) = %x", got)
	}
	if len(Checksum(nil)) != ChecksumBytes {
		t.Fatal("empty input: wrong length")
	}
}
