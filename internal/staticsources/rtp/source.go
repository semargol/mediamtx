// Package rtp contains the RTP static source.
package rtp

import (
	"fmt"
	"net"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/multicast"
	"github.com/pion/rtp"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/defs"
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
	ReadTimeout        conf.StringDuration
	Parent             defs.StaticSourceParent
}

// Log implements logger.Writer.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[RTP source] "+format, args...)
}

// Run implements StaticSource.
func (s *Source) Run(params defs.StaticSourceRunParams) error {
	s.Log(logger.Debug, "connecting")

	hostPort := s.ResolvedSource[len("udp://"):]

	addrVideo, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return err
	}

	var pcVideo packetConn
	var pcAudio packetConn

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
		pcVideo = tmp.(*net.UDPConn)
	}

	defer pcVideo.Close()

	err = pcVideo.SetReadBuffer(udpKernelReadBufferSize)
	if err != nil {
		return err
	}

	videoMedi := &description.Media{
		Type: description.MediaTypeVideo,
		Formats: []format.Format{&format.H264{
			PayloadTyp:        96,
			PacketizationMode: 1,
		}},
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
				PayloadTyp: 97,
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
	bufVideo := make([]byte, 2048)
	readerErrVideo := make(chan error)
	go func() {
		readerErrVideo <- s.runReaderVideo(pcVideo, stream, medias, bufVideo)
	}()

	readerErrAudio := make(chan error)
	if s.ResolvedAudiSource != "" {
		bufAudio := make([]byte, 2048)
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
		pcVideo.Close()
		pcAudio.Close()
		<-readerErrVideo
		<-readerErrAudio
		return fmt.Errorf("terminated")
	}
}

func (s *Source) runReaderVideo(pc net.PacketConn,
	stream *stream.Stream,
	medias []*description.Media, buf []byte) error {

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
		stream.WriteRTPPacket(medias[0],
			medias[0].Formats[0],
			&pkt, time.Now(), time.Duration(0))
	}
}

func (s *Source) runReaderAudio(pc net.PacketConn,
	stream *stream.Stream,
	medias []*description.Media, buf []byte) error {

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
		stream.WriteRTPPacket(medias[1],
			medias[1].Formats[0],
			&pkt, time.Now(), time.Duration(0))
	}
}

// APISourceDescribe implements StaticSource.
func (*Source) APISourceDescribe() defs.APIPathSourceOrReader {
	return defs.APIPathSourceOrReader{
		Type: "rtpSource",
		ID:   "",
	}
}
