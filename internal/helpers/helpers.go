package helpers

import (
	"bookings/internal/config"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ConvertStringToDate(str_date string) (time.Time, error) {
	// 2006-01-02 -> 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"
	date, err := time.Parse(layout, str_date)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
