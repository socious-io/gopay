package gopay_test

import (
	"net/http"
)

// MockHTTPClient is a mock implementation of http.RoundTripper for unit testing
type MockHTTPClient struct {
	Response *http.Response
	Err      error
}

// mockReadCloser is a mock implementation of io.ReadCloser for testing HTTP responses
type mockReadCloser struct {
	data []byte
}

// RoundTrip is the mock method that will be called when an HTTP request is made
func (m *MockHTTPClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	copy(p, m.data)
	return len(m.data), nil
}

func (m *mockReadCloser) Close() error {
	return nil
}
