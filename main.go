package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	flag.Parse()
	err := run(flag.Arg(0))
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func run(fname string) error {
	_, err := os.Stdout.Write(makeMLT2(fname))
	if err != nil {
		return err
	}
	return nil
}

type vInfo struct {
	fname  string
	width  string
	height string
	length string
}

func newVInfo(fname string) vInfo {
	cmd := exec.Command(
		"ffprove", "-hide_banner",
		"-i", fname,
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Stdout: %s\n", stdout.String())
		fmt.Printf("Stderr: %s\n", stderr.String())
	}
	buf := stderr.String()
	lines := strings.Split(buf, "\n")
	vi := vInfo{fname: fname, width: "", height: "", length: ""}
	for _, line := range lines {
		if strings.HasPrefix(line, "  Duration: ") {
			l := strings.TrimLeft(line, " ")
			items := strings.Split(l, " ")
			item := strings.TrimRight(items[1], ",")
			vi.length = item
		}
		if strings.Contains(line, " Video: ") && vi.width == "" {
			items := strings.Split(line, ",")
			item := strings.TrimLeft(items[2], " ")
			wh := strings.Split(item, "x")
			vi.width = wh[0]
			vi.height = wh[1]
		}
	}
	return vi
}

type chain struct {
	num int
	in  string
	out string
}

func makeMLT(fname string) []byte {
	cmd := exec.Command(
		"ffmpeg", "-hide_banner",
		"-i", fname,
		"-af", "silencedetect=noise=-50dB:d=0.3",
		// "-af", "silencedetect",
		"-f", "null",
		"-",
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Stdout: %s\n", stdout.String())
		fmt.Printf("Stderr: %s\n", stderr.String())
	}
	buf := stderr.String()
	lines := strings.Split(buf, "\n")
	chains := []chain{}
	i := 0
	start := "00:00:00.000"
	for _, line := range lines {
		if strings.HasPrefix(line, "[silencedetect ") {
			l := strings.Trim(line, "\r\n")
			items := strings.Split(l, " ")
			t := toTime(items[4])
			chain := chain{num: i, in: start, out: t}
			chains = append(chains, chain)
			start = t
			i = i + 1
		}
	}
	res := makeXML(newVInfo(fname), chains)
	return res
}

func makeMLT2(fname string) []byte {
	buf, _ := os.ReadFile(fname + ".txt")
	lines := strings.Split(string(buf), "\n")
	chains := []chain{}
	i := 0
	start := "00:00:00.000"
	for _, line := range lines {
		items := strings.Split(line, " ")
		// fmt.Println(items)
		if len(items) != 2 {
			continue
		}
		f := toTime(items[0])
		t := toTime(items[1])
		if i == 0 && f != "00:00:00.000" {
			chain := chain{num: i, in: start, out: f}
			chains = append(chains, chain)
			start = t
			i = i + 1
			continue
		}
		chains = append(chains, chain{num: i, in: start, out: f})
		i = i + 1
		chains = append(chains, chain{num: i, in: f, out: t})
		start = t
		i = i + 1
	}
	res := makeXML(newVInfo(fname), chains)
	return res
}

func toTime(str string) string {
	f, _ := strconv.ParseFloat(str, 64)
	i := int(f * 1000)
	ss := i % 1000
	s := ((i - ss) / 1000) % 60
	m := (((i - ss) / 1000) - s) / 60 % 60
	h := (((i - ss) / 1000) - s - 60*m) / 60 / 60
	res := fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ss)
	return res
}

func makeXML(vi vInfo, chains []chain) []byte {
	res := `<?xml version="1.0" standalone="no"?>
<mlt LC_NUMERIC="C" version="7.8.0" title="Shotcut version 22.06.23" producer="main_bin">
  <profile description="automatic" width="` + vi.width + `" height="` + vi.height + `" progressive="1" sample_aspect_num="1" sample_aspect_den="1" display_aspect_num="16" display_aspect_den="9" frame_rate_num="30000" frame_rate_den="1001" colorspace="709"/>
  <playlist id="main_bin">
    <property name="xml_retain">1</property>
  </playlist>
  <producer id="black" in="00:00:00.000" out="` + vi.length + `">
    <property name="length">` + vi.length + `</property>
    <property name="eof">pause</property>
    <property name="resource">0</property>
    <property name="aspect_ratio">1</property>
    <property name="mlt_service">color</property>
    <property name="mlt_image_format">rgba</property>
    <property name="set.test_audio">0</property>
  </producer>
  <playlist id="background">
    <entry producer="black" in="00:00:00.000" out="` + vi.length + `"/>
  </playlist>
`
	for _, c := range chains {
		res += `  <chain id="chain` + fmt.Sprint(c.num) + `" out="` + vi.length + `">
    <property name="length">` + vi.length + `</property>
    <property name="eof">pause</property>
    <property name="resource">` + vi.fname + `</property>
    <property name="mlt_service">avformat-novalidate</property>
    <property name="seekable">1</property>
    <property name="audio_index">1</property>
    <property name="video_index">0</property>
    <property name="mute_on_pause">0</property>
    <property name="shotcut:hash"></property>
    <property name="shotcut:caption">` + vi.fname + `</property>
    <property name="xml">was here</property>
  </chain>
`
	}
	res += `	<playlist id="playlist0">
    <property name="shotcut:video">1</property>
    <property name="shotcut:name">V1</property>
`
	for _, c := range chains {
		res += `    <entry producer="chain` + fmt.Sprint(c.num) + `" in="` + c.in + `" out="` + c.out + `"/>
`
	}
	res += `	</playlist>
`
	res += `    <tractor id="tractor0" title="Shotcut version 22.06.23" in="00:00:00.000" out="` + vi.length + `">
    <property name="shotcut">1</property>
    <property name="shotcut:projectAudioChannels">2</property>
    <property name="shotcut:projectFolder">0</property>
    <track producer="background"/>
    <track producer="playlist0"/>
    <transition id="transition0">
      <property name="a_track">0</property>
      <property name="b_track">1</property>
      <property name="mlt_service">mix</property>
      <property name="always_active">1</property>
      <property name="sum">1</property>
    </transition>
    <transition id="transition1">
      <property name="a_track">0</property>
      <property name="b_track">1</property>
      <property name="version">0.1</property>
      <property name="mlt_service">frei0r.cairoblend</property>
      <property name="threads">0</property>
      <property name="disable">1</property>
    </transition>
  </tractor>
</mlt>
`
	return []byte(res)
}
