package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var portStr string
var uploadDir string

//go:embed index.html
var formHTML []byte

//go:embed success.html
var successHTML []byte

func main() {
	parseArgs()
	http.HandleFunc("/", uploadForm)
	http.HandleFunc("/upload", uploadHandler)

	fmt.Printf(
		"\x1b[1m\x1b[38;5;159mhttp://localhost%s\n\x1b[38;5;158mhttp://%s%s\n\x1b[38;5;225mCtrl-C\x1b[0m to exit\n",
		portStr,
		getLocalAddr(),
		portStr,
	)

	if err := http.ListenAndServe(portStr, nil); err != nil {
		panic(err)
	}
}

func parseArgs() {
	port := flag.Int("port", 1080, "Port to listen")
	dir := flag.String("dir", "saved", "Directory to save uploaded files")
	flag.Parse()

	portStr = ":" + strconv.Itoa(*port)
	uploadDir = *dir
}

func uploadForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(formHTML)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	os.MkdirAll(uploadDir, 0755)

	// First, check if a file was uploaded
	file, handler, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		dst, err := os.Create(filepath.Join(uploadDir, filepath.Base(handler.Filename)))
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}
	} else {
		// If no file, check for pasted text
		pasted := r.FormValue("pastedText")
		if pasted == "" {
			http.Error(w, "No file or text provided", http.StatusBadRequest)
			return
		}

		filename := filepath.Join(uploadDir, fmt.Sprintf("pasted_%d.txt", time.Now().UnixNano()))
		if err := os.WriteFile(filename, []byte(pasted), 0644); err != nil {
			http.Error(w, "Error saving text", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(successHTML)
}

func getLocalAddr() string {
	addrs, err := net.InterfaceAddrs()

	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}
