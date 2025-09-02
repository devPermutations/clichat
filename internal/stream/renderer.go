package stream

import (
	"io"
	"os"
)

type Renderer struct {
	w io.Writer
}

func NewRenderer() *Renderer { return &Renderer{w: os.Stdout} }

func (r *Renderer) WriteToken(token string) error {
	_, err := io.WriteString(r.w, token)
	return err
}
