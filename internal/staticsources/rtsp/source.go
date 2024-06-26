// Package rtsp contains the RTSP static source.
package rtsp

import (
	"net"
	"strings"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/headers"
	"github.com/pion/rtp"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/defs"
	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/protocols/tls"
)

func createRangeHeader(cnf *conf.Path) (*headers.Range, error) {
	switch cnf.RTSPRangeType {
	case conf.RTSPRangeTypeClock:
		start, err := time.Parse("20060102T150405Z", cnf.RTSPRangeStart)
		if err != nil {
			return nil, err
		}

		return &headers.Range{
			Value: &headers.RangeUTC{
				Start: start,
			},
		}, nil

	case conf.RTSPRangeTypeNPT:
		start, err := time.ParseDuration(cnf.RTSPRangeStart)
		if err != nil {
			return nil, err
		}

		return &headers.Range{
			Value: &headers.RangeNPT{
				Start: start,
			},
		}, nil

	case conf.RTSPRangeTypeSMPTE:
		start, err := time.ParseDuration(cnf.RTSPRangeStart)
		if err != nil {
			return nil, err
		}

		return &headers.Range{
			Value: &headers.RangeSMPTE{
				Start: headers.RangeSMPTETime{
					Time: start,
				},
			},
		}, nil

	default:
		return nil, nil
	}
}

// Source is a RTSP static source.
type Source struct {
	ResolvedSource string
	ReadTimeout    conf.StringDuration
	WriteTimeout   conf.StringDuration
	WriteQueueSize int
	Parent         defs.StaticSourceParent

	RtpVideoURL string
	RtpAudioURL string

	rtpVideo *net.UDPConn
	rtpAudio *net.UDPConn

	rtpVideoAddr *net.UDPAddr
	rtpAudioAddr *net.UDPAddr
}

// Log implements logger.Writer.
func (s *Source) Log(level logger.Level, format string, args ...interface{}) {
	s.Parent.Log(level, "[RTSP source] "+format, args...)
}

func (c *Source) send(pkt *rtp.Packet) {
	//fmt.Println("send to ", addr, "msg", msg)
	data, _ := pkt.Marshal()
	if pkt.PayloadType == 96 {
		_, _ = c.rtpVideo.WriteToUDP(data, c.rtpVideoAddr)
	} else if pkt.PayloadType == 97 {
		_, _ = c.rtpAudio.WriteToUDP(data, c.rtpAudioAddr)
	}
}

// Run implements StaticSource.
func (s *Source) Run(params defs.StaticSourceRunParams) error {
	s.Log(logger.Debug, "connecting")

	decodeErrLogger := logger.NewLimitedLogger(s)

	var erv, era error
	var scheme, adress string
	var found bool

	if s.RtpVideoURL != "" {
		scheme, adress, found = strings.Cut(s.RtpVideoURL, "//") // rconf.VideoURL[len("udp://"):]
		if !found {
			adress = scheme
		}
		s.rtpVideoAddr, erv = net.ResolveUDPAddr("udp", adress)
		if erv == nil {
			s.rtpVideo, _ = net.ListenUDP("udp", nil)
		}
	}

	if s.RtpAudioURL != "" {
		scheme, adress, found = strings.Cut(s.RtpAudioURL, "//") // rconf.AudioURL[len("udp://"):]
		if !found {
			adress = scheme
		}
		s.rtpAudioAddr, era = net.ResolveUDPAddr("udp", adress)
		if era == nil {
			s.rtpAudio, _ = net.ListenUDP("udp", nil)
		}
	}

	c := &gortsplib.Client{
		Transport:      params.Conf.RTSPTransport.Transport,
		TLSConfig:      tls.ConfigForFingerprint(params.Conf.SourceFingerprint),
		ReadTimeout:    time.Duration(s.ReadTimeout),
		WriteTimeout:   time.Duration(s.WriteTimeout),
		WriteQueueSize: s.WriteQueueSize,
		AnyPortEnable:  params.Conf.RTSPAnyPort,
		OnRequest: func(req *base.Request) {
			s.Log(logger.Debug, "[c->s] %v", req)
		},
		OnResponse: func(res *base.Response) {
			s.Log(logger.Debug, "[s->c] %v", res)
		},
		OnTransportSwitch: func(err error) {
			s.Log(logger.Warn, err.Error())
		},
		OnPacketLost: func(err error) {
			decodeErrLogger.Log(logger.Warn, err.Error())
		},
		OnDecodeError: func(err error) {
			decodeErrLogger.Log(logger.Warn, err.Error())
		},
	}

	u, err := base.ParseURL(s.ResolvedSource)
	if err != nil {
		return err
	}

	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	defer c.Close()

	readErr := make(chan error)

	go func() {
		readErr <- func() error {
			desc, _, err := c.Describe(u)
			if err != nil {
				return err
			}

			err = c.SetupAll(desc.BaseURL, desc.Medias)
			if err != nil {
				return err
			}

			res := s.Parent.SetReady(defs.PathSourceStaticSetReadyReq{
				Desc:               desc,
				GenerateRTPPackets: false,
			})
			if res.Err != nil {
				return res.Err
			}

			defer s.Parent.SetNotReady(defs.PathSourceStaticSetNotReadyReq{})

			for _, medi := range desc.Medias {
				for _, forma := range medi.Formats {
					cmedi := medi
					cforma := forma

					c.OnPacketRTP(cmedi, cforma, func(pkt *rtp.Packet) {
						pts, ok := c.PacketPTS(cmedi, pkt)
						if !ok {
							return
						}

						if s.rtpVideo == nil && s.rtpAudio == nil {
							res.Stream.WriteRTPPacket(cmedi, cforma, pkt, time.Now(), pts)
						} else {
							s.send(pkt)
						}

					})
				}
			}

			rangeHeader, err := createRangeHeader(params.Conf)
			if err != nil {
				return err
			}

			_, err = c.Play(rangeHeader)
			if err != nil {
				return err
			}

			return c.Wait()
		}()
	}()

	for {
		select {
		case err := <-readErr:
			return err

		case <-params.ReloadConf:

		case <-params.Context.Done():
			c.Close()
			if s.rtpVideo != nil {
				_ = s.rtpVideo.Close()
			}
			if s.rtpAudio != nil {
				_ = s.rtpAudio.Close()
			}
			<-readErr
			return nil
		}
	}
}

// APISourceDescribe implements StaticSource.
func (*Source) APISourceDescribe() defs.APIPathSourceOrReader {
	return defs.APIPathSourceOrReader{
		Type: "rtspSource",
		ID:   "",
	}
}
