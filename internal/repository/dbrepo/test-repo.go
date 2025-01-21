package dbrepo

import (
	"bookings/internal/config"
	"bookings/internal/models"
	"bookings/internal/repository"
	"database/sql"
	"errors"
	"time"
)

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewTestingsRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}

func (m *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for roomID, and false if no availability
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of availabile rooms, if any, for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

// GetRoomByID gets a room by ID
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("ID is greater then 2 error")
	}
	return room, nil
}

// GetUserByID returns a user by ID
func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var user models.User

	return user, nil
}

// UpdateUser updates a user in the database
func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

// Authenticate authenticates a user
func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 1, "", nil
}

// AllReservations returns a slice of all reservations
func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// NewReservations returns a slice of new reservations
func (m *testDBRepo) NewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// GetReservationByID returns the reservation extracted by id
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var reservation models.Reservation

	return reservation, nil
}
