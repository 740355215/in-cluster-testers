package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var DelayTimeChan = make(chan int64, 1)

func WriteDelayTimeHandler(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)
	status := http.StatusOK

	vals := r.URL.Query()
	delay, ok := vals["seconds"]

	if !ok {
		res["result"] = "fail"
		res["errmsg"] = "required parameter delay is missing"
		status = http.StatusBadRequest
	} else {
		seconds, err := strconv.ParseInt(delay[0], 10, 64)
		if err != nil {
			res["result"] = "fail"
			res["errmsg"] = "required parameter delay must be int"
			status = http.StatusBadRequest
		} else {
			DelayTimeChan <- seconds
			res["result"] = "succ"
			res["seconds"] = delay[0]
			status = http.StatusOK
		}
	}

	response, _ := json.Marshal(res)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
