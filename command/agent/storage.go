package agent

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Storage struct {
	options    *Options
	httpServer *http.Server
}

type Options struct {
	StoragePath   string
	ListenAddress string
}

func NewStorage(o *Options) *Storage {
	s := &Storage{
		options: o,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(s.handler))

	s.httpServer = &http.Server{
		Addr:    o.ListenAddress,
		Handler: mux,
	}

	return s
}

func (s *Storage) Serve() error {
	log.Printf("Storage serving on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Storage) handler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path
	if filename == "/" {
		files, err := s.List()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(files)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	filename = filename[1:]

	http.ServeFile(w, r, s.options.StoragePath+"/"+filename)
}

type File struct {
	Name string `json:"name"`
}

func (s *Storage) List() ([]*File, error) {
	files, err := ioutil.ReadDir(s.options.StoragePath)
	if err != nil {
		return nil, err
	}

	var result []*File
	for _, file := range files {
		result = append(result, &File{
			Name: file.Name(),
		})
	}
	return result, nil
}
