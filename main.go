package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const HOST string = "0.0.0.0" // allow access from outside container
const PORT string = ":8000"

type Payloads struct {
	Text     string `json:"content"`
	Path     string `json:"path"`
	FileName string `json:"filename"`
}

type Image struct {
	Image multipart.File
	Name  string
	Type  string
}

func (img *Image) checkType() bool {
	return img.Type == "image/png" || img.Type == "image/jpeg"
}

func (img *Image) getFullName() string {
	fileNamePrefix := fmt.Sprintf("%s", time.Now())
	fullName := fileNamePrefix + "_" + img.Name
	return fullName
}

func (img *Image) saveImage(path string) error {
	imagePath, err := os.Create(path)
	if err != nil {
		return err
	}
	defer imagePath.Close()
	_, err = io.Copy(imagePath, img.Image)
	if err != nil {
		return err
	}
	return nil
}

func (img *Image) detectImage(path string) ([]byte, error) {
	cmd := exec.Command("tesseract", path, "-", "-l", "jav")
	// Suppress stderr by redirecting it to /dev/null
	cmd.Stderr = nil
	return cmd.Output()
}

func main() {
	fmt.Println("Server run on", HOST+PORT)

	o := http.FileServer(http.Dir("./out"))
	http.Handle("/out/", http.StripPrefix("/out/", o))

	s := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", s))

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
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
			// since from the frontend, we send the payloads with multipart form
			if err := r.ParseMultipartForm(20); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			image := &Image{
				Type: r.FormValue("type"),
				Name: r.FormValue("name"),
			}
			img, _, err := r.FormFile("image")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			image.Image = img

			if b := image.checkType(); !b {
				http.Error(w, "Tipe file tidak didukung!", http.StatusUnsupportedMediaType)
				return
			}

			outImagePath := "./out/" + image.getFullName()
			err = image.saveImage(outImagePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			output, err := image.detectImage(outImagePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = os.Remove(outImagePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// TODO: periodically we must clean up the dir. try using linux cron jobs
			txtPath := outImagePath + ".txt"
			file, err := os.Create(txtPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()
			_, err = file.WriteString(string(output))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			payloads := &Payloads{
				Text:     string(output),
				Path:     txtPath,
				FileName: image.Name + ".txt",
			}
			p, err := json.Marshal(payloads)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Write(p)
		}
	})

	err := http.ListenAndServe(HOST+PORT, nil)
	if err != nil {
		panic(err)
	}
}
