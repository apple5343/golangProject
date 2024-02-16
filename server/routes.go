package server

import "net/http"

func (s *Server) SetupRoutes() {
	fs := http.FileServer(http.Dir("static"))
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	s.router.Handle("/", s.HomeHandler())
	s.router.HandleFunc("/addTask", s.AddTask())
	s.router.HandleFunc("/getWorkersInfo", s.GetWorkersInfo())
	s.router.HandleFunc("/removeWorker", s.RemoveWorker())
	s.router.HandleFunc("/getDelays", s.GetDelays())
	s.router.HandleFunc("/updateDelays", s.UpdateDelays())
	s.router.HandleFunc("/getAllExpressions", s.GetAllTasks())
	s.router.HandleFunc("/getTask", s.GetTask())
}
