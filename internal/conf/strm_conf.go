package conf

import "strconv"

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

type BufConf struct {
	Size int
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
	BUF    BufConf
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

var StrmGlobalConf StrmConf

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
	return &StrmGlobalConf
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

var STRMGlobalConfiguration StrmGlobalConfiguration = StrmGlobalConfiguration{}

type StrmGlobalConfiguration struct {
	BUFF map[string]BufferGlobalConfiguration
	RTPR map[string]RtprGlobalConfiguration
	RTPS map[string]RtpsGlobalConfiguration
}

func init() {
	STRMGlobalConfiguration.init()
}

type BufferGlobalConfiguration struct {
	Conf map[string]string
}

type RtprGlobalConfiguration struct {
	Conf map[string]string
}

type RtpsGlobalConfiguration struct {
	Conf map[string]string
}

func (c *StrmGlobalConfiguration) init() {
	c.BUFF = make(map[string]BufferGlobalConfiguration)
	c.BUFF["1"] = BufferGlobalConfiguration{Conf: make(map[string]string)}
	c.BUFF["2"] = BufferGlobalConfiguration{Conf: make(map[string]string)}
	c.BUFF["3"] = BufferGlobalConfiguration{Conf: make(map[string]string)}
	c.BUFF["4"] = BufferGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPR = make(map[string]RtprGlobalConfiguration)
	c.RTPR["1"] = RtprGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPR["2"] = RtprGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPR["3"] = RtprGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPR["4"] = RtprGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPS = make(map[string]RtpsGlobalConfiguration)
	c.RTPS["1"] = RtpsGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPS["2"] = RtpsGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPS["3"] = RtpsGlobalConfiguration{Conf: make(map[string]string)}
	c.RTPS["4"] = RtpsGlobalConfiguration{Conf: make(map[string]string)}
}

func (c *StrmGlobalConfiguration) SetRtpr(data map[string]string) {
	if c.RTPR == nil {
		c.RTPR = make(map[string]RtprGlobalConfiguration)
	}
	for k, v := range data {
		c.RTPR[data["id"]].Conf[k] = v
	}
	c.BUFF[data["id"]].Conf["size"] = "0"
}

func (c *StrmGlobalConfiguration) SetRtps(data map[string]string) {
	if c.RTPS == nil {
		c.RTPS = make(map[string]RtpsGlobalConfiguration)
	}
	for k, v := range data {
		c.RTPS[data["id"]].Conf[k] = v
	}
}

func (c *StrmGlobalConfiguration) SetBuf(data map[string]string) {
	if c.BUFF == nil {
		c.BUFF = make(map[string]BufferGlobalConfiguration)
	}
	for k, v := range data {
		c.BUFF[data["id"]].Conf[k] = v
	}
}

func (c *StrmGlobalConfiguration) GetRtprId(url string) string {
	for k, v := range c.RTPR {
		for _, rv := range v.Conf {
			if url == rv {
				return k
			}
		}
	}
	return ""
}

func (c *StrmGlobalConfiguration) GetBufSize(id string) float64 {
	v, e := strconv.Atoi(c.BUFF[id].Conf["size"])
	if e != nil {
		return 0.0
	}
	return float64(v) * 0.001
}

func (c *StrmGlobalConfiguration) SetBufSize(id string, msec string) {
	c.BUFF[id].Conf["size"] = msec
}
