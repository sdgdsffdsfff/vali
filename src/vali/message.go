
package vali

import (
    "encoding/json"
    "fmt"
)

type Message struct {
    RunUser  string   `json:"runuser"`
    Command  string   `json:"command"`
    Timeout  int64    `json:"timeout"`
    Argument []string `json:"argument"`
    RunDir   string   `json:"rundir"`
}

func DecodeMessage(b []byte) (*Message, error) {
    m := &Message{}
    err := json.Unmarshal(b, m)
    return m, err
}

func Json(v interface{}) (s []byte) {
    s, _ = json.Marshal(v)
    length := []byte(fmt.Sprintf("%010d", len(s)))
    s = append(length, s...)
    return
}
