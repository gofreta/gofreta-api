package utils

import (
	"testing"
)

func TestGetExtAndFileTypeByMimeType(t *testing.T) {
	testItems := []struct {
		Ext  string
		Type string
		Data []byte
	}{
		// byte codes from https://golang.org/src/net/http/sniff.go?s=645:687#L54
		{"gif", FILE_TYPE_IMAGE, []byte("GIF87a")},
		{"mp3", FILE_TYPE_AUDIO, []byte("\x49\x44\x33")},
		{"webm", FILE_TYPE_VIDEO, []byte("\x1A\x45\xDF\xA3")},
		{"zip", FILE_TYPE_OTHER, []byte("\x50\x4B\x03\x04")},
		{"pdf", FILE_TYPE_DOC, []byte("%PDF-")},
		{"", "", []byte("invalid!")},
	}

	for _, item := range testItems {
		ext, ftype := GetExtAndFileTypeByMimeType(item.Data)

		if ext != item.Ext {
			t.Errorf("Expected extension %s, got %s", item.Ext, ext)
		}

		if ftype != item.Type {
			t.Errorf("Expected file type %s, got %s", item.Ext, ftype)
		}
	}
}

func TestGetMimeTypesByExt(t *testing.T) {
	expectedMimeTypes := []string{
		"audio/mp3",
		"audio/mpeg",
		"audio/mpeg3",
		"audio/x-mpeg-3",
		"application/x-rar",
		"application/x-rar-compressed",
	}

	result := GetMimeTypesByExt("mp3", "rar", "unknown!")

	if len(result) != len(expectedMimeTypes) {
		t.Error("The result mime types count should match with the expected ones, got", len(result))
	}

	for _, item := range result {
		exist := false
		for _, expected := range expectedMimeTypes {
			if expected == item {
				exist = true
				break
			}
		}

		if !exist {
			t.Error("The result contains mime type that is not expected - ", item)
		}
	}
}
