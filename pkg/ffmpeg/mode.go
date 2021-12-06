package ffmpeg

import (
	"fmt"
	"math"
	"strconv"
)

type VariableArgs struct {
	codec     Codec // Codec used
	CRF       int   // CRF of the video
	Tolerance int   // Tolerance for the CRF
	// Specify when using VP8!
	AudioQualityScale int // Quality scale for the audio (-1 to 10)
	// Specify when using VP9, AV1!
	AudioBitrate int // Bitrate of the audio in Kbps
}

func NewVariableArgs() *VariableArgs {
	return &VariableArgs{
		CRF:               -1,
		Tolerance:         -1,
		AudioQualityScale: -1,
		AudioBitrate:      -1,
	}
}

func (va *VariableArgs) Valid() (bool, error) {
	if _, ac := va.codec.ArgAudioCodec(); ac == "libopus" && va.AudioBitrate < 6 {
		return false, ErrAudioBitrate
	}

	return true, nil
}

func (va *VariableArgs) ArgAudioQuality() (string, string) {
	if va.codec == VP8 {
		return "-qscale:a", strconv.Itoa(va.AudioQualityScale)
	} else {
		return "-b:a", fmt.Sprintf("%dk", va.AudioBitrate)
	}
}

func (va *VariableArgs) ArgVideoArgs() [][]string {
	qMin := int(math.Max(0, float64(va.CRF-va.Tolerance)))
	qMax := int(math.Min(63, float64(va.CRF+va.Tolerance)))

	args := [][]string{
		{"-qmin", strconv.Itoa(qMin)},
		{"-crf", strconv.Itoa(va.CRF)},
		{"-qmax", strconv.Itoa(qMax)},
		{"-qcomp", "1"},
		{"-b:v", "0"},
	}

	return args
}
