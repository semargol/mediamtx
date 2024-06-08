package conf

type RTPRConf struct {
	VideoPort  int
	AudioPort  int
	VideoCodec string
	AudioCodec string
	VideoPT    int
	AudioPT    int
	SPS        string
	VPS        string
	PPS        string
	Name       string
	RunOnReady string
}

type RTPSCLConf struct {
	StreamID string
	Codec    string
}

type PipeConfig struct {
	ID     int
	Name   string
	Source string
	Sink   []string
	State  string
	RTPR   RTPRConf
	RTPS   RTPSConf
	RTSPCL RTSPCLConf
	Type   string
}

type RTPSConf struct {
	VideoURL   string
	AudioURL   string
	Name       string
	VideoCodec string
}

type RTSPCLConf struct {
	Url string
}

type StrmConf struct {
	LogLavel int
	Pipes    map[int]PipeConfig
	RTSPSRV  ServerConfig
	WebRTC   ServerConfig
}

type ServerConfig struct {
	Address string
	State   string
}

func InitializeDefaultStrmConf() StrmConf {
	return StrmConf{
		LogLavel: 4,
		RTSPSRV: ServerConfig{
			Address: ":8554",
			State:   "stop",
		},
	}
}

var StrmGlobalConf *StrmConf

func (s *StrmConf) LookupRTPSbyURL(url string) *RTPSConf {
	var pc *RTPSConf = nil
	for _, spc := range s.Pipes {
		if spc.RTSPCL.Url == url {
			pc = &spc.RTPS
		}
	}
	return pc
}

func GetStrmConf() *StrmConf {
	return StrmGlobalConf
}

func LookupRTPSbyURL(url string) *RTPSConf {
	s := GetStrmConf()
	var pc *RTPSConf = nil
	for _, spc := range s.Pipes {
		if spc.RTSPCL.Url == url {
			pc = &spc.RTPS
		}
	}
	return pc
}
