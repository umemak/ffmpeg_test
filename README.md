# ffmpeg_test

```sh
docker build -t ffmpeg:local .
docker run -it -v `pwd`:/work ffmpeg:local bash
cd /work
ffmpeg -i D0002160514_00000.mp4 -vf subtitles=test.srt out.mp4
```
