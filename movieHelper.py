#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys  # 終了時のエラー有無
import os  # ファイルパス分解
import datetime
import glob
from dataclasses import dataclass

import math
import numpy as np
from pydub import AudioSegment
import subprocess
import soundfile as sf
import ffmpeg
import hashlib

@dataclass(frozen=True)
class MovieValue:
    """
    動画の値オブジェクト
    """
    target_filepath: 'str 対象のファイルパス'
    target_basename: 'str 対象のファイル名+拡張子'
    target_dirname: 'str 対象のディレクトリ'
    target_filename: 'str 対象のファイル名'
    target_ext: 'str 対象の拡張子'
    video_info: 'Any 動画の情報'

    def __init__(self, target_filepath):
        """
        コンストラクタ

        :param target_filepath: str movieのファイルパス
        """
        if target_filepath is None:
            print('動画ファイルパスがNoneです')
            sys.exit()
        if not os.path.isfile(target_filepath):
            print(target_filepath, '動画ファイルが存在しません', sep=':')
            sys.exit()
        object.__setattr__(self, "target_filepath", target_filepath)
        target_basename = os.path.basename(target_filepath)
        object.__setattr__(self, "target_basename", target_basename)
        target_dirname = os.path.dirname(target_filepath)
        object.__setattr__(self, "target_dirname", target_dirname)
        target_filename = os.path.splitext(target_basename)[0]
        object.__setattr__(self, "target_filename", target_filename)
        target_ext = os.path.splitext(target_basename)[1]
        object.__setattr__(self, "target_ext", target_ext)
        object.__setattr__(self, "video_info", ffmpeg.probe(target_filepath))


class MovieHelper:
    """
    動画ファイルのヘルパー
    """
    movie_value: 'MovieValue movieの値オブジェクト'
    movie_filepath: 'str 動画ファイル入力パス'
    wave_filepath: 'str 音声ファイル出力パス'
    movie_dividing_filepath: 'list 分割動画ファイル出力パスリスト'

    def __init__(self, movie_value):
        """
        コンストラクタ

        :param movie_value: str 動画のファイルパス、または、MovieValue 動画の値オブジェクト
        """
        if movie_value is None:
            print('引数movie_valueがNoneです')
            sys.exit()
        if isinstance(movie_value, str):
            movie_value = MovieValue(movie_value)
        self.movie_value = movie_value
        self.movie_filepath = movie_value.target_filepath
        self.movie_dividing_filepath = []
        self.wave_filepath = os.path.join(self.movie_value.target_dirname,
                                          self.movie_value.target_filename + '.wav',
                                          )

    def movie_dividing(self,
                       threshold=0.05,
                       min_silence_duration=0.5,
                       padding_time=0.1,
                       ):
        """
        動画ファイルから無音部分をカットした部分動画ファイル群を作成する

        :param threshold: float 閾値
        :param min_silence_duration: float [秒]以上thresholdを下回っている個所を抽出する
        :param padding_time: float [秒]カットしない無音部分の長さ
        :return: list[str] 分割したファイルのパスリスト
        """
        if len(self.movie_dividing_filepath) > 0:
            print('すでに動画ファイルから無音部分をカットした部分動画ファイル群を作成済みです')
            sys.exit()
        command_output = ["ffmpeg",
                          "-i",
                          self.movie_filepath,
                          "-ac",
                          "1",
                          "-ar",
                          "44100",
                          "-acodec",
                          "pcm_s16le",
                          self.wave_filepath,
                          ]
        subprocess.run(command_output)

        # 音声ファイル読込
        data, frequency = sf.read(self.wave_filepath)  # file:音声ファイルのパス
        # 一定のレベル(振幅)以上の周波数にフラグを立てる
        amp = np.abs(data)
        list_over_threshold = amp > threshold

        # 一定時間以上、小音量が続く箇所を探す
        silences = []
        prev = 0
        entered = 0
        for i, v in enumerate(list_over_threshold):
            if prev == 1 and v == 0:  # enter silence
                entered = i
            if prev == 0 and v == 1:  # exit silence
                duration = (i - entered) / frequency
                if duration > min_silence_duration:
                    silences.append({"from": entered, "to": i, "suffix": "cut"})
                    entered = 0
            prev = v
        if 0 < entered < len(list_over_threshold):
            silences.append({"from": entered, "to": len(list_over_threshold), "suffix": "cut"})

        list_block = silences  # 無音部分のリスト：[{"from": 始点, "to": 終点}, {"from": ...}, ...]
        cut_blocks = [list_block[0]]
        for i, v in enumerate(list_block):
            if i == 0:
                continue
            moment = (v["from"] - cut_blocks[-1]["to"]) / frequency
            # カット対象だった場合
            if 0.3 > moment:
                cut_blocks[-1]["to"] = v["to"]  # １つ前のtoを書き換え
            # カット対象でない場合
            else:
                cut_blocks.append(v)  # そのまま追加

        # カットする箇所を反転させて、残す箇所を決める
        keep_blocks = []
        for i, block in enumerate(cut_blocks):
            if i == 0 and block["from"] > 0:
                keep_blocks.append({"from": 0, "to": block["from"], "suffix": "keep"})
            if i > 0:
                prev = cut_blocks[i - 1]
                keep_blocks.append({"from": prev["to"], "to": block["from"], "suffix": "keep"})
            if i == len(cut_blocks) - 1 and block["to"] < len(data):
                keep_blocks.append({"from": block["to"], "to": len(data), "suffix": "keep"})

        # list_keep 残す動画部分のリスト：[{"from": 始点, "to": 終点}, {"from": ...}, ...]
        for i, block in enumerate(keep_blocks):
            fr = max(block["from"] / frequency - padding_time, 0)
            to = min(block["to"] / frequency + padding_time, len(data) / frequency)
            print(fr)
            print(to)
        return self.movie_dividing_filepath
