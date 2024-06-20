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
	"github.com/pion/rtcp"
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
	VideoCodec         string
	VideoPT            int
	AudioPT            int
	SPS                string
	VPS                string
	PPS                string
	Jitter             int
	ReadTimeout        conf.StringDuration
	Parent             defs.StaticSourceParent

	videoNtpOffset    float64 // ntp - rtp
	audioNtpOffset    float64
	videoUtcOffset    float64 // utc - rtp
	audioUtcOffset    float64
	videoUtcOffsetOut float64 // utc - rtp
	audioUtcOffsetOut float64
	//videoShift     float64
	//audioShift     float64
	videoJitter float64
	audioJitter float64

	//videoShiftFiltered float64
	//audioShiftFiltered float64

	//videoShiftOut  float64
	//audioShiftOut  float64
	videoJitterOut float64
	audioJitterOut float64

	videoJitterDelay float64
	audioJitterDelay float64

	videoSSRC uint32
	audioSSRC uint32

	audioQueue SortedQueue
	videoQueue SortedQueue
}

var unixTimeShift float64

const DA bool = false
const DV bool = false

func (s *Source) init() {
	s.videoJitterDelay = float64(s.Jitter) / 1000.0
	s.audioJitterDelay = float64(s.Jitter) / 1000.0
	// fmt.Println("jitter: ", s.Jitter)
}

func init() {
	t1970 := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	t1900 := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	unixTimeShift = t1970.Sub(t1900).Seconds()
}

func UtcTime() float64 {
	return float64(time.Now().UTC().UnixMilli())/1000.0 + unixTimeShift // 2208988800
}

func (s *Source) VideoNtpToRtp(ntp float64) float64 {
	return ntp - s.videoNtpOffset // * 90000.0 // sec -> rtp
}

func (s *Source) VideoRtpToNtp(rtp float64) float64 {
	return s.videoNtpOffset + rtp // / 90000.0 // rtp -> sec
}

func (s *Source) AudioNtpToRtp(ntp float64) float64 {
	return ntp - s.audioNtpOffset // * 48000.0 // sec -> rtp
}

func (s *Source) AudioRtpToNtp(rtp float64) float64 {
	return s.audioNtpOffset + rtp // / 48000.0 // rtp -> sec
}

func (s *Source) AudioControlReceived(rp *rtcp.SenderReport) {
	if rp.NTPTime != 0 {
		ntpTime := float64(rp.NTPTime) / 65536.0 / 65536.0
		rtpTime := float64(rp.RTPTime) / 48000.0
		s.audioNtpOffset = ntpTime - rtpTime
		if DA {
			fmt.Println()
			fmt.Printf("%10.3f AC               MC %3d RTP %14.3f NTP %14.3f TZ %14.3f DIF %14.3f\n", UtcTime(), rp.Header().Type, rtpTime, ntpTime, s.audioNtpOffset, s.AudioNtpToRtp(ntpTime)-rtpTime) //rp.rtcp_packet_type,3} RTP {rp.rtp,14:F3} NPT {rp.ntp,14:F3} TZ {audioNtpOffset,14:F3} {AudioNtpToRtp(rp.ntp) - rp.rtp}");
			fmt.Println()
		}
	}
}

func (s *Source) AudioDataReceived(rp *rtp.Packet) float64 {
	if s.audioJitterDelay == 0 {
		s.init()
	}
	utcTime := UtcTime()
	rtpTime := float64(rp.Timestamp) / 48000.0
	sendingTime := 0.0
	if s.audioUtcOffset == 0 {
		s.audioUtcOffset = utcTime - rtpTime
		//ntpTime := s.AudioRtpToNtp(rtpTime)
		//s.audioShift = utcTime - ntpTime
	}
	if s.audioNtpOffset == 0 {
		s.audioNtpOffset = utcTime - rtpTime
	}
	if s.audioNtpOffset != 0 {
		ntpTime := s.AudioRtpToNtp(rtpTime)
		s.audioJitter = utcTime - rtpTime - s.audioUtcOffset
		s.audioUtcOffset += s.audioJitter / 4
		s.audioUtcOffsetOut = s.audioUtcOffset + +s.audioJitterDelay
		//s.audioShift = utcTime - ntpTime
		//s.audioShiftFiltered = s.audioShift*0.125 + s.audioShiftFiltered*0.875
		sendingTime = rtpTime + s.audioUtcOffsetOut
		s.audioQueue.Put(rp, sendingTime)
		if DA {
			fmt.Printf("%14.3f AD SEQ %9d PT %3d RTP %14.3f NTP %14.3f JTR %14.3f DLY %14.3f\n", utcTime, rp.SequenceNumber, rp.PayloadType, rtpTime, ntpTime, s.audioJitter, sendingTime-utcTime)
		}
	}
	return sendingTime
}

func (s *Source) VideoControlReceived(rp *rtcp.SenderReport) {
	if rp.NTPTime != 0 {
		utcTime := UtcTime()
		rtpTime := float64(rp.RTPTime) / 48000.0
		ntpTime := float64(rp.NTPTime) / 65536.0 / 65536.0
		s.videoNtpOffset = ntpTime - rtpTime
		if DV {
			fmt.Println()
			fmt.Printf("%10.3f VC               MC %3d RTP %14.3f NTP %14.3f TZ %14.3f DIF %14.3f\n", utcTime, rp.Header().Type, rtpTime, ntpTime, s.videoNtpOffset, s.VideoNtpToRtp(ntpTime)-rtpTime) //rp.rtcp_packet_type,3} RTP {rp.rtp,14:F3} NPT {rp.ntp,14:F3} TZ {audioNtpOffset,14:F3} {AudioNtpToRtp(rp.ntp) - rp.rtp}");
			fmt.Println()
		}
	}
}

func (s *Source) VideoDataReceived(rp *rtp.Packet) float64 {
	if s.videoJitterDelay == 0 {
		s.init()
	}
	utcTime := UtcTime()
	rtpTime := float64(rp.Timestamp) / 90000.0
	sendingTime := 0.0
	if s.videoUtcOffset == 0 {
		s.videoUtcOffset = utcTime - rtpTime
		//ntpTime := s.AudioRtpToNtp(rtpTime)
		//s.audioShift = utcTime - ntpTime
	}
	if s.videoNtpOffset == 0 {
		s.videoNtpOffset = utcTime - rtpTime
	}
	if s.videoNtpOffset != 0 {
		ntpTime := s.VideoRtpToNtp(rtpTime)
		s.videoJitter = utcTime - rtpTime - s.videoUtcOffset
		s.videoUtcOffset += s.videoJitter / 4
		s.videoUtcOffsetOut = s.videoUtcOffset + s.videoJitterDelay
		//s.videoShift = utcTime - ntpTime
		//s.videoShiftFiltered = s.videoShift*0.125 + s.videoShiftFiltered*0.875
		sendingTime = rtpTime + s.videoUtcOffsetOut
		s.videoQueue.Put(rp, sendingTime)
		if DV {
			fmt.Printf("%14.3f VD SEQ %9d PT %3d RTP %14.3f NTP %14.3f JTR %14.3f DLY %14.3f\n", utcTime, rp.SequenceNumber, rp.PayloadType, rtpTime, ntpTime, s.videoJitter, sendingTime-utcTime)
		}
	}
	return sendingTime
}

// Log implements logger.Writer.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[RTP source] "+format, args...)
}

/*
// Global variables for storing NTP timestamps and SSRCs
var (
	videoNTPTime uint64
	audioNTPTime uint64
	td           float64
	tv           uint64
	nv           uint64
	pv           uint64
	ta           uint64
	na           uint64
	pa           uint64
	vscale       int64
	voff         int64
	ascale       int64
	aoff         int64
	mu           sync.Mutex
)
*/

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

	go runRTCPReader(s, pcRTCPVideo, s.videoSSRC)

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

		// Create RTCP connection for audio
		// rtcpAddrAudio := fmt.Sprintf("%s:%d", addrAudio.IP.String(), addrAudio.Port+1)
		rtcpAddrAudio := fmt.Sprintf(":%d", addrAudio.Port+1)
		tmp, err = net.ListenPacket(restrictnetwork.Restrict("udp", rtcpAddrAudio))
		if err != nil {
			return err
		}
		pcRTCPAudio = tmp.(*net.UDPConn)
		defer pcRTCPAudio.Close()

		go runRTCPReader(s, pcRTCPAudio, s.audioSSRC)

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
		if s.videoSSRC == 0 {
			s.videoSSRC = pkt.SSRC
			fmt.Printf("Initialized videoSSRC to %d\n", s.videoSSRC)
		}
		pts, ok := timeDecoder.Decode(medias[0].Formats[0], &pkt)
		if !ok {
			continue
		}

		s.VideoDataReceived(&pkt)
		//stream.WriteRTPPacket(medias[0], medias[0].Formats[0], &pkt, time.Now(), pts) // time.Duration(0))
		for {
			rp, st := s.videoQueue.Get(UtcTime())
			if rp == nil {
				break
			}
			utcTime := UtcTime()
			rtpTime := float64(rp.Timestamp) / 90000.0
			ntpTime := s.VideoRtpToNtp(rtpTime)
			s.videoJitterOut = utcTime - rtpTime - s.videoUtcOffsetOut
			//s.videoShiftOut = utcTime - ntpTime
			if DV {
				fmt.Printf("               VD SEQ %9d PT %3d RTP %14.3f NTP %14.3f JTR %14.3f DLY %14.3f QUE %4d\n", rp.SequenceNumber, rp.PayloadType, rtpTime, ntpTime, s.videoJitterOut, st-utcTime, s.videoQueue.Count())
			}
			stream.WriteRTPPacket(medias[0], medias[0].Formats[0], rp, time.Now(), pts) // time.Duration(0))
		}

	}
}

func (s *Source) runReaderAudio(pc net.PacketConn,
	stream *stream.Stream,
	medias []*description.Media, buf []byte) error {
	timeDecoder := rtptime.NewGlobalDecoder()
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
		if s.audioSSRC == 0 {
			s.audioSSRC = pkt.SSRC
			fmt.Printf("Initialized audioSSRC to %d\n", s.audioSSRC)
		}

		pts, ok := timeDecoder.Decode(medias[1].Formats[0], &pkt)
		if !ok {
			continue
		}
		// fmt.Println("pts audio: ", pts)
		// mu.Lock()
		// defer mu.Unlock()

		s.AudioDataReceived(&pkt)
		//stream.WriteRTPPacket(medias[1], medias[1].Formats[0], &pkt, time.Now(), pts)
		for {
			rp, st := s.audioQueue.Get(UtcTime())
			if rp == nil {
				break
			}
			utcTime := UtcTime()
			rtpTime := float64(rp.Timestamp) / 48000.0
			ntpTime := s.AudioRtpToNtp(rtpTime)
			s.audioJitterOut = utcTime - rtpTime - s.audioUtcOffsetOut
			//s.audioShiftOut = utcTime - ntpTime
			if DA {
				fmt.Printf("               AD SEQ %9d PT %3d RTP %14.3f NTP %14.3f JTR %14.3f DLY %14.3f QUE %4d\n", rp.SequenceNumber, rp.PayloadType, rtpTime, ntpTime, s.audioJitterOut, st-utcTime, s.videoQueue.Count())
			}
			stream.WriteRTPPacket(medias[1], medias[1].Formats[0], rp, time.Now(), pts)
		}

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

func handleRTCPPacket(s *Source, packet rtcp.Packet) {
	//mu.Lock()
	//defer mu.Unlock()

	//fmt.Printf("Received RTCP packet of type: %T\n", packet)

	switch pkt := packet.(type) {
	case *rtcp.SenderReport:
		ssrc := pkt.SSRC
		//fmt.Println("Received SenderReport:", pkt)
		if ssrc == s.videoSSRC {
			s.VideoControlReceived(pkt)
		} else if ssrc == s.audioSSRC {
			s.AudioControlReceived(pkt)
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

func runRTCPReader(s *Source, pc net.PacketConn, ssrc uint32) {
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
			handleRTCPPacket(s, pkt)
		}
	}
}
