package ffmpeg

import (
	"fmt"
	"strconv"
)

// Resources:
// https://ffmpeg.org/ffmpeg-all.html#libvpx
// https://ffmpeg.org/ffmpeg-all.html#libaom_002dav1
// https://github.com/Kagami/webm.py/wiki
// https://groups.google.com/a/webmproject.org/g/codec-devel/c/oiHjgEdii2U
// https://www.webmproject.org/docs/encoder-parameters/
// http://wiki.webmproject.org/ffmpeg
// https://trac.ffmpeg.org/wiki/Encode/VP9
// http://underpop.online.fr/f/ffmpeg/help/libaom_002dav1.htm.gz

// Inputs collects variables to create an ffmpeg command.
// If you don't want to include a command then set its
// value to -1 or nil if it's a pointer
type Inputs struct {
	c *Command

	// Input
	InputFp  string
	OutputFp string

	// Options common to VP8, VP9, AV1
	Codec             Codec      // Which codec to use: VP8, VP9, AV1
	AudioTrack        AudioTrack // Index of the audio track to include
	Threads           int        // How many threads to encode with
	Slices            int        // Split the video into how many slices; 1, 2, 4 or 8 slices for VP8; 1, 2, 4, 8, 16, 32, 64 slices for VP9
	AudioEnabled      bool       // Should audio be included in the video
	RowMultithreading bool       // Whether to enable row multithreading, only works for VP9/AV1
	Title             string     // Title of the video in metadata
	Framerate         float64    // Output framerate of the final video
	TwoPass           bool

	// Args for variable encoding
	VarArgs *VariableArgs

	// Filter options
	Dub         *DubFilter
	Crop        *CropFilter
	Trim        *TrimFilter
	Resize      *ResizeFilter
	Denoise     *DenoiseFilter
	Deinterlace *DeinterlaceFilter

	// Dimensions
	Width, Height int
}

func NewInputs() *Inputs {
	return &Inputs{
		Codec:             0,
		AudioTrack:        NewAudioTrack(),
		Threads:           0,
		Slices:            0,
		AudioEnabled:      false,
		RowMultithreading: false,
		Title:             "",
		Framerate:         0,
		VarArgs:           NewVariableArgs(),
		Dub:               NewDubFilter(),
		Crop:              NewCropFilter(),
		Trim:              NewTrimFilter(),
		Resize:            NewResizeFilter(),
		Denoise:           nil,
		Deinterlace:       nil,
		Width:             -1,
		Height:            -1,
		TwoPass:           false,
	}
}

func (i *Inputs) Command() (*Command, error) {
	i.VarArgs.codec = i.Codec

	if err := i.preprocess(); err != nil {
		return nil, err
	}

	i.c = newCommand()

	// Input args
	i.c.twoPass = i.TwoPass
	i.c.inputFp = i.InputFp
	i.c.outputFp = i.OutputFp
	i.processDubInput()

	// General Args
	i.c.addGeneralArg("-metadata", fmt.Sprintf("title=\"%s\"", i.Title))
	i.c.addGeneralArg("-threads", strconv.Itoa(i.Threads))
	i.processFramerate()
	i.c.addGeneralArg("-pix_fmt", "yuv420p")
	i.c.addGeneralArg("-f", "webm")
	i.processDubShortest()

	// Map args
	i.processMapStreams()

	// Video filter args
	i.processTrim()
	i.processCrop()
	i.processDeinterlace()
	i.processDenoise()
	i.processResize()
	i.processDubLoop()

	// Video Args
	i.processVideoCodecAndModeArg()
	i.processSlices()
	i.c.addVideoArgs(i.Codec.ArgRowMT(i.RowMultithreading))
	i.c.addVideoArg("-auto-alt-ref", "1")
	i.c.addVideoArg("-g", "128")
	i.c.addVideoArgs(i.Codec.ArgVideoCodecSpecific())

	// Audio args
	i.processAudioCodec()

	return i.c, nil
}

func (i *Inputs) preprocess() error {
	// Validate general args
	if i.Threads == -1 {
		i.Threads = 1
	}

	if i.Threads < 1 || i.Threads > 16 {
		return ErrThreadNum
	}

	if i.Framerate == 0 {
		return ErrFramerate
	}

	// Validate filter args
	if i.Resize != nil && !i.Resize.ValidResolution() {
		return ErrResize
	}
	if i.Crop != nil && !i.Crop.ValidCrop() {
		return ErrCrop
	}
	if i.Dub != nil && !i.Dub.Valid() {
		return ErrDub
	}

	// If trimming then video duration has changed, if dubbing then we need to update the video duration
	if i.Trim != nil && i.Dub != nil {
		d, err := i.Trim.Duration()
		if err != nil {
			return err
		}
		if d < 0 {
			return ErrNegTrimDur
		}

		i.Dub.VideoDuration = d.Seconds()
	}

	// Validate mode arguments
	if valid, err := i.VarArgs.Valid(); !valid {
		return err
	}

	return nil
}

func (i *Inputs) processDenoise() {
	if i.Denoise != nil && i.Denoise.Valid() {
		i.c.addVideoFilterArg(i.Denoise.Args())
	}
}

func (i *Inputs) processDeinterlace() {
	if i.Deinterlace != nil && i.Deinterlace.Valid() {
		i.c.addVideoFilterArg(i.Deinterlace.Args())
	}
}

func (i *Inputs) processTrim() {
	if i.Trim != nil && (i.Trim.ValidStart() || i.Trim.ValidEnd()) {
		i.c.addVideoFilterArg("trim", i.Trim.FilterArg())
		i.c.addVideoFilterArg("setpts", "PTS-STARTPTS")

		if !i.usingDubFilter() {
			i.c.addAudioFilterArg("atrim", i.Trim.FilterArg())
			i.c.addAudioFilterArg("asetpts", "PTS-STARTPTS")
			i.c.audioFilterInput = 0
		}
	}
}

func (i *Inputs) processResize() {
	if i.Resize != nil && i.Resize.ValidResolution() {
		i.c.addVideoFilterArg(i.Resize.Args())
	}
}

func (i *Inputs) processCrop() {
	if i.Crop != nil && i.Crop.ValidCrop() {
		i.c.addVideoFilterArg(i.Crop.Args())
	}
}

func (i *Inputs) processMapStreams() {
	// If no dubbing then we map the standard streams
	if !i.usingDubFilter() {
		i.c.addMapArgs("0:v:0")
		if i.AudioEnabled && i.AudioTrack.Index > -1 {
			i.c.addMapArgs("0:a:" + strconv.Itoa(i.AudioTrack.Index))
		}

		return
	}

	// Dubbing enabled, if looping we only map one stream
	// because the other stream will be mapped with a
	// video/audio filter
	if i.Dub.LoopMode() == None {
		i.c.addMapArgs("0:v:0")
		i.c.addMapArgs("1:a")
	} else if i.Dub.LoopMode() == Audio {
		i.c.addMapArgs("0:v:0")
	} else if i.Dub.LoopMode() == Video {
		i.c.addMapArgs("1:a")
	}
}

func (i *Inputs) processDubInput() {
	if i.usingDubFilter() {
		i.c.audioFilterInput = 1
		i.c.addInputArgs(i.Dub.Filepath)
	}
}

func (i *Inputs) processDubShortest() {
	if i.usingDubFilter() && (i.Dub.Shortest || i.Dub.LoopMode() != None) {
		i.c.addGeneralArg(i.Dub.ArgShortest())
	}
}

func (i *Inputs) processDubLoop() {
	if i.usingDubFilter() {
		if i.Dub.LoopMode() == Audio {
			i.c.addAudioFilterArg("asetpts", "PTS-STARTPTS")
			i.c.addAudioFilterArg(i.Dub.ArgLoop())
		} else if i.Dub.LoopMode() == Video {
			i.c.addVideoFilterArg(i.Dub.ArgLoop())
		}
	}
}

func (i *Inputs) processFramerate() {
	// Output framerate
	if i.Framerate > -1 {
		i.c.addGeneralArg("-r", strconv.FormatFloat(i.Framerate, 'f', -1, 64))
	}
}

func (i *Inputs) processSlices() {
	w := i.Width
	h := i.Height
	i.Slices = RecommendedSlices(w, h)
	i.c.addVideoArgs(i.Codec.ArgSlices(i.Slices, w, h, i.Threads))
}

func (i *Inputs) processVideoCodecAndModeArg() {
	// Which codec should ffmpeg use
	// Quality: AV1 > VP9 > VP8
	// Speed: VP8 >= VP9 > AV1, you can get comparable VP8/VP9 encoding times with slices and row-mt
	// Support: 4chan=VP8, discord=VP8,VP9
	i.c.addVideoArg(i.Codec.ArgVideoCodec())
	i.c.addVideoArgs(i.VarArgs.ArgVideoArgs())
}

func (i *Inputs) processAudioCodec() {
	// VP8 uses vorbis whereas VP9 and AV1 use the more efficient opus codec

	// Vorbis qscale mapping to bitrate
	// -1: 45,
	//  0: 64,
	//  1: 80,
	//  2: 96,
	//  3: 112,
	//  4: 128,
	//  5: 160,
	//  6: 192,
	//  7: 224,
	//  8: 256,
	//  9: 320,
	//  10: 500,

	if i.AudioEnabled {
		i.c.addAudioArg("-ac", "2") // 2 audio channels
		i.c.addAudioArg(i.Codec.ArgAudioCodec())
		i.c.addAudioArg(i.VarArgs.ArgAudioQuality())
	}
}

func (i *Inputs) usingDubFilter() bool {
	return i.AudioEnabled && i.Dub != nil && i.Dub.Valid()
}
