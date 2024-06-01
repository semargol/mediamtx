package conf

type RTPRConf struct {
	VideoURL   string
	AudioURL   string
	VideoCodec string
	AudioCodec string
	VideoPT    int
	AudioPT    int
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
	VideoPort  int
	AudioPort  int
	Name       string
	VideoCodec string
}

type RTSPCLConf struct {
	Url string
}

type StrmConf struct {
	Pipes  map[int]PipeConfig
	RTSP   ServerConfig
	WebRTC ServerConfig
}

type ServerConfig struct {
	Address string
	State   string
}

func InitializeDefaultStrmConf() StrmConf {
	return StrmConf{
		RTSP: ServerConfig{
			Address: ":8554",
			State:   "stop",
		},
	}
}
