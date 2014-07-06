package ahago

import (
	"testing"
)

func Test_getSIDResponse(t *testing.T) {
	var challenge string = "1234567z"
	var pass string = "äbc"
	soll := "1234567z-9e224a41eeefa284df7bb0f26c2913e2"
	response := getSIDResponse(challenge, pass)
	if getSIDResponse(challenge, pass) != soll {
		t.Error("SIDResponse IST: ", response, " SOLL: ", soll)
	}
}

func Test_utf8ToUtf16le(t *testing.T) {
	soll := []byte{65, 00, 66, 00, 67, 00} //A,B,C
	utf8 := []byte("ABC")
	utf16le := utf8ToUtf16le(utf8)
	for i, b := range utf16le {
		if b != soll[i] {
			t.Error("UTF8 to UTF16LE IST:", b, " SOLL:", soll[i])
		}
	}

}
