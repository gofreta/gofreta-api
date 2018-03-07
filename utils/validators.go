package utils

import (
	"encoding/binary"
	"math"
	"net/http"
)

// ValidateMimeType validates data mime type.
func ValidateMimeType(data []byte, validTypes []string) bool {
	filetype := http.DetectContentType(data)

	if validTypes != nil {
		for _, t := range validTypes {
			if t == filetype {
				return true
			}
		}

		return false
	}

	return true
}

// ValidateSize validates data size.
func ValidateSize(data []byte, maxValidSize float64) bool {
	size := binary.Size(data)

	mb := float64(size) * math.Pow10(-6)

	return mb <= maxValidSize
}
