package main

type Flags struct {
	FrdReq bool
	FrdAck bool
}

type FrReqInd struct {
	HostID string
	PeerID string
	Flags
}
