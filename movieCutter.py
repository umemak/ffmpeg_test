#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
動画から音声のある部分だけ切り出す
"""
import sys
from movieHelper import MovieHelper

if __name__ == '__main__':  # インポート時には動かない
    # 引数チェック
    if 2 == len(sys.argv):
        # Pythonに以下の2つ引数を渡す想定
        # 0は固定でスクリプト名
        # 1.対象のファイルパス
        target_file_path = sys.argv[1]
    else:
        print('引数が不正です。')
        sys.exit()

    mh = MovieHelper(target_file_path)
    movie_list = mh.movie_dividing()
