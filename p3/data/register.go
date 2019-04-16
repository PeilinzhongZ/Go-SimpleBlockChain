package data

import "encoding/json"

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {
	return RegisterData{id, peerMapJson}
}

func (data *RegisterData) EncodeToJson() (string, error) {
	str, error := json.Marshal(data)
	return string(str), error
}
