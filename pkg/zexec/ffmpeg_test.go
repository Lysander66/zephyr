package zexec

import (
	"testing"
)

func TestFFmpegFilters_Export(t *testing.T) {
	const (
		i  = "rtmp://test-streams.dev/live/test"
		o  = "rtmp://localhost:1935/live/test"
		w1 = "ffmpeg -hwaccel cuda -i 'rtmp://test-streams.dev/live/test' -vf 'delogo=x=1:y=640:w=270:h=79:show=0' -c:v h264_nvenc -preset fast -r 30 -g 60 -c:a copy -f flv 'rtmp://localhost:1935/live/test'"
		w2 = `ffmpeg -headers $'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36\r\nReferer: https://test.live/\r\n' -i 'rtmp://test-streams.dev/live/test' -c copy -f flv 'rtmp://localhost:1935/live/test'`
	)

	case1 := &FFmpegFilters{}
	case1.Input(i).Output(o).
		AdvancedOption("-hwaccel", "cuda").
		Option("-vf", "delogo=x=1:y=640:w=270:h=79:show=0").
		Option("-c:v", "h264_nvenc").
		Option("-preset", "fast").
		Option("-r", "30").
		Option("-g", "60").
		Option("-c:a", "copy").
		Option("-f", "flv")

	case2 := &FFmpegFilters{}
	case2.Input(i).Output(o).
		AddHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36").
		AddHeader("Referer", "https://test.live/").
		Option("-c", "copy").
		Option("-f", "flv")

	tests := []struct {
		name        string
		fields      *FFmpegFilters
		wantCommand string
	}{
		{
			name:        "case 1",
			fields:      case1,
			wantCommand: w1,
		},
		{
			name:        "case 2",
			fields:      case2,
			wantCommand: w2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCommand := tt.fields.Export(); gotCommand != tt.wantCommand {
				t.Errorf("Export() = %v, want %v", gotCommand, tt.wantCommand)
			}
		})
	}
}
