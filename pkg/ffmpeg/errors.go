package ffmpeg

import "errors"

var (
	ErrThreadNum    = errors.New("invalid number of threads")
	ErrInvalidCodec = errors.New("invalid codec specified")
	ErrInvalidCRF   = errors.New("crf is not between 0 and 63")
	ErrFramerate    = errors.New("framerate is too low")
	ErrResize       = errors.New("invalid resize resolution")
	ErrCrop         = errors.New("invalid crop dimensions")
	ErrDub          = errors.New("invalid dub")
	ErrNegTrimDur   = errors.New("trim duration is negative")
	ErrAudioBitrate = errors.New("audio bitrate is too low")
)
