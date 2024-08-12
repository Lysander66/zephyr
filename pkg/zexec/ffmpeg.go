package zexec

import (
	"fmt"

	"github.com/Lysander66/zephyr/pkg/z"
)

// https://www.ffmpeg.org/ffmpeg.html

type FFmpegFilters struct {
	input           []string
	output          []string
	headers         []z.Pair[string, string]
	mainOptions     []z.Pair[string, string]
	advancedOptions []z.Pair[string, string]
}

func (f *FFmpegFilters) Input(input string) *FFmpegFilters {
	f.input = append(f.input, input)
	return f
}

func (f *FFmpegFilters) Output(output string) *FFmpegFilters {
	f.output = append(f.output, output)
	return f
}

func (f *FFmpegFilters) AddHeader(key, value string) *FFmpegFilters {
	f.headers = append(f.headers, z.MakePair(key, value))
	return f
}

func (f *FFmpegFilters) Option(key, value string) *FFmpegFilters {
	f.mainOptions = append(f.mainOptions, z.MakePair(key, value))
	return f
}

func (f *FFmpegFilters) AdvancedOption(key, value string) *FFmpegFilters {
	f.advancedOptions = append(f.advancedOptions, z.MakePair(key, value))
	return f
}

func (f *FFmpegFilters) Export() string {
	args := "ffmpeg"

	if len(f.headers) > 0 {
		args += " -headers $'"
		for _, header := range f.headers {
			args += fmt.Sprintf(`%s: %s\r\n`, header.First, header.Second)
		}
		args += "'"
	}

	for _, opt := range f.advancedOptions {
		args += fmt.Sprintf(" %s %s", opt.First, opt.Second)
	}

	for _, i := range f.input {
		args += fmt.Sprintf(" -i '%s'", i)
	}

	for _, opt := range f.mainOptions {
		switch opt.First {
		case "-vf":
			args += fmt.Sprintf(" %s '%s'", opt.First, opt.Second)
		default:
			args += fmt.Sprintf(" %s %s", opt.First, opt.Second)
		}
	}

	for _, o := range f.output {
		args += fmt.Sprintf(" '%s'", o)
	}

	return args
}
