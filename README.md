# ffmpeg_test

```sh
docker build -t ffmpeg:local .
docker run -it -v `pwd`:/work ffmpeg:local bash
cd /work
ffmpeg -i D0002160514_00000.mp4 -vf subtitles=test.srt out.mp4
```

## 参考

- [Pythonとffmpegで動画の無音部分をカットする - Qiita](https://qiita.com/igapon1/items/3faa83fc8af1543bc672)
