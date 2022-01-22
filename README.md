# knafeh
command line wrapper for webm encoding with ffmpeg

## Features:
- VP8/VP9/AV1/Opus/Vorbis/2-Pass/CRF support
- Industry-grade codec settings
- Simple interface
- Filters
    - Resize
    - Trim
    - Crop
    - Dub
    - Deinterlace
    - Denoise

## Build
```console
$ make build
```

## Usage
```console
$ ./knafeh --help
Usage: ./knafeh -i in.mp4 out.webm
  -an
        removes audio from the video
  -b:a int
        bitrate of the audio in kbps (default 96)
  -c:v string
        which video codec to use i.e. "vp8/vp9/av1" (default "vp9")
  -crf int
        quality of the video from 0 (best) to 63 (worst) (default 40)
  -crop string
        crops the video in the format "x:y:width:height"
  -deinterlace
        deinterlaces the video
  -denoise
        denoises the video
  -dub string
        filepath to the dubbed file
  -i string
        input filepath
  -loop
        if the dubbed audio is shorter than the video or vice versa this will loop the streams to achieve the full length
  -r float
        framerate of the video "-1" means unset (default -1)
  -scale string
        resizes the video, specified as "width:height"
  -shortest
        stops the output at the shortest video/audio stream (when dubbing)
  -sp
        use single pass encoding, output quality is lower but is quicker to encode
  -ss string
        when to trim the video, accepts "HH:MM:SS.MS/HH:MM:SS/S"
  -title string
        metadata title of the video
  -to string
        when to stop trimming the video, accepts "HH:MM:SS.MS/HH:MM:SS/S"

$ ./knafeh -i in.mp4 -c:v vp8 -b:a 96 -ss 5 -to 6 out.webm
```

## License
```
BSD-3-Clause
```