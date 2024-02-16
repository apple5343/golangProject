package server

import (
	"encoding/json"
	"html/template"
	"net/http"
)

var tpl, _ = template.ParseGlob("static/templates/*.html")

func (s *Server) AddTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		type Request struct {
			Task string `json:"task"`
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		task, err := s.Calculator.NewTask(req.Task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := s.db.AddTask(task.Expression, task.Created)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Id = id
		go task.Start()
		result := map[string]interface{}{"id": task.Id, "expression": task.Expression, "status": "processing"}
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func (s *Server) GetAllTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		result, err := s.db.GetAllTasks()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func (s *Server) GetWorkersInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		result := s.Calculator.Worker.GetWorkersInfo()
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func (s *Server) RemoveWorker() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		type Request struct {
			Id int `json:"id"`
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.Calculator.Worker.RemoveWorker(req.Id)
	}
}

func (s *Server) UpdateDelays() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		type Request struct {
			Delays map[string]int `json:"delays"`
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, v := range req.Delays {
			if v < 0 {
				http.Error(w, "Отрицательная задержка", http.StatusBadRequest)
				return
			}
		}
		err := s.db.UpdateDelays(req.Delays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result, err := s.db.GetDelays()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.Calculator.Worker.Delays = result
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func (s *Server) GetDelays() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := s.db.GetDelays()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func (s *Server) GetTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Request struct {
			Id int `json:"id"`
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := s.db.GetTaskById(req.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func (s *Server) HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.ExecuteTemplate(w, "index.html", nil)
	}
}
