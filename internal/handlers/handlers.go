package handlers

import (
	"bookings/internal/config"
	"bookings/internal/driver"
	"bookings/internal/forms"
	"bookings/internal/helpers"
	"bookings/internal/models"
	"bookings/internal/render"
	"bookings/internal/repository"
	"bookings/internal/repository/dbrepo"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingsRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(reservation.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "cannot find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", reservation)

	stringMap := make(map[string]string)
	stringMap["start_date"] = helpers.ConvertDateToString(reservation.StartDate)
	stringMap["end_date"] = helpers.ConvertDateToString(reservation.EndDate)

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("cannot get reservation from session"))
		return
	}

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
	}

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
	}

	// send notifications
	htmlMessage := fmt.Sprintf(`
		<p><strong>Reservation Confirmation</strong><br/></p>
		<p>Dear %s, <br/> This is a confirmation of your booking from %s to %s.</p>
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "me@here.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability handles post
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	startDate, err := helpers.ConvertStringToDate(start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := helpers.ConvertStringToDate(end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		// No availability
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// ChooseRoom extracts roomID parameter from request, puts it to the session "reservation" object
// and redirects to /make-reservation page
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID
	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL parameters, builds a session variable, and takes user to make res screen
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {

	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	startDate, err := helpers.ConvertStringToDate(r.URL.Query().Get("sd"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := helpers.ConvertStringToDate(r.URL.Query().Get("ed"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var reservation models.Reservation

	reservation.RoomID = roomID
	reservation.StartDate = startDate
	reservation.EndDate = endDate
	reservation.Room.RoomName = room.RoomName
	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	startDate, err := helpers.ConvertStringToDate(sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := helpers.ConvertStringToDate(ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomId, _ := strconv.Atoi(r.Form.Get("room_id"))

	isRoomAvailable, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomId)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var message string
	if isRoomAvailable {
		message = "Available!"
	} else {
		message = "Not Available!"
	}

	resp := jsonResponse{
		OK:        isRoomAvailable,
		Message:   message,
		RoomID:    strconv.Itoa(roomId),
		StartDate: sd,
		EndDate:   ed,
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the res summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get reservation from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	stringMap := make(map[string]string)
	stringMap["start_date"] = helpers.ConvertDateToString(reservation.StartDate)
	stringMap["end_date"] = helpers.ConvertDateToString(reservation.EndDate)

	data := make(map[string]interface{})
	data["reservation"] = reservation

	m.App.Session.Remove(r.Context(), "reservation")

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ShowLogin renders the login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostLogin handles logging the user in
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PostLogin logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// AdminDashbord renders the admin-dashboard page
func (m *Repository) AdminDashbord(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminNewReservations shows new reservation in admin dashboard page
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.NewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations shows all reservation in admin dashboard page
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation renders the Reservation form
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	// exploded := strings.Split(r.RequestURI, "/")

	res, err := m.DB.GetReservationByID(resID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	stringMap := make(map[string]string)
	stringMap["src"] = chi.URLParam(r, "src")
	stringMap["start_date"] = helpers.ConvertDateToString(res.StartDate)
	stringMap["end_date"] = helpers.ConvertDateToString(res.EndDate)

	render.Template(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
		Data:      data,
		Form:      forms.New(nil),
		StringMap: stringMap,
	})
}

// AdminUpdateReservation updates the reservation
func (m *Repository) AdminUpdateReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation, err := m.DB.GetReservationByID(resID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}
	err = m.DB.UpdateReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%s", chi.URLParam(r, "src")), http.StatusSeeOther)
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	processed, err := strconv.Atoi(r.URL.Query().Get("processed"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = m.DB.UpdateProcessedForReservation(resID, processed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%s", chi.URLParam(r, "src")), http.StatusSeeOther)
}

// AdminDeleteReservation deletes a reservation by id
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	resID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	err = m.DB.DeleteReservation(resID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations/%s", chi.URLParam(r, "src")), http.StatusSeeOther)
}

// AdminReservationsCalendar displays the reservations calendar
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{})
}
