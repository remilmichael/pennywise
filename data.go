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

type Friend struct {
	ID       string
	NickName string
	Pubkey   string
}

//stores incoming friend requests
var reqBkt = []byte("frequest")

//store peerid to which a request is sent
var reqdump = []byte("frqdump")

//store friends
var friendsBkt = []byte("friends")

//store items for peer forwarding
var queueName = []byte("sentqueue")

//contacts
var bucketName = []byte("buddies")
