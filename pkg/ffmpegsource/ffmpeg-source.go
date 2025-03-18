package ffmpegsource

import (
	"fmt"
	"os/exec"
)

type HTTPHeaders map[string]string

type Params []string

type FFMPEGSource struct {
	HTTPHeaders map[string]string
	SourceUrl   string
	ExtraParams []string
}

func NewFFMPEGSource(url string) FFMPEGSource {
	return FFMPEGSource{
		HTTPHeaders: make(map[string]string),
		SourceUrl:   url,
		ExtraParams: make([]string, 0),
	}
}

func (s FFMPEGSource) Arguments() []string {
	arguments := make([]string, 0)

	for header, value := range s.HTTPHeaders {
		arguments = append(arguments, "-headers")
		arguments = append(arguments, fmt.Sprintf("%s: %s", header, value))
	}

	arguments = append(arguments, "-i", s.SourceUrl)
	arguments = append(arguments, s.ExtraParams...)

	return arguments
}

func (s FFMPEGSource) Start(*exec.Cmd) error {
	return nil
}

func (s FFMPEGSource) Done() {
}
