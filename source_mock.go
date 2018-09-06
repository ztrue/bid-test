package main

import "net/http"
import "strings"
import "time"

// Preset is a config for prepared response for a specific request path
type Preset struct {
  path string
  status int
  body string
  timeout time.Duration
}

// SourceMock is a mock for source HTTP service
type SourceMock struct {
  addr string
  presets []Preset
}

// NewSourceMock creates a source mock
func NewSourceMock(addr string) *SourceMock {
  return &SourceMock{addr: addr}
}

// SetUp prepares a preset for source mock
func (m *SourceMock) SetUp(
  path string,
  status int,
  body string,
  timeout ...time.Duration,
) {
  p := Preset{
    path: path,
    status: status,
    body: body,
  }
  if len(timeout) > 0 {
    p.timeout = timeout[0]
  }
  m.presets = append(m.presets, p)
}

// Run runs a source mock HTTP service
func (m *SourceMock) Run() error {
  h := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
    path := strings.Trim(r.URL.Path, "/")
    for _, p := range m.presets {
      if p.path == path {
        time.Sleep(p.timeout)
        w.WriteHeader(p.status)
        w.Write([]byte(p.body))
        return
      }
    }
    w.WriteHeader(http.StatusNotFound)
  })
  return http.ListenAndServe(m.addr, h)
}

// URL returns full URL for a specific source path
func (m *SourceMock) URL(path string) string {
  return "http://" + m.addr + "/" + path
}
