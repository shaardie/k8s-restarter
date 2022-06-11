package server

import "net/http"

// LoggerHandlerFunc is a logging middleware
func (s *Server) LoggerHandlerFunc(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recorder := &StatusRecorder{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}
		h.ServeHTTP(recorder, r)
		s.logger.Sugar().Infow("server request",
			"method", r.Method,
			"uri", r.RequestURI,
			"version", r.Proto,
			"status", recorder.Status,
			"remote address", r.RemoteAddr,
		)
	}
}

// StatusRecorder extends the ResponseWriter interface with a status code
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

// WriteHeader extends the ResponseWriter.WriteHeader to also safe the status
// code
func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}
