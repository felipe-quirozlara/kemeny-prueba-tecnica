package llm

import "io"

// bytesReader returns an io.Reader from bytes. Shared helper used by HTTP clients.
func bytesReader(b []byte) *reader { return &reader{b: b} }

type reader struct{ b []byte; i int }
func (r *reader) Read(p []byte) (n int, err error) {
    if r.i >= len(r.b) { return 0, io.EOF }
    n = copy(p, r.b[r.i:])
    r.i += n
    return n, nil
}
