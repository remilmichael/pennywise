package main

type Flags struct {
	FrdReq     bool `json:"frdReq"`
	FrdAck     bool `json:"frdAck"`
	Billup     bool `json:"billup"`
	Billedit   bool `json:"billedit"`
	Billdelete bool `json:"billdel"`
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
}

type FrdSettle struct {
	ID       string
	NickName string
	Owns     string
	Owes     string
}

type ViewSettlement struct {
	NickName string
	Total    string
	Message  string
}

type toSentBill struct {
	id     string
	name   string
	amount string
}

//to push into queue
type BillUpload struct {
	Flags     `json:"flags"`
	SignMe    `json:"signme"`
	PubKey    []byte `json:"pubkey"`
	Signature []byte `json:"signature"`
}

//for signing the bill
type SignMe struct {
	UUID        string `json:"uuid"`
	HostID      string `json:"hostid"`
	PeerID      string `json:"peerid"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
	Date        string `json:"date"`
	DateAdded   string `json:"dateadd"`
}

//for saving to disk
type BillSave struct {
	PeerID      string
	Description string
	Amount      string
	Date        string
	DateAdded   string
}

//stores incoming friend requests
var reqBkt = []byte("frequest")

//store peerid to which a request is sent
var reqdump = []byte("frqdump")

//store friends
//key = hostid, value = byte(struct Friend)
var friendsBkt = []byte("friends")

//store items for peer forwarding
var queueName = []byte("sentqueue")

//contacts
var buddyBkt = []byte("buddies")

//store all friends(buckets)
//key = 8 byte random key, value = byte(struct FrdSettle)
var allBkts = []byte("buckets")
