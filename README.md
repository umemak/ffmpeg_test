# ffmpeg_test

## docker build

```sh
docker build -t ffmpeg:local .
```

## 動画に字幕を入れる

```sh
docker run -it -v `pwd`:/work ffmpeg:local bash
cd /work
ffmpeg -i D0002160514_00000.mp4 -vf subtitles=test.srt out.mp4
```

## 動画を無音の開始終了部分で切る

```sh
docker run -it -v `pwd`:/work ffmpeg:local bash
cd /work
python movieCutter.py D0002100130_00000.mp4 > D0002100130_00000.mp4.txt
go run main.go D0002100130_00000.mp4 > D0002100130_00000.mlt
```

## 動画の無音部分を削除する

```sh
docker run -it -v `pwd`:/work ffmpeg:local bash
cd /work
python movieCutter.py D0002100130_00000.mp4 > D0002100130_00000.mp4.txt
go run main.go D0002100130_00000.mp4 del > D0002100130_00000.mlt
```

## 参考

- [Pythonとffmpegで動画の無音部分をカットする - Qiita](https://qiita.com/igapon1/items/3faa83fc8af1543bc672)
