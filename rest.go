package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type MangoHttpServer struct {
	config      *ServiceConfiguration
	store       FileStore
	identifiers map[string]string
}

func listenRequests(port string, store FileStore, config *ServiceConfiguration) error {
	server := &MangoHttpServer{store: store, config: config, identifiers: map[string]string{}}
	router := mux.NewRouter()
	router.HandleFunc("/upload", server.upload).Methods("POST")
	router.HandleFunc("/get/{id}", server.get)
	return http.ListenAndServe(":"+port, router)
}

//#region Download

type FileResponse struct {
	Status  string `json:"status"` // "ok" or "error"
	Error   string `json:"error,omitempty"`
	Present bool   `json:"present"`
	File    string `json:"file,omitempty"`
}

// Handles the download of previously stored files,
// must have an "id" parameter
func (server *MangoHttpServer) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	file, err := server.store.Read(id)

	var response *FileResponse
	if err != nil {
		// failed to read file from store
		response = &FileResponse{
			Status:  "error",
			Present: false,
			Error:   err.Error(),
		}
		fmt.Printf("[info] requested file '%s', but can't be get: %v\n", id, err)
	} else {
		data, err := io.ReadAll(file)
		if err != nil {
			// failed to read file bytes
			response = &FileResponse{
				Status:  "ok",
				Present: false,
				Error:   err.Error(),
			}
			fmt.Printf("[info] requested file '%s', but bytes can't be read: %v\n", id, err)
		} else {
			// Success!
			_ = file.Close()
			response = &FileResponse{
				Status:  "ok",
				Present: true,
				File:    base64.StdEncoding.EncodeToString(data),
			}

			// don't forget to remove file
			_, _ = server.store.Delete(id)
			fmt.Printf("[info] requested file '%s', downloaded\n", id)
		}
	}

	_ = json.NewEncoder(w).Encode(response)
}

// #endregion
// #region Upload

type FileUploadResponse struct {
	Ok    bool   `json:"ok"`
	Code  int    `json:"code"`
	Error string `json:"error,omitempty"`
	Id    string `json:"id,omitempty"`
}

func (server *MangoHttpServer) GetRemoteAddr(r *http.Request) string {
	if server.config.TrustProxy {
		forwardedFor := r.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			ips := strings.Split(forwardedFor, ", ")
			if len(ips) > 1 {
				// use the first address if we got an array
				return ips[0]
			}
			// otherwise just return the full string
			return forwardedFor
		}
	}
	// only use host
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func (server *MangoHttpServer) RemoteAddrToIdentifier(remoteAddr string) string {
	id, ok := server.identifiers[remoteAddr]
	if !ok {
		// generate id, note that ip cannot be
		// obtained from the id
		millis := time.Now().UnixMilli()
		id = fmt.Sprintf("%x", millis)
		server.identifiers[remoteAddr] = id
	}
	return id
}

func fail(w http.ResponseWriter, code int, error string) {
	fmt.Printf("[info] failed to upload file: %s (code %d)\n", error, code)
	response := &FileUploadResponse{
		Ok:    false,
		Code:  code,
		Error: error,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("[error] failed to write JSON error response: %v\n", err)
	}
}

func (server *MangoHttpServer) upload(w http.ResponseWriter, r *http.Request) {

	remoteAddr := server.GetRemoteAddr(r)
	id, hasId := server.identifiers[remoteAddr]

	if hasId {
		fail(w, 400, "Only one file per IP address")
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		fmt.Printf("[error] failed to parse multipart form data: %v\n", err)
	}

	files, ok := r.MultipartForm.File["file"]
	if !ok || files == nil {
		fail(w, 400, "No file specified")
		return
	}
	if len(files) != 1 {
		fail(w, 400, "Please specify a file (not more)")
		return
	}
	file := files[0]

	if file.Size > server.config.SizeLimit {
		fail(w, 413, fmt.Sprintf("File is too large (%d bytes), limit is %d", file.Size, server.config.SizeLimit))
		return
	}
	f, err := file.Open()
	if err != nil {
		fail(w, 400, fmt.Sprintf("Failed to open uploaded file: %v", err))
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		fail(w, 400, fmt.Sprintf("Failed to read file data: %v", err))
		return
	}
	if int64(len(data)) > server.config.SizeLimit {
		// did they just try to hack us ??
		fail(w, 413, fmt.Sprintf("File is too large (%d bytes), limit is %d (?)", len(data), server.config.SizeLimit))
		return
	}

	id = server.RemoteAddrToIdentifier(remoteAddr)
	err = server.store.Create(id, bytes.NewReader(data))
	if err != nil {
		fail(w, 400, fmt.Sprintf("Failed to save file: %v", err))
		return
	}

	response := &FileUploadResponse{
		Ok:   true,
		Code: 200,
		Id:   id,
	}
	json.NewEncoder(w).Encode(response)
	fmt.Printf("[info] uploaded file '%s'\n", id)

	// schedule deletion of file
	go func() {
		<-time.After(time.Duration(server.config.Lifetime) * time.Millisecond)
		delete(server.identifiers, remoteAddr)
		existed, _ := server.store.Delete(id)
		if existed {
			fmt.Printf("[info] expired file '%s'\n", id)
		}
	}()
}

//#endregion
