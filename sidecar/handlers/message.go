package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const MsgSize = 1024

var Message string

func WriteMessageHandler(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)
	status := http.StatusOK

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		res["result"] = "fail"
		res["errmsg"] = "read request body failed"
		status = http.StatusInternalServerError
	} else {
		if len(body) <= 1024 {
			Message = string(body)
		} else {
			Message = string(body[:MsgSize])
		}
		res["result"] = "succ"
		status = http.StatusOK
	}

	response, _ := json.Marshal(res)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func ReadMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(Message))
}
