package web

func (s *WebServer) Run(addr string) error {
	s.setupRoutes()
	return s.engine.Run(addr)
}
