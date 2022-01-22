package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/fiwippi/knafeh/pkg/ffmpeg"
)

func ParseFlags() (*ffmpeg.Inputs, error) {
	// Input/Output
	input := flag.String("i", "", "input filepath")
	// Passes
	singlePass := flag.Bool("sp", false, "use single pass encoding, output quality is lower but is quicker to encode")
	// Metadata
	title := flag.String("title", "", "metadata title of the video")
	// Video
	codec := flag.String("c:v", "vp9", "which video codec to use i.e. \"vp8/vp9/av1\"")
	crf := flag.Int("crf", 40, "quality of the video from 0 (best) to 63 (worst)")
	framerate := flag.Float64("r", -1, "framerate of the video \"-1\" means unset")
	// Audio
	audioBitrate := flag.Int("b:a", 96, "bitrate of the audio in kbps")
	noAudio := flag.Bool("an", false, "removes audio from the video")
	// Filters
	denoise := flag.Bool("denoise", false, "denoises the video")
	deinterlace := flag.Bool("deinterlace", false, "deinterlaces the video")
	scale := flag.String("scale", "", "resizes the video, specified as \"width:height\"")
	trimStart := flag.String("ss", "", "when to trim the video, accepts \"HH:MM:SS.MS/HH:MM:SS/S\"")
	trimEnd := flag.String("to", "", "when to stop trimming the video, accepts \"HH:MM:SS.MS/HH:MM:SS/S\"")
	dubFp := flag.String("dub", "", "filepath to the dubbed file")
	dubLoop := flag.Bool("loop", false, "if the dubbed audio is shorter than the video or vice versa this will loop the streams to achieve the full length")
	dubShortest := flag.Bool("shortest", false, "stops the output at the shortest video/audio stream (when dubbing)")
	crop := flag.String("crop", "", "crops the video in the format \"x:y:width:height\"")

	// Validate the input and output flags exist
	flag.Usage = func() {
		fmt.Printf("Usage: ./knafeh -i in.mp4 out.webm\n")
		flag.PrintDefaults()
	}

	var output string
	if flag.Parse(); len(flag.Args()) > 0 {
		output = flag.Args()[0]
	} else {
		flag.Usage()
		os.Exit(1)
	}

	fd, err := ffmpeg.Probe(*input)
	if err != nil {
		return nil, err
	}

	// Create the inputs
	i := ffmpeg.NewInputs()
	i.InputFp = *input
	i.OutputFp = output
	i.VarArgs.Tolerance = 2
	i.Threads = runtime.NumCPU()

	// Video args
	err = i.ParseCodec(*codec)
	if err != nil {
		return nil, err
	}
	err = i.ParseCRF(*crf)
	if err != nil {
		return nil, err
	}
	i.Width = fd.Width
	i.Height = fd.Height

	// Audio args
	err = i.ParseAudioBitrate(*audioBitrate)
	if err != nil {
		return nil, err
	}
	i.AudioEnabled = !(*noAudio)
	if len(fd.AudioStreams) > 0 {
		i.AudioTrack.Index = 0
		i.AudioTrack.Title = fd.AudioStreams[0].Tags.Title
	}

	// Miscellaneous
	i.Title = fd.Title
	if *title != "" {
		i.Title = *title
	}
	i.Framerate = *framerate
	i.RowMultithreading = true

	// Filter args
	if *denoise {
		i.Denoise = &ffmpeg.DenoiseFilter{}
	}
	if *deinterlace {
		i.Deinterlace = &ffmpeg.DeinterlaceFilter{}
	}
	if *scale != "" {
		err = i.ParseResize(*scale)
		if err != nil {
			return nil, err
		}
	} else {
		i.Resize = nil
	}
	i.Trim.Start = *trimStart
	i.Trim.End = *trimEnd
	if *dubFp != "" {
		dfd, err := ffmpeg.Probe(*dubFp)
		if err != nil {
			return nil, err
		}
		i.Dub.Filepath = *dubFp
		i.Dub.Shortest = *dubShortest
		i.Dub.Loop = *dubLoop
		i.Dub.VideoDuration = fd.DurationSeconds
		i.Dub.AudioDuration = dfd.DurationSeconds
	} else {
		i.Dub = nil
	}
	if *crop != "" {
		err = i.ParseCrop(*crop)
		if err != nil {
			return nil, err
		}
	} else {
		i.Crop = nil
	}
	i.TwoPass = !(*singlePass)

	return i, nil
}
