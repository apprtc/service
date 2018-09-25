// Copyright (c) 2014 The WebRTC project authors. All Rights Reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// Package collider implements a signaling server based on WebSocket.
package collider

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

const registerTimeoutSec = 10

// This is a temporary solution to avoid holding a zombie connection forever, by
// setting a 1 day timeout on reading from the WebSocket connection.
const wsReadTimeoutSec = 60 * 60 * 24

type Collider struct {
	*roomTable
	dash *dashboard
}

func NewCollider(rs string) *Collider {
	return &Collider{
		roomTable: newRoomTable(time.Second*registerTimeoutSec, rs),
		dash:      newDashboard(),
	}
}

// Run starts the collider server and blocks the thread until the program exits.
func (c *Collider) Run(p int, useTls bool) {
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("html"))))

	http.Handle("/ws", websocket.Handler(c.wsHandler))
	http.HandleFunc("/status", c.httpStatusHandler)
	http.HandleFunc("/room", c.httpHandler)
	http.HandleFunc("/join/", c.joinHandler)

	var e error

	pstr := ":" + strconv.Itoa(p)
	if useTls {
		config := &tls.Config{
			// Only allow ciphers that support forward secrecy for iOS9 compatibility:
			// https://developer.apple.com/library/prerelease/ios/technotes/App-Transport-Security-Technote/
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			},
			PreferServerCipherSuites: true,
		}
		server := &http.Server{Addr: pstr, Handler: nil, TLSConfig: config}

		e = server.ListenAndServeTLS("/cert/cert.pem", "/cert/key.pem")
	} else {
		e = http.ListenAndServe(pstr, nil)
	}

	if e != nil {
		log.Fatal("Run: " + e.Error())
	}
}

// httpStatusHandler is a HTTP handler that handles GET requests to get the
// status of collider.
func (c *Collider) httpStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET")

	rp := c.dash.getReport(c.roomTable)
	enc := json.NewEncoder(w)
	if err := enc.Encode(rp); err != nil {
		err = errors.New("Failed to encode to JSON: err=" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		c.dash.onHttpErr(err)
	}
}

// httpHandler is a HTTP handler that handles GET/POST/DELETE requests.
// POST request to path "/$ROOMID/$CLIENTID" is used to send a message to the other client of the room.
// $CLIENTID is the source client ID.
// The request must have a form value "msg", which is the message to send.
// DELETE request to path "/$ROOMID/$CLIENTID" is used to delete all records of a client, including the queued message from the client.
// "OK" is returned if the request is valid.
func (c *Collider) httpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, DELETE")

	fmt.Printf("url=%v", r.URL.Path)

	p := strings.Split(r.URL.Path, "/")
	if len(p) != 3 {
		c.httpError("Invalid path: "+html.EscapeString(r.URL.Path), w)
		return
	}
	rid, cid := p[1], p[2]

	switch r.Method {
	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.httpError("Failed to read request body: "+err.Error(), w)
			return
		}
		m := string(body)
		if m == "" {
			c.httpError("Empty request body", w)
			return
		}
		if err := c.roomTable.send(rid, cid, m); err != nil {
			c.httpError("Failed to send the message: "+err.Error(), w)
			return
		}
	case "DELETE":
		c.roomTable.remove(rid, cid)
	default:
		return
	}

	io.WriteString(w, "OK\n")
}

// Params Params
type Params struct {
	IsInitiator      string `json:"is_initiator"`
	RoomLink         string `json:"room_link"`
	ClientID         string `json:"client_id"`
	WsURL            string `json:"ws_url"`
	WssURL           string `json:"wss_url"`
	WsPostURL        string `json:"ws_post_url"`
	WssPostURL       string `json:"wss_post_url"`
	MediaConstraints string `json:"media_constraints"`
	IsLoopback       string `json:"is_loopback"`
	RoomID           string `json:"room_id"`
}

// {
//     "params": {
//         "is_initiator": "true",
//         "room_link": "https://appr.tc/r/aaaa1",
//         "version_info": "{\"gitHash\": \"20cdd7652d58c9cf47ef92ba0190a5505760dc05\", \"branch\": \"master\", \"time\": \"Fri Mar 9 17:06:42 2018 +0100\"}",
//         "messages": [],
//         "error_messages": [],
//         "client_id": "95834280",
//         "ice_server_transports": "",
//         "bypass_join_confirmation": "false",
//         "wss_url": "wss://apprtc-ws.webrtc.org:443/ws",
//         "media_constraints": "{\"audio\": true, \"video\": true}",
//         "include_loopback_js": "",
//         "is_loopback": "false",
//         "offer_options": "{}",
//         "pc_constraints": "{\"optional\": []}",
//         "pc_config": "{\"rtcpMuxPolicy\": \"require\", \"bundlePolicy\": \"max-bundle\", \"iceServers\": []}",
//         "wss_post_url": "https://apprtc-ws.webrtc.org:443",
//         "ice_server_url": "https://networktraversal.googleapis.com/v1alpha/iceconfig?key=AIzaSyAJdh2HkajseEIltlZ3SIXO02Tze9sO3NY",
//         "warning_messages": [],
//         "room_id": "aaaa1",
//         "include_rtstats_js": "<script src=\"/js/rtstats.js\"></script><script src=\"/pako/pako.min.js\"></script>"
//     },
//     "result": "SUCCESS"
// }

// RoomInfo RoomInfo
type RoomInfo struct {
	Params Params `json:"params"`
	Result string `json:"result"`
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func (c *Collider) joinHandler(w http.ResponseWriter, r *http.Request) {
	var room RoomInfo
	room.Result = "SUCCESS"

	var params Params
	params.IsInitiator = "true"

	fmt.Printf("url=%v", r.URL.Path)

	p := strings.Split(r.URL.Path, "/")
	fmt.Printf("url=%s", strings.Join(p, ","))
	if len(p) < 3 {
		c.httpError("Invalid path: "+html.EscapeString(r.URL.Path), w)
		return
	}
	rid := p[2]
	cid := ""

	if len(p) >= 3 {
		cid = p[3]
	}
	if cid == "" {
		cid = RandStringBytesMaskImprSrc(10)
	}

	params.RoomLink = fmt.Sprintf("%s/r/%s", r.Host, rid)
	params.ClientID = cid
	params.WsURL = fmt.Sprintf("ws://%s/ws", r.Host)
	params.WssURL = fmt.Sprintf("ws://%s/ws", r.Host)
	params.WsPostURL = fmt.Sprintf("http://%s", r.Host)
	params.WssPostURL = fmt.Sprintf("http://%s", r.Host)
	params.MediaConstraints = "{\"audio\": true, \"video\": true}"
	params.IsLoopback = "false"
	params.RoomID = rid

	room.Params = params

	enc := json.NewEncoder(w)
	if err := enc.Encode(room); err != nil {
		err = errors.New("Failed to encode to JSON: err=" + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		c.dash.onHttpErr(err)
	}
}

// wsHandler is a WebSocket server that handles requests from the WebSocket client in the form of:
// 1. { 'cmd': 'register', 'roomid': $ROOM, 'clientid': $CLIENT' },
// which binds the WebSocket client to a client ID and room ID.
// A client should send this message only once right after the connection is open.
// or
// 2. { 'cmd': 'send', 'msg': $MSG }, which sends the message to the other client of the room.
// It should be sent to the server only after 'regiser' has been sent.
// The message may be cached by the server if the other client has not joined.
//
// Unexpected messages will cause the WebSocket connection to be closed.
func (c *Collider) wsHandler(ws *websocket.Conn) {
	var rid, cid string

	registered := false

	var msg wsClientMsg
loop:
	for {
		err := ws.SetReadDeadline(time.Now().Add(time.Duration(wsReadTimeoutSec) * time.Second))
		if err != nil {
			c.wsError("ws.SetReadDeadline error: "+err.Error(), ws)
			break
		}

		err = websocket.JSON.Receive(ws, &msg)
		if err != nil {
			if err.Error() != "EOF" {
				c.wsError("websocket.JSON.Receive error: "+err.Error(), ws)
			}
			break
		}

		switch msg.Cmd {
		case "register":
			if registered {
				c.wsError("Duplicated register request", ws)
				break loop
			}
			if msg.RoomID == "" || msg.ClientID == "" {
				c.wsError("Invalid register request: missing 'clientid' or 'roomid'", ws)
				break loop
			}
			if err = c.roomTable.register(msg.RoomID, msg.ClientID, ws); err != nil {
				c.wsError(err.Error(), ws)
				break loop
			}
			registered, rid, cid = true, msg.RoomID, msg.ClientID
			c.dash.incrWs()

			defer c.roomTable.deregister(rid, cid)
			break
		case "send":
			if !registered {
				c.wsError("Client not registered", ws)
				break loop
			}
			if msg.Msg == "" {
				c.wsError("Invalid send request: missing 'msg'", ws)
				break loop
			}
			c.roomTable.send(rid, cid, msg.Msg)
			break
		default:
			c.wsError("Invalid message: unexpected 'cmd'", ws)
			break
		}
	}
	// This should be unnecessary but just be safe.
	ws.Close()
}

func (c *Collider) httpError(msg string, w http.ResponseWriter) {
	err := errors.New(msg)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	c.dash.onHttpErr(err)
}

func (c *Collider) wsError(msg string, ws *websocket.Conn) {
	err := errors.New(msg)
	sendServerErr(ws, msg)
	c.dash.onWsErr(err)
}
