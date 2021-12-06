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
```sh
% ./knafeh --help
Usage:
  knafeh [OPTIONS] [output]

Application Options:
  -i, --input=       input filepath
      --sp           use single pass encoding, output quality is lower but is quicker to encode
  -t, --title=       metadata title of the video
  -c, --codec=       which video codec to use i.e. "vp8/vp9/av1" (default: vp9)
      --crf=         quality of the video from 0 (best) to 63 (worst) (default: 40)
  -r, --framerate=   framerate of the video "-1" means unset (default: -1)
      --b:a=         bitrate of the audio in kbps (default: 96)
      --an           removes audio from the video
      --denoise      denoises the video
      --deinterlace  deinterlaces the video
      --resize=      resizes the video, specified as "width:height"
      --ss=          when to trim the video, accepts "HH:MM:SS.MS/HH:MM:SS/S"
      --to=          when to stop trimming the video, accepts "HH:MM:SS.MS/HH:MM:SS/S"
      --dubfp=       filepath to the dubbed file
      --loop         if the dubbed audio is shorter than the video or vice versa this will loop the streams to achieve
                     the full length
      --shortest     stops the output at the shortest video/audio stream (when dubbing)
      --crop=        crops the video in the format "x:y:width:height"

Help Options:
  -h, --help         Show this help message

Arguments:
  output:            output filepath

% ./knafeh -i in.mp4  -c vp8 --ss 5 --to 10 out.webm
```

## License
```
BSD-3-Clause
```
