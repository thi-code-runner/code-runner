package server

import (
	"code-runner/services/codeRunner"
	"fmt"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Server struct {
	mux  *http.ServeMux
	port int
	addr string

	CodeRunner *codeRunner.Service
}

func NewServer(port int, addr string) (*Server, error) {
	if port < 1024 && port > 49151 {
		return nil, fmt.Errorf("validation error: port must be between [1024;49151] but was %d", port)
	}
	if len(addr) <= 0 {
		addr = "localhost"
	}
	return &Server{mux: &http.ServeMux{}, port: port, addr: addr}, nil
}

func (s *Server) Run() {
	fmt.Printf("\n\u2584\u2584\u2584\u2584\u2584\u2584\u2584 \u2584\u2584\u2584\u2584\u2584\u2584\u2584 \u2584\u2584\u2584\u2584\u2584\u2584  \u2584\u2584\u2584\u2584\u2584\u2584\u2584    \u2584\u2584\u2584\u2584\u2584\u2584   \u2584\u2584   \u2584\u2584 \u2584\u2584    \u2584 \u2584\u2584    \u2584 \u2584\u2584\u2584\u2584\u2584\u2584\u2584 \u2584\u2584\u2584\u2584\u2584\u2584   \n\u2588       \u2588       \u2588      \u2588\u2588       \u2588  \u2588   \u2584  \u2588 \u2588  \u2588 \u2588  \u2588  \u2588  \u2588 \u2588  \u2588  \u2588 \u2588       \u2588   \u2584  \u2588  \n\u2588       \u2588   \u2584   \u2588  \u2584    \u2588    \u2584\u2584\u2584\u2588  \u2588  \u2588 \u2588 \u2588 \u2588  \u2588 \u2588  \u2588   \u2588\u2584\u2588 \u2588   \u2588\u2584\u2588 \u2588    \u2584\u2584\u2584\u2588  \u2588 \u2588 \u2588  \n\u2588     \u2584\u2584\u2588  \u2588 \u2588  \u2588 \u2588 \u2588   \u2588   \u2588\u2584\u2584\u2584   \u2588   \u2588\u2584\u2584\u2588\u2584\u2588  \u2588\u2584\u2588  \u2588       \u2588       \u2588   \u2588\u2584\u2584\u2584\u2588   \u2588\u2584\u2584\u2588\u2584 \n\u2588    \u2588  \u2588  \u2588\u2584\u2588  \u2588 \u2588\u2584\u2588   \u2588    \u2584\u2584\u2584\u2588  \u2588    \u2584\u2584  \u2588       \u2588  \u2584    \u2588  \u2584    \u2588    \u2584\u2584\u2584\u2588    \u2584\u2584  \u2588\n\u2588    \u2588\u2584\u2584\u2588       \u2588       \u2588   \u2588\u2584\u2584\u2584   \u2588   \u2588  \u2588 \u2588       \u2588 \u2588 \u2588   \u2588 \u2588 \u2588   \u2588   \u2588\u2584\u2584\u2584\u2588   \u2588  \u2588 \u2588\n\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588  \u2588\u2584\u2584\u2584\u2588  \u2588\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2588  \u2588\u2584\u2584\u2588\u2584\u2588  \u2588\u2584\u2584\u2588\u2584\u2584\u2584\u2584\u2584\u2584\u2584\u2588\u2584\u2584\u2584\u2588  \u2588\u2584\u2588\n")
	log.Printf("starting code-runner on %s\n", fmt.Sprintf("%s:%d", s.addr, s.port))
	s.initRoutes()
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.addr, s.port), s.mux))
}

func (s *Server) initRoutes() {
	s.mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			// ...
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		for {
			var v interface{}
			err = wsjson.Read(r.Context(), c, &v)
			if err != nil {
				// ...
				log.Printf("%s\n", err)
				return
			}
			log.Printf("received: %v", v)
		}

		c.Close(websocket.StatusNormalClosure, "")
	})
}
