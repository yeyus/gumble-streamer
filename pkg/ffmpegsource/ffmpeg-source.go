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

	if len(s.HTTPHeaders) > 0 {
		arguments = append(arguments, "-headers")
		headers := ""
		for header, value := range s.HTTPHeaders {
			headers += fmt.Sprintf("%s: %s\r\n", header, value)
		}
		arguments = append(arguments, headers)
	}

	arguments = append(arguments, "-i", s.SourceUrl)
	arguments = append(arguments, s.ExtraParams...)

	fmt.Printf("[ffmpeg-source] arguments: %v\n", arguments)

	return arguments
}

func (s FFMPEGSource) Start(*exec.Cmd) error {
	return nil
}

func (s FFMPEGSource) Done() {
}
