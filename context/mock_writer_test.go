package context

import (
	"io"
	"net/http"
)

// mockErrorWriter implements http.ResponseWriter and returns an error on Write
type mockErrorWriter struct {
	headers http.Header
	code    int
}

func newMockErrorWriter() *mockErrorWriter {
	return &mockErrorWriter{
		headers: make(http.Header),
		code:    http.StatusOK,
	}
}

func (m *mockErrorWriter) Header() http.Header {
	return m.headers
}

func (m *mockErrorWriter) Write(b []byte) (int, error) {
	// Always return an error
	return 0, io.ErrUnexpectedEOF
}

func (m *mockErrorWriter) WriteHeader(code int) {
	m.code = code
}

// errorReader is a mock io.Reader that always returns an error
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
