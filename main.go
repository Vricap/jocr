package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	// fs := http.FileServer(http.Dir("static"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			data := map[string]string{
				"Title": "JOCR",
			}
			tmpl.Execute(w, data)
		}
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if err := r.ParseMultipartForm(20); err != nil { // since from the frontend, we send the payloads with multipart form
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fileType := r.FormValue("type")
			fileName := r.FormValue("name")
			image, _, err := r.FormFile("image")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			if fileType != "image/png" && fileType != "image/jpeg" {
				http.Error(w, "Tipe file tidak didukung!", http.StatusUnsupportedMediaType)
				return
			}
			fileNamePrefix := fmt.Sprintf("%s", time.Now())
			fileName = fileNamePrefix + fileName
			fullPath := "./out/" + fileName

			path, err := os.Create(fullPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer path.Close()
			_, err = io.Copy(path, image)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			cmd := exec.Command("tesseract", fullPath, "-", "-l", "jav", "2>/dev/null")
			// Suppress stderr by redirecting it to /dev/null
			cmd.Stderr = nil // or:
			cmd.Stderr, _ = os.Open(os.DevNull)
			output, err := cmd.Output()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = os.Remove(fullPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(output)
		}
	})

	err := http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		panic(err)
	}
}
