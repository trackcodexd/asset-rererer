package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/trackcodexd/asset-rererer/internal/app/assets"
	"github.com/trackcodexd/asset-rererer/internal/app/request"
	"github.com/trackcodexd/asset-rererer/internal/app/response"
	"github.com/trackcodexd/asset-rererer/internal/color"
	"github.com/trackcodexd/asset-rererer/internal/files"
	"github.com/trackcodexd/asset-rererer/internal/roblox"
)

var CompatiblePluginVersion = ""

// makes a unique output filename for json logs
func getOutputFileName(reuploadType string) string {
	t := time.Now()
	return fmt.Sprintf("Output_%s_%s.json", reuploadType, t.Format("2006-01-02_15-04-05"))
}

// serve always binds on localhost:38073
func serve(c *roblox.Client) error {
	port := "38073"

	var exportedJSONName string
	var exportJSON bool
	var busy bool
	finished := true

	respHistory := make([]response.ResponseItem, 0)
	resp := response.New(func(i response.ResponseItem) {
		if exportJSON {
			respHistory = append(respHistory, i)

			j, err := json.Marshal(respHistory)
			if err != nil {
				log.Fatal(err)
			}

			if err := files.Write(exportedJSONName, string(j)); err != nil {
				log.Fatal(err)
			}
		}
	})

	// health check / data pull
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if resp.Len() == 0 && !busy {
			if !finished {
				finished = true
				busy = false
				exportJSON = false

				resp.Clear()
				respHistory = make([]response.ResponseItem, 0)

				fmt.Fprint(w, "done")
				fmt.Println("Finished reuploading. (you can rerun without restarting)")
			}
			return
		}

		if err := resp.EncodeJSON(json.NewEncoder(w)); err != nil {
			log.Fatal(err)
		} else {
			resp.Clear()
		}
	})

	// reupload handler
	http.HandleFunc("/reupload", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hit:", r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}

		if busy || !finished {
			fmt.Println("server busy, rejecting request")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		fmt.Println("body decoding start")
		var req request.RawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			color.Error.Println("decode error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("decoded request: %+v\n", req)

		// check version match
		if CompatiblePluginVersion != "" && req.PluginVersion != CompatiblePluginVersion {
			fmt.Println("incompatible plugin version")
			w.WriteHeader(http.StatusConflict)
			return
		}

		// make sure asset type is supported
		if exists := assets.DoesModuleExist(req.AssetType); !exists {
			fmt.Println("unknown asset type:", req.AssetType)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		startReupload, err := assets.NewReuploadHandlerWithType(req.AssetType, c, &req, resp)
		if err != nil {
			color.Error.Println("handler error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if exportJSON = req.ExportJSON; exportJSON {
			exportedJSONName = getOutputFileName(req.AssetType)
		}

		busy = true
		finished = false

		go func() {
			start := time.Now()
			err := startReupload()
			busy = false
			if err != nil {
				finished = true
				color.Error.Println("Failed to start reuploading:", err)
				return
			}

			duration := time.Since(start)
			fmt.Printf("Reuploading took %d hours, %d minutes, and %d seconds\n",
				int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
			fmt.Println("Waiting for client to finish changing ids...")
		}()

		w.WriteHeader(http.StatusOK)
	})

	// also catch trailing slash just in case
	http.HandleFunc("/reupload/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/reupload", http.StatusTemporaryRedirect)
	})

	fmt.Println("listening on http://localhost:" + port)
	// actually start the server
	return http.ListenAndServe(":"+port, nil)
}
