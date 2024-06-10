package rtp

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/multicast"
	"github.com/bluenviron/gortsplib/v4/pkg/rtptime"
	"github.com/pion/rtp"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/defs"
	"github.com/bluenviron/mediamtx/internal/formatprocessor"
	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/restrictnetwork"
	"github.com/bluenviron/mediamtx/internal/stream"
)

const (
	// same size as GStreamer's rtspsrc
	udpKernelReadBufferSize = 0x80000
)

type packetConn interface {
	net.PacketConn
	SetReadBuffer(int) error
}

// Source is a RTP static source.
type Source struct {
	ResolvedSource     string
	ResolvedAudiSource string
	VideoCodec         string
	VideoPT            int
	AudioPT            int
	SPS                string
	VPS                string
	PPS                string
	ReadTimeout        conf.StringDuration
	Parent             defs.StaticSourceParent
}

// Log implements logger.Writer.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[RTP source] "+format, args...)
}

func (s *Source) Run(params defs.StaticSourceRunParams) error {
	s.Log(logger.Debug, "connecting")

	hostPort := s.ResolvedSource[len("udp://"):]

	addrVideo, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return err
	}

	var pcVideo packetConn
	var pcAudio packetConn
	var pcRTCPVideo packetConn
	var pcRTCPAudio packetConn

	if ip4 := addrVideo.IP.To4(); ip4 != nil && addrVideo.IP.IsMulticast() {
		pcVideo, err = multicast.NewMultiConn(hostPort, true, net.ListenPacket)
		if err != nil {
			return err
		}
	} else {
		tmp, err := net.ListenPacket(restrictnetwork.Restrict("udp", addrVideo.String()))
		if err != nil {
			return err
		}
		clearUDPQueue(tmp)
		pcVideo = tmp.(*net.UDPConn)
	}

	defer pcVideo.Close()

	err = pcVideo.SetReadBuffer(udpKernelReadBufferSize)
	if err != nil {
		return err
	}

	sps, err := base64.StdEncoding.DecodeString(s.SPS)
	if err != nil {
		fmt.Println("Invalid SPS base64:", err)
	}

	pps, err := base64.StdEncoding.DecodeString(s.PPS)
	if err != nil {
		fmt.Println("Invalid PPS base64:", err)
	}

	s.Log(logger.Debug, "SPS: %s", s.SPS)
	s.Log(logger.Debug, "PPS: %s", s.PPS)

	videoMedi := &description.Media{
		Type: description.MediaTypeVideo,
		Formats: []format.Format{&format.H264{
			PayloadTyp:        uint8(s.VideoPT),
			PacketizationMode: 1,
			SPS:               sps,
			PPS:               pps,
		}},
	}

	if strings.EqualFold(s.VideoCodec, "h265") {
		vps, err := base64.StdEncoding.DecodeString(s.VPS)
		s.Log(logger.Debug, "VPS: %s", s.VPS)
		if err != nil {
			fmt.Println("Invalid PPS base64:", err)
		}

		videoMedi = &description.Media{
			Type: description.MediaTypeVideo,
			Formats: []format.Format{&format.H265{
				PayloadTyp: uint8(s.VideoPT),
				SPS:        sps,
				VPS:        vps,
				PPS:        pps,
			}},
		}
	}
	medias := []*description.Media{videoMedi}

	if s.ResolvedAudiSource != "" {
		hostPort = s.ResolvedAudiSource[len("udp://"):]

		addrAudio, err := net.ResolveUDPAddr("udp", hostPort)
		if err != nil {
			return err
		}

		if ip4 := addrAudio.IP.To4(); ip4 != nil && addrAudio.IP.IsMulticast() {
			pcAudio, err = multicast.NewMultiConn(hostPort, true, net.ListenPacket)
			if err != nil {
				return err
			}
		} else {
			tmp, err := net.ListenPacket(restrictnetwork.Restrict("udp", addrAudio.String()))
			if err != nil {
				return err
			}
			clearUDPQueue(tmp)
			pcAudio = tmp.(*net.UDPConn)
		}

		defer pcAudio.Close()

		err = pcAudio.SetReadBuffer(udpKernelReadBufferSize)
		if err != nil {
			return err
		}

		audioMedi := &description.Media{
			Type: description.MediaTypeAudio,
			Formats: []format.Format{&format.Opus{
				PayloadTyp: uint8(s.AudioPT),
				IsStereo:   true,
			}},
		}
		medias = []*description.Media{videoMedi, audioMedi}
	}

	var stream *stream.Stream

	if stream == nil {
		res := s.Parent.SetReady(defs.PathSourceStaticSetReadyReq{
			Desc:               &description.Session{Medias: medias},
			GenerateRTPPackets: false,
		})
		if res.Err != nil {
			return res.Err
		}

		stream = res.Stream
	}

	defer s.Parent.SetNotReady(defs.PathSourceStaticSetNotReadyReq{})
	bufVideo := make([]byte, udpKernelReadBufferSize)
	readerErrVideo := make(chan error)
	go func() {
		readerErrVideo <- s.runReaderVideo(pcVideo, stream, medias, bufVideo)
	}()

	readerErrAudio := make(chan error)
	if s.ResolvedAudiSource != "" {
		bufAudio := make([]byte, udpKernelReadBufferSize)
		go func() {
			readerErrAudio <- s.runReaderAudio(pcAudio, stream, medias, bufAudio)
		}()
	}
	select {
	case err := <-readerErrVideo:
		return err
	case err := <-readerErrAudio:
		return err
	case <-params.Context.Done():
		if pcVideo != nil {
			pcVideo.Close()
		}
		if pcAudio != nil {
			pcAudio.Close()
		}
		if pcRTCPVideo != nil {
			pcRTCPVideo.Close()
		}
		if pcRTCPAudio != nil {
			pcRTCPAudio.Close()
		}

		<-readerErrVideo
		// <-readerErrAudio
		return fmt.Errorf("terminated")
	}
}

func (s *Source) runReaderVideo(pc net.PacketConn,
	stream *stream.Stream,
	medias []*description.Media, buf []byte) error {
	// trackWrapper := &webrtc.TrackWrapper{ClockRat: medias[0].Formats[0].ClockRate()}
	timeDecoder := rtptime.NewGlobalDecoder()
	p, _ := formatprocessor.New(udpKernelReadBufferSize, medias[0].Formats[0], false)
	for {
		n, _, err := pc.ReadFrom(buf)

		if err != nil {
			return err
		}
		var pkt rtp.Packet
		if err := pkt.Unmarshal(buf[:n]); err != nil {
			fmt.Println("Failed to unmarshal RTP packet:", err)
			continue
		}

		pts, ok := timeDecoder.Decode(medias[0].Formats[0], &pkt)
		if !ok {
			continue
		}
		// fmt.Println("pts video: ", pts)

		un, err := p.ProcessRTPPacket(&pkt, time.Now(), pts, false)
		if err != nil {
			fmt.Println("err: ", err)
		}

		stream.WriteUnit(medias[0],
			medias[0].Formats[0],
			un)

		// stream.WriteRTPPacket(medias[0],
		// 	medias[0].Formats[0],
		// 	&pkt, time.Now(), time.Duration(0))
	}
}

func (s *Source) runReaderAudio(pc net.PacketConn,
	stream *stream.Stream,
	medias []*description.Media, buf []byte) error {
	timeDecoder := rtptime.NewGlobalDecoder()
	// p, _ := formatprocessor.New(udpKernelReadBufferSize, medias[1].Formats[0], true)
	for {
		n, _, err := pc.ReadFrom(buf)

		if err != nil {
			return err
		}
		var pkt rtp.Packet
		if err := pkt.Unmarshal(buf[:n]); err != nil {
			fmt.Println("Failed to unmarshal RTP packet:", err)
			continue
		}
		pts, ok := timeDecoder.Decode(medias[1].Formats[0], &pkt)
		if !ok {
			continue
		}
		// fmt.Println("pts audio: ", pts)

		stream.WriteRTPPacket(medias[1],
			medias[1].Formats[0],
			&pkt, time.Now(), pts)
		// stream.WriteRTPPacket(medias[1],
		// 	medias[1].Formats[0],
		// 	&pkt, time.Now(), time.Duration(0))
	}
}

func (s *Source) APIConfig() interface{} {
	return &struct {
		Source               string
		AudioSource          string
		VideoPayloadType     int
		AudioPayloadType     int
		ReadTimeout          conf.StringDuration
		AdditionalProtocols  []string
		AdditionalSources    map[string]string
		DisableSRTCPReceiver bool
	}{}
}

// APISourceDescribe implements StaticSource.
func (*Source) APISourceDescribe() defs.APIPathSourceOrReader {
	return defs.APIPathSourceOrReader{
		Type: "rtpSource",
		ID:   "",
	}
}
func clearUDPQueue(pc net.PacketConn) {
	buf := make([]byte, 2048)
	deadline := time.Now().Add(50 * time.Millisecond)
	pc.SetReadDeadline(deadline)

	for {
		_, _, err := pc.ReadFrom(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				break
			}
			fmt.Println("Error reading from UDP:", err)
			return
		}
	}
	// Reset the deadline to no deadline
	pc.SetReadDeadline(time.Time{})
}
