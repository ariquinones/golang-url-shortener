package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type yamlRedirectURL struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

type jsonRedirectURL struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

func main() {
	yamlFileName := flag.String("yaml-urls", "yaml-urls.yaml", "YAML format url shorts")
	jsonFileName := flag.String("json-urls", "json-urls.json", "JSON format url shorts")
	flag.Parse()

	// Open and parse json data
	jsonFile, jsonOpenE := os.Open(*jsonFileName)
	if jsonOpenE != nil {
		exitProgram(fmt.Sprintf("Error opening file %s %s", *jsonFileName, jsonOpenE.Error()))
	}
	defer jsonFile.Close()

	jsonData, jsonReadE := ioutil.ReadAll(jsonFile)
	if jsonReadE != nil {
		exitProgram(fmt.Sprintf("Error reading file %s %s", *jsonFileName, jsonReadE.Error()))
	}

	// Create default mux
	mux := defaultMux()

	jsonHandler, jsonMapE := mapJsonHandler(jsonData, mux)
	if jsonMapE != nil {
		exitProgram(fmt.Sprintf("Error creating json handler %s", jsonMapE.Error()))
	}

	yamlFile, yamlOpenE := os.Open(*yamlFileName)
	if yamlOpenE != nil {
		exitProgram(fmt.Sprintf("Error opening file %s %s", *yamlFileName, yamlOpenE.Error()))
	}
	defer yamlFile.Close()

	yamlData, yamlReadE := ioutil.ReadAll(yamlFile)
	if yamlReadE != nil {
		exitProgram(fmt.Sprintf("Error reading file %s %s", *yamlFileName, yamlReadE.Error()))
	}

	yamlHandler, yamlMapE := mapYamlHandler(yamlData, jsonHandler)
	if yamlMapE != nil {
		exitProgram(fmt.Sprintf("Error creating yaml handler %s", yamlMapE.Error()))
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", yamlHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	return mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func exitProgram(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func mapHandler(urlMap map[string]string, fallbackHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		if dest, ok := urlMap[urlPath]; ok {
			http.Redirect(w, r, dest, http.StatusFound)
		}
		fallbackHandler.ServeHTTP(w, r)
	}
}

func mapJsonHandler(jsonData []byte, fallbackHandler http.Handler) (http.HandlerFunc, error) {
	parsedJsonUrls, err := parseJson(jsonData)
	if err != nil {
		return nil, err
	}
	urlMap := mapJsonUrls(parsedJsonUrls)
	return mapHandler(urlMap, fallbackHandler), nil
}

func parseJson(jsonData []byte) ([]jsonRedirectURL, error) {
	var urlRedirects []jsonRedirectURL
	parseError := json.Unmarshal(jsonData, &urlRedirects)
	if parseError != nil {
		return nil, parseError
	}
	return urlRedirects, nil
}

func mapJsonUrls(urlRedirects []jsonRedirectURL) map[string]string {
	urlMap := make(map[string]string)
	for _, url := range urlRedirects {
		urlMap[url.Path] = url.URL
	}
	return urlMap
}

func mapYamlHandler(yamlData []byte, fallbackHandler http.Handler) (http.HandlerFunc, error) {
	parsedYamlUrls, err := parseYaml(yamlData)
	if err != nil {
		return nil, err
	}
	urlMap := mapYamlUrls(parsedYamlUrls)
	return mapHandler(urlMap, fallbackHandler), nil
}

func parseYaml(yamlData []byte) ([]yamlRedirectURL, error) {
	var urlRedirects []yamlRedirectURL
	parseErr := yaml.Unmarshal(yamlData, &urlRedirects)
	if parseErr != nil {
		return nil, parseErr
	}
	return urlRedirects, nil
}

func mapYamlUrls(yamlURLRedirectSlice []yamlRedirectURL) map[string]string {
	urlMap := make(map[string]string)
	for _, ur := range yamlURLRedirectSlice {
		urlMap[ur.Path] = ur.URL
	}
	return urlMap
}
