package rtp

import (
	"encoding/base64"
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

	videoNTPTime uint64
	videoOffset  int64
	audioNTPTime uint64
	audioOffset  int64
	videoSSRC    uint32
	audioSSRC    uint32
	td           float64
	mu           sync.Mutex
}

func ConvertNTPTime(NTPTime uint64) time.Time {
	const ntpUnixEpochOffset = 2208988800
	secs := int64(NTPTime >> 32)
	frac := uint64((NTPTime & 0xFFFFFFFF) * 1e9 >> 32)
	ntpTime := time.Unix(secs-ntpUnixEpochOffset, int64(frac))
	return ntpTime
}

func CalculateOffset(NTPTime uint64, RTPTime uint32, RTPClockRate uint32) int64 {
	ntpTime := ConvertNTPTime(NTPTime)

	// Convert RTPTime to NTP format (seconds and fraction of a second)
	RTPNTP := (uint64(RTPTime)) / uint64(RTPClockRate)

	// Calculate the difference between NTPTime and RTPNTP in seconds
	offset := (int64(ntpTime.UnixMilli()) - int64(RTPNTP*1000))
	return offset
}

// func CalculateNTPTime(RTPTime uint32, RTPClockRate uint32, offset int64) time.Time {
// 	ntpSeconds := int64(offset) + int64(RTPTime)/int64(RTPClockRate)
// 	// ntpTime := time.Unix(ntpSeconds, 0)
// 	ntpTime := time.Unix(ntpSeconds, 0)
// 	fmt.Println("RTPClockRate: ", RTPClockRate)
// 	fmt.Println(" NTP newTime:", ntpTime)
// 	return ntpTime
// }

func CalculateNTPTime(RTPTime uint32, RTPClockRate uint32, offset int64) time.Time {
	ntpMilSeconds := offset + 1000*int64(RTPTime)/int64(RTPClockRate)
	ntpTime := time.UnixMilli(ntpMilSeconds)
	return ntpTime
}

// Log implements logger.Writer.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[RTP source] "+format, args...)
}

func (s *Source) Run(params defs.StaticSourceRunParams) error {
	s.Log(logger.Debug, "connecting")

	var pcVideo packetConn
	var pcAudio packetConn
	var pcRTCPVideo packetConn
	var pcRTCPAudio packetConn

	hostPort := s.ResolvedSource[len("udp://"):]

	addrVideo, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return err
	}

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
	rtcpAddrVideo := fmt.Sprintf(":%d", addrVideo.Port+1)
	fmt.Println(rtcpAddrVideo)
	tmp, err := net.ListenPacket(restrictnetwork.Restrict("udp", rtcpAddrVideo))
	if err != nil {
		return err
	}
	pcRTCPVideo = tmp.(*net.UDPConn)
	defer pcRTCPVideo.Close()

	go s.runRTCPReader(pcRTCPVideo, s.videoSSRC)
	sprop := "Z2QAH6zZQFAFuwFqAgICgAAAAwCAAAAZB4wYyw==,aOvjyyLA"

	// Разделяем строку на две части
	parts := strings.Split(sprop, ",")
	if len(parts) != 2 {
		fmt.Println("Invalid sprop-parameter-sets")

	}

	// Декодируем SPS
	sps, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		fmt.Println("Invalid SPS base64:", err)

	}

	// Декодируем PPS
	pps, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Println("Invalid PPS base64:", err)

	}

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
		rtcpAddrAudio := fmt.Sprintf(":%d", addrAudio.Port+1)
		tmp, err = net.ListenPacket(restrictnetwork.Restrict("udp", rtcpAddrAudio))
		if err != nil {
			return err
		}
		pcRTCPAudio = tmp.(*net.UDPConn)
		defer pcRTCPAudio.Close()

		go s.runRTCPReader(pcRTCPAudio, s.audioSSRC)

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
			GenerateRTPPackets: true,
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

		s.videoSSRC = pkt.SSRC

		pts, ok := timeDecoder.Decode(medias[0].Formats[0], &pkt)
		if !ok {
			continue
		}
		// fmt.Println("pts video: ", pts)
		newNTPTime := CalculateNTPTime(pkt.Timestamp, 90000, s.videoOffset)
		fmt.Println("New v NTPTime:", newNTPTime)
		un, err := p.ProcessRTPPacket(&pkt, time.Now(), pts, false)
		if err != nil {
			fmt.Println("err: ", err)
		}

		stream.WriteUnit(medias[0],
			medias[0].Formats[0],
			un)

		// stream.WriteRTPPacket(medias[0],
		// 	medias[0].Formats[0],
		// 	&pkt, newNTPTime, pts)
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

		s.audioSSRC = pkt.SSRC

		pts, ok := timeDecoder.Decode(medias[1].Formats[0], &pkt)
		if !ok {
			continue
		}

		newNTPTime := CalculateNTPTime(pkt.Timestamp, 48000, s.audioOffset)
		fmt.Println("New a NTPTime:", newNTPTime)
		fmt.Println("Time.Now:", time.Now())
		stream.WriteRTPPacket(medias[1],
			medias[1].Formats[0],
			&pkt, time.Now(), pts)
		// mu.Unlock()
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

func (s *Source) handleRTCPPacket(packet rtcp.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// fmt.Printf("Received RTCP packet of type: %T\n", packet)

	switch pkt := packet.(type) {
	case *rtcp.SenderReport:
		ssrc := pkt.SSRC
		if ssrc == s.videoSSRC {
			s.videoNTPTime = pkt.NTPTime
			s.videoOffset = CalculateOffset(pkt.NTPTime, pkt.RTPTime, 90000)
		} else if ssrc == s.audioSSRC {
			s.audioNTPTime = pkt.NTPTime
			s.audioOffset = CalculateOffset(pkt.NTPTime, pkt.RTPTime, 48000)
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

func (s *Source) runRTCPReader(pc net.PacketConn, ssrc uint32) {
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
			// fmt.Printf("Processing RTCP packet of type: %T\n", pkt)
			s.handleRTCPPacket(pkt)
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
