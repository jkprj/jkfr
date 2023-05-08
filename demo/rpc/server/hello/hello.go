package hello

type URequest struct {
	Name  string `json:"Name,omitempty"`
	Pause uint64 `json:"Pause,omitempty"`
}

type URespone struct {
	Msg   string `json:"Msg,omitempty"`
	Pause uint64 `json:"Pause,omitempty"`
}

type HelloRpc struct {
}

func NewHelloRpc() *HelloRpc {
	return &HelloRpc{}
}

func (h *HelloRpc) Hello(request URequest, response *URespone) error {
	*response = URespone{Msg: "hello baby, My name is jk", Pause: request.Pause}
	// time.Sleep(time.Second * 20)
	return nil
}

func (h *HelloRpc) HowAreYou(request URequest, response *URespone) error {
	*response = URespone{Msg: "fine, thank you, and you.", Pause: request.Pause}
	return nil
}

func (h *HelloRpc) WhatName(request URequest, response *URespone) (err error) {
	*response = URespone{Msg: "my name is LiLei, and you.", Pause: request.Pause}
	return nil
}
