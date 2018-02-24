package utils

import "net/http"

const FILE_TYPE_IMAGE = "image"
const FILE_TYPE_DOC = "doc"
const FILE_TYPE_AUDIO = "audio"
const FILE_TYPE_VIDEO = "video"
const FILE_TYPE_OTHER = "other"

// mimeTypes specifies mapped list of mime types and their appropriate extensions.
var mimeTypes = map[string]map[string]string{
	FILE_TYPE_IMAGE: {
		"image/jpeg": "jpeg",
		"image/jpg":  "jpg",
		"image/png":  "png",
		"image/gif":  "gif",
	},
	FILE_TYPE_DOC: {
		"application/pdf":                                 "pdf",
		"application/msword":                              "doc",
		"application/vnd.ms-excel":                        "xls",
		"application/vnd.ms-powerpoint":                   "ppt",
		"application/vnd.oasis.opendocument.text":         "txt",
		"application/vnd.oasis.opendocument.graphics":     "odg",
		"application/vnd.oasis.opendocument.presentation": "odp",
		"application/vnd.oasis.opendocument.spreadsheet":  "ods",
	},
	FILE_TYPE_AUDIO: {
		"audio/aac":      "aac",
		"audio/midi":     "midi",
		"audio/ogg":      "oga",
		"audio/x-wav":    "wav",
		"audio/webm":     "weba",
		"audio/mp3":      "mp3",
		"audio/mpeg":     "mp3",
		"audio/mpeg3":    "mp3",
		"audio/x-mpeg-3": "mp3",
	},
	FILE_TYPE_VIDEO: {
		"video/mpeg":      "mpeg",
		"video/ogg":       "ogv",
		"video/webm":      "webm",
		"image/webp":      "webp",
		"video/x-msvideo": "avi",
		"video/mp4":       "mp4",
	},
	FILE_TYPE_OTHER: {
		"application/zip":              "zip",
		"application/x-tar":            "tar",
		"application/x-rar":            "rar",
		"application/x-rar-compressed": "rar",
		"application/x-7z-compressed":  "7z",
	},
}

// GetExtAndFileTypeByMimeType returns data extension and file type by its mime type.
func GetExtAndFileTypeByMimeType(data []byte) (string, string) {
	contentType := http.DetectContentType(data)

	for fileType, list := range mimeTypes {
		for mimeType, ext := range list {
			if mimeType == contentType {
				return ext, fileType
			}
		}
	}

	return "", ""
}

// GetMimeTypesByExt returns all mime types related to extension(s).
func GetMimeTypesByExt(extensions ...string) []string {
	var result []string

	for _, list := range mimeTypes {
		for mimeType, ext := range list {
			for _, item := range extensions {
				if ext == item {
					result = append(result, mimeType)
				}
			}
		}
	}

	return result
}
