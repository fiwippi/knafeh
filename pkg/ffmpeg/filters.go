package ffmpeg

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// TrimFilter trims a video with a start and end time
type TrimFilter struct {
	Start, End    string  // Time in HH:MM:SS.MS, HH:MM:SS or S
	VideoDuration float64 // Needed if dubbing
}

func NewTrimFilter() *TrimFilter {
	return &TrimFilter{
		Start: "",
		End:   "",
	}
}

func (tf *TrimFilter) ValidStart() bool {
	return tf.Start != ""
}

func (tf *TrimFilter) ValidEnd() bool {
	return tf.End != ""
}

func (tf *TrimFilter) FilterArg() string {
	var arg string

	if tf.ValidStart() {
		arg += fmt.Sprintf("start=%s", strings.ReplaceAll(tf.Start, ":", "\\\\:"))
		//arg += fmt.Sprintf("start='%s'", strings.ReplaceAll(tf.Start, ":", "\\:"))
	}
	if tf.ValidEnd() {
		arg += fmt.Sprintf(":end=%s", strings.ReplaceAll(tf.End, ":", "\\\\:"))
		//arg += fmt.Sprintf(":end='%s'", strings.ReplaceAll(tf.End, ":", "\\:"))
	}

	return strings.TrimPrefix(arg, ":")
}

func (tf *TrimFilter) StartDuration() (time.Duration, error) {
	start, err := parseTrimTime(tf.Start)
	if err != nil {
		return 0, err
	}
	return start, nil
}

func (tf *TrimFilter) Duration() (time.Duration, error) {
	var err error
	var start, end time.Duration

	// If both start and end are supplied then subtract the two
	if tf.ValidStart() && tf.ValidEnd() {
		start, err = parseTrimTime(tf.Start)
		if err != nil {
			return 0, err
		}

		end, err = parseTrimTime(tf.End)
		if err != nil {
			return 0, err
		}

		return end - start, nil
	}

	// If only end is supplied then it is the new duration
	if tf.ValidEnd() {
		end, err = parseTrimTime(tf.End)
		if err != nil {
			return 0, err
		}

		return end, nil
	}

	// If only start is supplied then subtract it from the duration
	if tf.ValidStart() {
		start, err = parseTrimTime(tf.Start)
		if err != nil {
			return 0, err
		}

		dur, err := time.ParseDuration(fmt.Sprintf("%ds", int(tf.VideoDuration)))
		if err != nil {
			return 0, err
		}

		return dur - start, nil
	}

	return 0, errors.New("shouldn't be here")
}

// ResizeFilter resizes a video into a new width and height
type ResizeFilter struct {
	Width, Height int // Width and Height in pixels
}

func NewResizeFilter() *ResizeFilter {
	return &ResizeFilter{
		Width:  0,
		Height: 0,
	}
}

func (rf *ResizeFilter) ValidResolution() bool {
	return true // We allow -1 etc.
	//return rf.Width > 0 && rf.Height > 0
}

func (rf *ResizeFilter) Args() (string, string) {
	return "scale", fmt.Sprintf("%d:%d:flags=lanczos", rf.Width, rf.Height)
}

// CropFilter crops a video onto a specified crop window
type CropFilter struct {
	X, Y int // Coordinates where the origin (top-left) of the crop window should start
	W, H int // Width and Height of the crop window
}

func NewCropFilter() *CropFilter {
	return &CropFilter{
		X: 0,
		Y: 0,
		W: 0,
		H: 0,
	}
}

func (cf *CropFilter) ValidCrop() bool {
	return cf.X >= 0 && cf.Y >= 0 && cf.W > 0 && cf.H > 0
}

func (cf *CropFilter) Args() (string, string) {
	return "crop", fmt.Sprintf("%d:%d:%d:%d", cf.W, cf.H, cf.X, cf.Y)
}

// DeinterlaceFilter deinterlaces the video
type DeinterlaceFilter struct{}

func (df *DeinterlaceFilter) Valid() bool {
	return true
}

func (df *DeinterlaceFilter) Args() (string, string) {
	return "yadif", "0:-1:0"
}

// DenoiseFilter denoises the video
type DenoiseFilter struct{}

func (df *DenoiseFilter) Valid() bool {
	return true
}

func (df *DenoiseFilter) Args() (string, string) {
	return "hqdn3d", "4.0:3.0:6.0:4.5"
}

// DubLoopMode If looping with the dub filter, this
// specifies whether the audio or video should be looped
type DubLoopMode int

const (
	None DubLoopMode = iota
	Video
	Audio
)

// DubFilter dubs a video with audio from a file
type DubFilter struct {
	Filepath      string  // Filepath to the file
	VideoDuration float64 // Duration of the original video
	AudioDuration float64 // Duration of the dubbed audio
	Loop          bool    // Loop the video/audio to achieve the full video length
	Shortest      bool    // Include -shortest, automatically applied if looping
}

func NewDubFilter() *DubFilter {
	return &DubFilter{
		Filepath:      "",
		VideoDuration: -1,
		AudioDuration: -1,
		Loop:          false,
		Shortest:      false,
	}
}

func (df *DubFilter) Valid() bool {
	return df.Filepath != "" && df.AudioDuration > 0 && df.VideoDuration >= 0
}

func (df *DubFilter) LoopMode() DubLoopMode {
	if df.VideoDuration > df.AudioDuration {
		return Audio
	} else if df.AudioDuration > df.VideoDuration {
		return Video
	}
	return None
}

func (df *DubFilter) ArgLoop() (string, string) {
	if df.LoopMode() == Audio {
		return "aloop", "-1:2147483647:0"
	} else if df.LoopMode() == Video {
		return "loop", "-1:32767:0"
	}
	return "", ""
}

func (df *DubFilter) ArgShortest() (string, string) {
	return "-shortest", ""
}

func (df *DubFilter) ArgFilepath() string {
	return df.Filepath
}
