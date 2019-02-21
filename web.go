// go build server.go frame.go robot.go egg.go parseConfig.go client.go web
package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"time"
)

func webStart(addr string) {
	if addr == "" {
		return
	}
	Vf(2, "Web Listening: %v\n\n", addr)

	cmdCh := make(chan string, 1)
	backlogCh := make(chan struct{}, 2)
	go func() {
		for {
			<-backlogCh

			cmd := <-cmdCh
			//Vln(4, "[web][cmd]worker", cmd)
			reloadConfig(cmd)
		}
	}()
	go func() {
		for {
			select {
			case backlogCh <-struct{}{}:
			default:
				goto START
			}
		}

	START:
		Vln(4, "[web][backlog]", len(backlogCh), cap(backlogCh))
		for {
			backlogCh <-struct{}{}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	http.HandleFunc("/api/", func (w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "HEAD":
		case "GET":
			get(w, r)
		case "POST", "PUT":
			set(w, r, cmdCh)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/debug/", http.StripPrefix("/debug/", http.FileServer(http.Dir("./www"))))
	http.ListenAndServe(*webAddr, nil)
}

func get(w http.ResponseWriter, r *http.Request) {
	var data interface{}

	switch r.URL.Path {
	case "/api/user":
		data = user
	case "/api/egg":
		data = eggPool
	case "/api/bot":
		data = grid
	default:
	}
	Vln(4, "[web][list]", r.URL.Path, data)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		Vln(3, "[web][err]", err)
	}
}

func set(w http.ResponseWriter, r *http.Request, cmdCh chan string) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	switch r.URL.Path {
	case "/api/user":
		err = json.Unmarshal(buf, &user)
		grid.SetName(user.Name)
		grid.SetGoPos(int(user.GO))
		grid.SetPageCount(user.PageCount)
		user.PageCount = grid.PageCount // for PageCount*6 < len(robot)
		grid.GP = user.GP
	case "/api/egg":
		err = json.Unmarshal(buf, &eggPool)
	case "/api/bot":
		err = json.Unmarshal(buf, &grid)
		user.GO = int(grid.GO)
		user.PageCount = grid.PageCount
		user.GP = grid.GP
	case "/api/do":
		cmd := ""
		err = json.Unmarshal(buf, &cmd)
		if err == nil {
			select {
			case cmdCh <- cmd:
			default:
			}
		}
	default:
		return
	}
	Vln(4, "[web][put]", r.URL.Path, err)
	if err == nil {
		clients.Flush()
		return
	}
	http.Error(w, "bad request", http.StatusBadRequest)
}
