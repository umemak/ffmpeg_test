FROM python:3.9-buster

RUN apt update -y && \
    apt install -y ffmpeg wget unzip
RUN pip install \
    ffmpeg-python \
    moviepy \
    numpy \
    pydub \
    soundfile \
    xmltodict

RUN wget https://go.dev/dl/go1.19.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz && \
    rm go1.19.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# RUN apt install -y fonts-ipafont
# RUN fc-cache -fv
RUN wget https://github.com/googlefonts/morisawa-biz-ud-gothic/releases/download/v1.05/morisawa-biz-ud-gothic-fonts.zip && \
    unzip morisawa-biz-ud-gothic-fonts.zip && \
    rm morisawa-biz-ud-gothic-fonts.zip && \
    cp morisawa-biz-ud-gothic-fonts/fonts/ttf/BIZUDPGothic-Bold.ttf /usr/share/fonts/
