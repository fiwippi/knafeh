package ffmpeg

type AudioTrack struct {
	Index int
	Title string
}

func NewAudioTrack() AudioTrack {
	return AudioTrack{
		Index: -1,
		Title: "",
	}
}
