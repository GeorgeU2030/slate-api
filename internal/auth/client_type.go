package auth

import "net/http"

type ClientType string

const (
	ClientWeb    ClientType = "web"
	ClientMobile ClientType = "mobile"
)

func GetClientType(r *http.Request) ClientType {
	header := r.Header.Get("X-Token-Delivery")
	if header == "mobile" {
		return ClientMobile
	}
	return ClientWeb
}