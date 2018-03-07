package utils

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestValidateMimeType(t *testing.T) {
	validTypes := []string{
		"image/jpg",
		"image/png",
		"application/pdf",
		"application/zip",
	}

	gifBytes := []byte("GIF87a")
	if ValidateMimeType(gifBytes, validTypes) {
		t.Error("Expected to return false, got true")
	}

	pdfBytes := []byte("%PDF-")
	if !ValidateMimeType(pdfBytes, validTypes) {
		t.Error("Expected to return true, got false")
	}

	emptyBytes := []byte("")
	if ValidateMimeType(emptyBytes, validTypes) {
		t.Error("Expected to return false, got true")
	}
}

func TestValidateSize(t *testing.T) {
	data := []byte("Lorem ipsum dolor sit amet")
	sizeInMb := float64(binary.Size(data)) * math.Pow10(-6)

	if ValidateSize(data, sizeInMb-0.1) {
		t.Error("Expected to return false, since data size is larger that the max allowed size, but got true")
	}

	if !ValidateSize(data, sizeInMb) {
		t.Error("Expected to return true, since data size is exactly the same as the max allowed size, but got false")
	}

	if !ValidateSize(data, sizeInMb+0.1) {
		t.Error("Expected to return true, since data size is less than the max allowed size, but got false")
	}
}
