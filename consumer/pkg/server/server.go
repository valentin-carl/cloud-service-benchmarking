package server

import (
	"net/http"
)

type Server struct {
	ListenAddress string
	FileDir       string
}

// Remember to open ports on VM

func NewServer(listenAddr, fileDir string) *Server {
	return &Server{
		ListenAddress: listenAddr,
		FileDir:       fileDir,
	}
}

func (s *Server) Serve() error {
	http.Handle("/download/", http.StripPrefix(
		"/download/", http.FileServer(http.Dir(s.FileDir)),
	))
	return http.ListenAndServe(s.ListenAddress, nil)
}
