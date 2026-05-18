package provider

import (
	"encoding/base64"
	"strings"
	"sync"
)

var imageBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 3*1024*1024)
		return &b
	},
}

func DecodeImageBase64(b64Str string) ([]byte, error) {
	if b64Str == "" {
		return nil, ErrEmptyBase64
	}

	maxLen := base64.StdEncoding.DecodedLen(len(b64Str))

	ptr := imageBufferPool.Get().(*[]byte)
	buf := *ptr

	if cap(buf) < maxLen {
		buf = make([]byte, maxLen)
	} else {
		buf = buf[:maxLen]
	}

	n, err := base64.StdEncoding.Decode(buf, []byte(b64Str))
	if err != nil {
		imageBufferPool.Put(&buf)
		return nil, err
	}

	resultData := buf[:n]

	finalImage := make([]byte, n)
	copy(finalImage, resultData)

	imageBufferPool.Put(&buf)

	return finalImage, nil
}

func DetectMimeFromBase64(b64 string) string {
	if strings.HasPrefix(b64, "data:") {
		if idx := strings.Index(b64, ";"); idx > 5 {
			return b64[5:idx]
		}
	}

	switch {
	case strings.HasPrefix(b64, "/9j/"):
		return "image/jpeg"
	case strings.HasPrefix(b64, "iVBORw0KGgo"):
		return "image/png"
	case strings.HasPrefix(b64, "R0lGODlh"), strings.HasPrefix(b64, "R0lGODdh"):
		return "image/gif"
	case strings.HasPrefix(b64, "UklGR"):
		return "image/webp"
	}

	return "application/octet-stream"
}
