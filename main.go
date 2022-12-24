package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	flag.Parse()
	err := run(flag.Arg(0), flag.Arg(1))
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func run(fname string, mode string) error {
	if mode == "del" {
		_, err := os.Stdout.Write(makeMLT2(fname))
		if err != nil {
			return err
		}
		return nil
	}
	if mode == "mark" {
		_, err := os.Stdout.Write(makeMLT3(fname))
		if err != nil {
			return err
		}
		return nil
	}
	if mode == "vtt2mlt" {
		_, err := os.Stdout.Write(vtt2mlt(fname))
		if err != nil {
			return err
		}
		return nil
	}
	if mode == "mlt2vtt" {
		_, err := os.Stdout.Write(mlt2vtt(fname))
		if err != nil {
			return err
		}
		return nil
	}
	_, err := os.Stdout.Write(makeMLT(fname))
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
		"ffprobe", "-hide_banner",
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
	num  int
	in   string
	out  string
	text string
}

func makeMLT(fname string) []byte {
	buf, _ := os.ReadFile(fname + ".txt")
	lines := strings.Split(string(buf), "\n")
	chains := []chain{}
	i := 0
	start := "00:00:00.000"
	for _, line := range lines {
		if line == "" {
			continue
		}
		t := toTime(line)
		if t == "00:00:00.000" {
			continue
		}
		chains = append(chains, chain{num: i, in: start, out: t})
		start = t
		i = i + 1
	}
	vi := newVInfo(fname)
	if start != vi.length {
		chains = append(chains, chain{num: i, in: start, out: vi.length})
	}
	res := makeXML(vi, chains)
	return res
}

func makeMLT2(fname string) []byte {
	buf, _ := os.ReadFile(fname + ".txt")
	lines := strings.Split(string(buf), "\n")
	chains := []chain{}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		f := toTime(line)
		i = i + 1
		t := toTime(lines[i])
		chains = append(chains, chain{num: i, in: f, out: t})
	}
	vi := newVInfo(fname)
	res := makeXML(vi, chains)
	return res
}

func makeMLT3(fname string) []byte {
	buf, _ := os.ReadFile(fname + ".txt")
	lines := strings.Split(string(buf), "\n")
	chains := []chain{}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		f := toTime(line)
		i = i + 1
		t := toTime(lines[i])
		chains = append(chains, chain{num: i, in: f, out: t, text: "Marker" + fmt.Sprint(i)})
	}
	vi := newVInfo(fname)
	res := makeXML(vi, chains)
	return res
}

func vtt2mlt(fname string) []byte {
	buf, _ := os.ReadFile(fname + ".vtt")
	lines := strings.Split(string(buf), "\n")
	chains := []chain{}
	n := 0
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, " --> ") {
			c := chain{num: n}
			items := strings.Split(line, " --> ")
			c.in = items[0]
			c.out = items[1]
			i++
			c.text = lines[i]
			chains = append(chains, c)
		}
	}
	vi := newVInfo(fname)
	res := makeXML(vi, chains)
	return res
}

type Mlt struct {
	XMLName      xml.Name `xml:"mlt"`
	Text         string   `xml:",chardata"`
	LCNUMERIC    string   `xml:"LC_NUMERIC,attr"`
	Version      string   `xml:"version,attr"`
	Title        string   `xml:"title,attr"`
	AttrProducer string   `xml:"producer,attr"`
	Profile      struct {
		Text             string `xml:",chardata"`
		Description      string `xml:"description,attr"`
		Width            string `xml:"width,attr"`
		Height           string `xml:"height,attr"`
		Progressive      string `xml:"progressive,attr"`
		SampleAspectNum  string `xml:"sample_aspect_num,attr"`
		SampleAspectDen  string `xml:"sample_aspect_den,attr"`
		DisplayAspectNum string `xml:"display_aspect_num,attr"`
		DisplayAspectDen string `xml:"display_aspect_den,attr"`
		FrameRateNum     string `xml:"frame_rate_num,attr"`
		FrameRateDen     string `xml:"frame_rate_den,attr"`
		Colorspace       string `xml:"colorspace,attr"`
	} `xml:"profile"`
	Chain []struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"id,attr"`
		Out      string `xml:"out,attr"`
		Property []struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"property"`
		Filter struct {
			Text     string `xml:",chardata"`
			ID       string `xml:"id,attr"`
			Out      string `xml:"out,attr"`
			Property []struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"property"`
		} `xml:"filter"`
	} `xml:"chain"`
	Producer []struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"id,attr"`
		In       string `xml:"in,attr"`
		Out      string `xml:"out,attr"`
		Property []struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"property"`
		Filter struct {
			Text     string `xml:",chardata"`
			ID       string `xml:"id,attr"`
			Out      string `xml:"out,attr"`
			Property []struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"property"`
		} `xml:"filter"`
	} `xml:"producer"`
	Playlist []struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"id,attr"`
		Title    string `xml:"title,attr"`
		Property []struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"property"`
		Entry []struct {
			Text     string `xml:",chardata"`
			Producer string `xml:"producer,attr"`
			In       string `xml:"in,attr"`
			Out      string `xml:"out,attr"`
		} `xml:"entry"`
		Blank struct {
			Text   string `xml:",chardata"`
			Length string `xml:"length,attr"`
		} `xml:"blank"`
	} `xml:"playlist"`
	Tractor struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"id,attr"`
		Title    string `xml:"title,attr"`
		In       string `xml:"in,attr"`
		Out      string `xml:"out,attr"`
		Property []struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"property"`
		Properties struct {
			Text       string `xml:",chardata"`
			Name       string `xml:"name,attr"`
			Properties []struct {
				Text     string `xml:",chardata"`
				Name     string `xml:"name,attr"`
				Property []struct {
					Text string `xml:",chardata"`
					Name string `xml:"name,attr"`
				} `xml:"property"`
			} `xml:"properties"`
		} `xml:"properties"`
		Track []struct {
			Text     string `xml:",chardata"`
			Producer string `xml:"producer,attr"`
		} `xml:"track"`
		Transition []struct {
			Text     string `xml:",chardata"`
			ID       string `xml:"id,attr"`
			Property []struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"property"`
		} `xml:"transition"`
	} `xml:"tractor"`
}

type marker struct {
	text  string
	start string
	end   string
	color string
}

func mlt2vtt(fname string) []byte {
	buf, _ := os.ReadFile(fname + ".mlt")
	mlt := Mlt{}
	err := xml.Unmarshal(buf, &mlt)
	if err != nil {
		fmt.Printf("%v", err)
		return nil
	}
	fmt.Printf("WebVTT\n\n")
	if mlt.Tractor.Properties.Name == "shotcut:markers" {
		for i, ps := range mlt.Tractor.Properties.Properties {
			fmt.Println(i)
			m := marker{}
			for _, p := range ps.Property {
				switch p.Name {
				case "text":
					m.text = p.Text
				case "start":
					m.start = p.Text
				case "end":
					m.end = p.Text
				case "color":
					m.color = p.Text
				}
			}
			fmt.Printf("%s --> %s\n", m.start, m.end)
			fmt.Printf("%s\n\n", m.text)
		}
	}
	return []byte{}
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
		if c.text != "" {
			break
		}
	}
	res += `	<playlist id="playlist0">
    <property name="shotcut:video">1</property>
    <property name="shotcut:name">V1</property>
`
	for _, c := range chains {
		if c.text != "" {
			res += `    <entry producer="chain` + fmt.Sprint(c.num) + `" in="00:00:00.000" out="` + vi.length + `"/>
			`
			break
		}
		res += `    <entry producer="chain` + fmt.Sprint(c.num) + `" in="` + c.in + `" out="` + c.out + `"/>
`
	}
	res += `	</playlist>
`
	res += `    <tractor id="tractor0" title="Shotcut version 22.06.23" in="00:00:00.000" out="` + vi.length + `">
    <property name="shotcut">1</property>
    <property name="shotcut:projectAudioChannels">2</property>
    <property name="shotcut:projectFolder">0</property>
`
	if chains[0].text != "" {
		res += `	<properties name="shotcut:markers">
`
		for i, c := range chains {
			res += `	<properties name="` + fmt.Sprint(i) + `">
		  <property name="text">` + c.text + `</property>
		  <property name="start">` + c.in + `</property>
		  <property name="end">` + c.out + `</property>
		  <property name="color">#008000</property>
		</properties>
`
		}
		res += `	  </properties>
  `
	}
	res += `	<track producer="background"/>
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
