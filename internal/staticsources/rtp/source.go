package rtp

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/multicast"
	"github.com/bluenviron/gortsplib/v4/pkg/rtptime"
	"github.com/pion/rtcp"
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
	ReadTimeout        conf.StringDuration
	Parent             defs.StaticSourceParent
}

// Log implements logger.Writer.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[RTP source] "+format, args...)
}

// Global variables for storing NTP timestamps and SSRCs
var (
	videoNTPTime uint64
	audioNTPTime uint64
	videoSSRC    uint32
	audioSSRC    uint32
	td           float64
	mu           sync.Mutex
)

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

	// Create RTCP connection for video
	// rtcpAddrVideo := fmt.Sprintf("%s:%d", addrVideo.IP.String(), addrVideo.Port+1)
	rtcpAddrVideo := fmt.Sprintf(":%d", addrVideo.Port+1)
	fmt.Println(rtcpAddrVideo)
	tmp, err := net.ListenPacket(restrictnetwork.Restrict("udp", rtcpAddrVideo))
	if err != nil {
		return err
	}
	pcRTCPVideo = tmp.(*net.UDPConn)
	defer pcRTCPVideo.Close()

	go runRTCPReader(pcRTCPVideo, videoSSRC)

	videoMedi := &description.Media{
		Type: description.MediaTypeVideo,
		Formats: []format.Format{&format.H264{
			PayloadTyp:        uint8(s.VideoPT),
			PacketizationMode: 1,
		}},
	}

	if strings.EqualFold(s.VideoCodec, "h265") {
		videoMedi = &description.Media{
			Type: description.MediaTypeVideo,
			Formats: []format.Format{&format.H265{
				PayloadTyp: uint8(s.VideoPT),
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

		// Create RTCP connection for audio
		// rtcpAddrAudio := fmt.Sprintf("%s:%d", addrAudio.IP.String(), addrAudio.Port+1)
		rtcpAddrAudio := fmt.Sprintf(":%d", addrAudio.Port+1)
		tmp, err = net.ListenPacket(restrictnetwork.Restrict("udp", rtcpAddrAudio))
		if err != nil {
			return err
		}
		pcRTCPAudio = tmp.(*net.UDPConn)
		defer pcRTCPAudio.Close()

		go runRTCPReader(pcRTCPAudio, audioSSRC)

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
		// Initialize videoSSRC
		if videoSSRC == 0 {
			videoSSRC = pkt.SSRC
			fmt.Printf("Initialized videoSSRC to %d\n", videoSSRC)
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
		// Initialize audioSSRC
		if audioSSRC == 0 {
			audioSSRC = pkt.SSRC
			fmt.Printf("Initialized audioSSRC to %d\n", audioSSRC)
		}

		pts, ok := timeDecoder.Decode(medias[1].Formats[0], &pkt)
		if !ok {
			continue
		}
		// fmt.Println("pts audio: ", pts)
		mu.Lock()
		defer mu.Unlock()
		stream.WriteRTPPacket(medias[1],
			medias[1].Formats[0],
			&pkt, time.Now().Add(time.Duration(td)*time.Millisecond), pts)
		mu.Unlock()
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

func handleRTCPPacket(packet rtcp.Packet) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Received RTCP packet of type: %T\n", packet)

	switch pkt := packet.(type) {
	case *rtcp.SenderReport:
		ssrc := pkt.SSRC
		fmt.Println("Received SenderReport:", pkt)
		if ssrc == videoSSRC {
			videoNTPTime = pkt.NTPTime
			fmt.Printf("Updated videoNTPTime to %d\n", videoNTPTime)
			// Calculate time difference
			if videoNTPTime != 0 && audioNTPTime != 0 {
				td = calculateNTPTimestampDifference(audioNTPTime, videoNTPTime)
				fmt.Println("NTP time difference between video and audio:", td)
			}
		} else if ssrc == audioSSRC {
			audioNTPTime = pkt.NTPTime
			fmt.Printf("Updated audioNTPTime to %d\n", audioNTPTime)
		} else {
			fmt.Printf("Unknown SSRC: %d\n", pkt.SSRC)
		}
	case *rtcp.ReceiverReport:
		// Ignore ReceiverReports as we don't need to process them here
		fmt.Println("Ignoring ReceiverReport")
	default:
		// Ignore all other unexpected types
		fmt.Printf("Ignoring RTCP packet of unexpected type: %T\n", pkt)
	}
}

func runRTCPReader(pc net.PacketConn, ssrc uint32) {
	buf := make([]byte, 1500)
	for {
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			}
			fmt.Println("Error reading from RTCP:", err)
			return
		}

		packets, err := rtcp.Unmarshal(buf[:n])
		if err != nil {
			fmt.Println("Failed to unmarshal RTCP packet:", err)
			continue
		}

		for _, pkt := range packets {
			fmt.Printf("Processing RTCP packet of type: %T\n", pkt)
			handleRTCPPacket(pkt)
		}
	}
}

// Goroutine to send RTCP Receiver Reports
func runRTCPSender(pc net.PacketConn) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()

		rr := &rtcp.ReceiverReport{
			SSRC: videoSSRC,
			Reports: []rtcp.ReceptionReport{
				{
					SSRC:               videoSSRC,
					FractionLost:       0,
					TotalLost:          0,
					LastSequenceNumber: 0,
					Jitter:             0,
					LastSenderReport:   uint32(videoNTPTime >> 16),
					Delay:              0,
				},
			},
		}

		buf, err := rr.Marshal()
		mu.Unlock()
		if err != nil {
			fmt.Println("Failed to marshal RTCP Receiver Report:", err)
			continue
		}

		_, err = pc.WriteTo(buf, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5001}) // Update IP and port if needed
		if err != nil {
			fmt.Println("Failed to send RTCP Receiver Report:", err)
		}
	}
}

// Calculate the difference between two NTP timestamps in seconds
func calculateNTPTimestampDifference(ntp1, ntp2 uint64) float64 {
	// NTP epoch offset from Unix epoch in seconds
	const ntpEpochOffset = 2208988800

	// Extract the integer and fractional parts of the NTP timestamps
	secs1 := float64(ntp1>>32) - ntpEpochOffset
	fracs1 := float64(ntp1&0xFFFFFFFF) / 4294967296.0

	secs2 := float64(ntp2>>32) - ntpEpochOffset
	fracs2 := float64(ntp2&0xFFFFFFFF) / 4294967296.0

	// Combine integer and fractional parts and calculate the difference in seconds
	time1 := secs1 + fracs1
	time2 := secs2 + fracs2
	diff := time2 - time1

	// Convert the difference to milliseconds
	return diff * 1000
}
