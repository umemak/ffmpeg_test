FROM ubuntu

RUN apt update
RUN apt install -y ffmpeg
# RUN apt install -y fonts-ipafont
# RUN fc-cache -fv
RUN apt install -y wget unzip
RUN wget https://github.com/googlefonts/morisawa-biz-ud-gothic/releases/download/v1.05/morisawa-biz-ud-gothic-fonts.zip
RUN unzip morisawa-biz-ud-gothic-fonts.zip
RUN cp morisawa-biz-ud-gothic-fonts/fonts/ttf/BIZUDPGothic-Bold.ttf /usr/share/fonts/
