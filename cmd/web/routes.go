package main

import (
	"bookings/internal/config"
	"bookings/internal/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	// creating multiplexer
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/users/login", handlers.Repo.ShowLogin)
	mux.Post("/users/login", handlers.Repo.PostLogin)
	mux.Get("/users/logout", handlers.Repo.Logout)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)

	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		// mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashbord)
		mux.Get("/reservations/new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations/all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations/calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Get("/reservations/{src}/{id}", handlers.Repo.AdminShowReservation)
		mux.Get("/process-reservation/{src}/{id}", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservations/{src}/{id}", handlers.Repo.AdminDeleteReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminUpdateReservation)
	})

	return mux
}
