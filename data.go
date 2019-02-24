package main

type Flags struct {
	FrdReq bool
	FrdAck bool
}

type FrReqInd struct {
	flag   Flags
	HostID string
	PeerID string
}
