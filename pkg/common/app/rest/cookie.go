package rest

import (
	"net/http"
	"time"
)

var zeroTime = time.Unix(0, 0)

// DeleteCookie set the specified cookie to expired.
func DeleteCookie(w http.ResponseWriter, key, path, domain string) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    "",
		Path:     path,
		Domain:   domain,
		Expires:  zeroTime,
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}
