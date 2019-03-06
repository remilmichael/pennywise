package main

type Flags struct {
	FrdReq bool `json:"frdReq"`
	FrdAck bool `json:"frdAck"`
}

type FrReqInd struct {
	Flags  `json:"flags"`
	HostID string `json:"hostid"`
	PeerID string `json:"peerid"`
}

type Receive struct {
	Flags `json:"flags"`
}

var reqBkt []byte = []byte("frequest")
