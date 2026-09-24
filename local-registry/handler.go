package local_registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func (s *LocalRegistryServer) AddRoute(uri, method string, f func(...string) any) {
	s.serverMux.HandleFunc(fmt.Sprintf("%s /%s", strings.ToUpper(method), strings.TrimLeft(uri, "/")), func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		re := regexp.MustCompile(`\{([^}]+)\}`)
		var varNames []string
		for _, match := range re.FindAllStringSubmatch(uri, -1) {
			varNames = append(varNames, req.PathValue(match[1]))
		}
		res := f(varNames...)
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			return
		}
	})
}

func (s *LocalRegistryServer) AddFileRoute(uri, method, filename string) {
	s.serverMux.HandleFunc(fmt.Sprintf("%s /%s", strings.ToUpper(method), strings.TrimLeft(uri, "/")), func(writer http.ResponseWriter, req *http.Request) {
		absAssetPath := s.absoluteAssetPath
		if uri == "" {
			pwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			absAssetPath = pwd
		}

		reqFilename := req.PathValue(filename)

		binary, err := os.ReadFile(fmt.Sprintf("%s/assets/binaries/%s", absAssetPath, reqFilename))
		if err != nil {
			panic(err)
		}
		_, err = writer.Write(binary)
		if err != nil {
			return
		}

		// http.ServeFile()

		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; finame="%s"`, reqFilename))

		// http.ServeFile(writer, req, reqFilename)
	})
}
