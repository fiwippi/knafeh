package main

import (
	"runtime"

	"github.com/fiwippi/knafeh/pkg/ffmpeg"
)

type options struct {
	// Input/Output
	Input   string `short:"i" long:"input" description:"input filepath" required:"true"`
	PosArgs struct {
		Output string `positional-arg-name:"output" description:"output filepath" required:"true"`
	} `positional-args:"true"`

	// Encoding
	SinglePass bool `long:"sp" description:"use single pass encoding, output quality is lower but is quicker to encode"`

	// Metadata
	Title string `short:"t" long:"title" description:"metadata title of the video"`

	// Video
	Codec     string  `short:"c" long:"codec" description:"which video codec to use i.e. \"vp8/vp9/av1\"" default:"vp9"`
	CRF       int     `long:"crf" description:"quality of the video from 0 (best) to 63 (worst)" default:"40"`
	Framerate float64 `short:"r" long:"framerate" description:"framerate of the video \"-1\" means unset" default:"-1"`

	// Audio
	AudioBitrate int  `long:"b:a" description:"bitrate of the audio in kbps" default:"96"`
	NoAudio      bool `long:"an" description:"removes audio from the video"`

	// Filters
	Denoise     bool   `long:"denoise" description:"denoises the video"`
	Deinterlace bool   `long:"deinterlace" description:"deinterlaces the video"`
	Resize      string `long:"resize" description:"resizes the video, specified as \"width:height\""`
	TrimStart   string `long:"ss" description:"when to trim the video, accepts \"HH:MM:SS.MS/HH:MM:SS/S\""`
	TrimEnd     string `long:"to" description:"when to stop trimming the video, accepts \"HH:MM:SS.MS/HH:MM:SS/S\""`
	DubFp       string `long:"dubfp" description:"filepath to the dubbed file"`
	DubLoop     bool   `long:"loop" description:"if the dubbed audio is shorter than the video or vice versa this will loop the streams to achieve the full length"`
	DubShortest bool   `long:"shortest" description:"stops the output at the shortest video/audio stream (when dubbing)"`
	Crop        string `long:"crop" description:"crops the video in the format \"x:y:width:height\""`
}

func ParseOpts(opts options) (*ffmpeg.Inputs, error) {
	// ffprobe the file
	fd, err := ffmpeg.Probe(opts.Input)
	if err != nil {
		return nil, err
	}

	// Start creating the inputs
	i := ffmpeg.NewInputs()
	i.InputFp = opts.Input
	i.OutputFp = opts.PosArgs.Output
	i.VarArgs.Tolerance = 2
	i.Threads = runtime.NumCPU()

	// Video args
	err = i.ParseCodec(opts.Codec)
	if err != nil {
		return nil, err
	}
	err = i.ParseCRF(opts.CRF)
	if err != nil {
		return nil, err
	}
	i.Width = fd.Width
	i.Height = fd.Height

	// Audio args
	err = i.ParseAudioBitrate(opts.AudioBitrate)
	if err != nil {
		return nil, err
	}
	i.AudioEnabled = !(opts.NoAudio)
	if len(fd.AudioStreams) > 0 {
		i.AudioTrack.Index = 0
		i.AudioTrack.Title = fd.AudioStreams[0].Tags.Title
	}

	// Miscellaneous
	i.Title = fd.Title
	if opts.Title != "" {
		i.Title = opts.Title
	}
	i.Framerate = opts.Framerate
	i.RowMultithreading = true

	// Filter args
	if opts.Denoise {
		i.Denoise = &ffmpeg.DenoiseFilter{}
	}
	if opts.Deinterlace {
		i.Deinterlace = &ffmpeg.DeinterlaceFilter{}
	}
	if opts.Resize != "" {
		err = i.ParseResize(opts.Resize)
		if err != nil {
			return nil, err
		}
	} else {
		i.Resize = nil
	}
	i.Trim.Start = opts.TrimStart
	i.Trim.End = opts.TrimEnd
	if opts.DubFp != "" {
		dfd, err := ffmpeg.Probe(opts.DubFp)
		if err != nil {
			return nil, err
		}
		i.Dub.Filepath = opts.DubFp
		i.Dub.Shortest = opts.DubShortest
		i.Dub.Loop = opts.DubLoop
		i.Dub.VideoDuration = fd.DurationSeconds
		i.Dub.AudioDuration = dfd.DurationSeconds
	} else {
		i.Dub = nil
	}
	if opts.Crop != "" {
		err = i.ParseCrop(opts.Crop)
		if err != nil {
			return nil, err
		}
	} else {
		i.Crop = nil
	}
	i.TwoPass = !opts.SinglePass

	return i, nil
}
