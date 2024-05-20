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
	Syncs  []string
	State  string
	RTPR   RTPRConf
	RTPS   RTPSConf
	Type   string
}

type RTPSConf struct {
	VideoPort  int
	AudioPort  int
	Name       string
	VideoCodec string
}

type StrmConf struct {
	Pipes  map[int]PipeConfig
	RTSP   ServerConfig
	WebRTC ServerConfig
}

type ServerConfig struct {
	Port  int
	State string
}

func InitializeDefaultStrmConf() StrmConf {
	return StrmConf{ /*
			Pipes: map[int]PipeConfig{
				1: {ID: 1,
					Name:   "pipe1",
					Source: "RTPR",
					RTPR: RTPRConf{VideoURL: "rtp://127.0.0.1:5000",
						AudioURL:   "rtp://127.0.0.1:5002",
						RunOnReady: "ffmpeg -re -stream_loop -1 -i videos/ts1920x1080h264.mp4 -an -c:v libx264 -f rtp rtp://127.0.0.1:5000 -vn -c:a copy -f  rtp rtp://127.0.0.1:5002",
						VideoCodec: "h264"}, Syncs: []string{"sync1"},
					State: "active"},
				2: {ID: 2,
					Name:   "pipe2",
					Source: "RTPR",
					RTPR: RTPRConf{VideoURL: "rtp://127.0.0.1:5010",
						AudioURL:   "rtp://127.0.0.1:5012",
						RunOnReady: "ffmpeg -re -stream_loop -1 -i videos/ts1920x1080h264.mp4 -an -c:v libx264 -f rtp rtp://127.0.0.1:5000 -vn -c:a copy -f  rtp rtp://127.0.0.1:5002",
						VideoCodec: "h264"}, Syncs: []string{"sync1"},
					State: "active"},
				3: {ID: 3,
					Name:   "pipe3",
					Source: "RTPR",
					RTPR: RTPRConf{VideoURL: "rtp://127.0.0.1:5020",
						AudioURL:   "rtp://127.0.0.1:5022",
						RunOnReady: "ffmpeg -re -stream_loop -1 -i videos/ts1920x1080h264.mp4 -an -c:v libx264 -f rtp rtp://127.0.0.1:5000 -vn -c:a copy -f  rtp rtp://127.0.0.1:5002",
						VideoCodec: "h264"}, Syncs: []string{"sync1"},
					State: "active"},
				4: {ID: 4,
					Name:   "pipe4",
					Source: "RTPR",
					RTPR: RTPRConf{VideoURL: "rtp://127.0.0.1:5030",
						AudioURL:   "rtp://127.0.0.1:50312",
						RunOnReady: "ffmpeg -re -stream_loodel p -1 -i videos/ts1920x1080h264.mp4 -an -c:v libx264 -f rtp rtp://127.0.0.1:5000 -vn -c:a copy -f  rtp rtp://127.0.0.1:5002",
						VideoCodec: "h264"}, Syncs: []string{"sync1"},
					State: "active"},
			},
			RTSP:   ServerConfig{Port: 8888, State: "active"},
			WebRTC: ServerConfig{Port: 8880, State: "active"},
		*/
	}
}
