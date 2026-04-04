package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/andybalholm/brotli"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/klauspost/compress/zstd"
)

// NewEncoderMiddleware creates a middleware for compressing HTTP responses.
func NewEncoderMiddleware(level int) func(next http.Handler) http.Handler {
	compressor := middleware.NewCompressor(level)

	compressor.SetEncoder("br", func(w io.Writer, level int) io.Writer {
		return brotli.NewWriterLevel(w, level)
	})

	compressor.SetEncoder("zstd", func(w io.Writer, level int) io.Writer {
		enc, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
		if err != nil {
			slog.Error("could not create zstd writer", slog.Any("err", err))
			os.Exit(1)
		}
		return enc
	})

	return compressor.Handler
}
