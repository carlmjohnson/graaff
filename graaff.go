package graaff

import "net/http"

func Must(h http.Handler, m map[string]string, err error) (http.Handler, map[string]string) {
	if err != nil {
		panic(err.Error())
	}
	return h, m
}

func Handle(filepath, baseurl string) (http.Handler, map[string]string, error) {
	h := http.StripPrefix(baseurl, http.FileServer(http.Dir(filepath)))
	return h, nil, nil
}
