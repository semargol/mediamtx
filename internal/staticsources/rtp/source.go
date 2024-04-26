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

type packetConnReader struct {
	net.PacketConn
}

func newPacketConnReader(pc net.PacketConn) *packetConnReader {
	return &packetConnReader{
		PacketConn: pc,
	}
}

func (r *packetConnReader) Read(p []byte) (int, error) {
	n, _, err := r.PacketConn.ReadFrom(p)
	return n, err
}

type packetConn interface {
	net.PacketConn
	SetReadBuffer(int) error
}

// Source is a RTP static source.
type Source struct {
	ResolvedSource string
	ReadTimeout    conf.StringDuration
	Parent         defs.StaticSourceParent
}

// Log implements logger.Writer.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[RTP source] "+format, args...)
}

// Run implements StaticSource.
func (s *Source) Run(params defs.StaticSourceRunParams) error {
	s.Log(logger.Debug, "connecting")

	hostPort := s.ResolvedSource[len("udp://"):]

	addr, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return err
	}

	var pc packetConn

	if ip4 := addr.IP.To4(); ip4 != nil && addr.IP.IsMulticast() {
		pc, err = multicast.NewMultiConn(hostPort, true, net.ListenPacket)
		if err != nil {
			return err
		}
	} else {
		tmp, err := net.ListenPacket(restrictnetwork.Restrict("udp", addr.String()))
		if err != nil {
			return err
		}
		pc = tmp.(*net.UDPConn)
	}

	defer pc.Close()

	err = pc.SetReadBuffer(udpKernelReadBufferSize)
	if err != nil {
		return err
	}

	readerErr := make(chan error)
	go func() {
		readerErr <- s.runReader(pc)
	}()

	select {
	case err := <-readerErr:
		return err

	case <-params.Context.Done():
		pc.Close()
		<-readerErr
		return fmt.Errorf("terminated")
	}
}

func (s *Source) runReader(pc net.PacketConn) error {
	pc.SetReadDeadline(time.Now().Add(time.Duration(s.ReadTimeout)))

	medi := &description.Media{
		Type: description.MediaTypeVideo,
		Formats: []format.Format{&format.H264{
			PayloadTyp:        96,
			PacketizationMode: 1,
		}},
	}
	medias := []*description.Media{medi}

	var stream *stream.Stream

	res := s.Parent.SetReady(defs.PathSourceStaticSetReadyReq{
		Desc:               &description.Session{Medias: medias},
		GenerateRTPPackets: false,
	})
	if res.Err != nil {
		return res.Err
	}

	defer s.Parent.SetNotReady(defs.PathSourceStaticSetNotReadyReq{})

	stream = res.Stream
	// Buffer to read incoming packets
	buf := make([]byte, 2048) // Adjust the size based on expected
	for {
		pc.SetReadDeadline(time.Now().Add(time.Duration(s.ReadTimeout)))
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			return err
		}
		var pkt rtp.Packet
		if err := pkt.Unmarshal(buf[:n]); err != nil {
			fmt.Println("Failed to unmarshal RTP packet:", err)
			continue
		}
		//fmt.Printf("Received %+v\n", pkt)
		//pts := time.Duration(pkt.Timestamp) * time.Millisecond
		pts := time.Duration(0)

		stream.WriteRTPPacket(medi,
			medi.Formats[0],
			&pkt, time.Now(), pts)
	}
}

// APISourceDescribe implements StaticSource.
func (*Source) APISourceDescribe() defs.APIPathSourceOrReader {
	return defs.APIPathSourceOrReader{
		Type: "rtpSource",
		ID:   "",
	}
}
