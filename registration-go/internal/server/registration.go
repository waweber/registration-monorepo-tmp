package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"registration/internal/batch"
	"registration/internal/db"
	"registration/internal/registration"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

type registrationRequest struct {
	Registration registration.Registration `json:"registration"`
}

type registrationResponse struct {
	Registration registration.Registration `json:"registration"`
}

type registrationUpdateErrorResponse struct {
	Errors []batch.ChangeError `json:"errors"`
}

func (s *Server) HandleCreateRegistration(w http.ResponseWriter, req *http.Request) {
	eventId := req.PathValue("eventId")

	regReq := registrationRequest{
		Registration: registration.Registration{
			RegistrationFields: registration.RegistrationFields{
				Status:  registration.STATUS_PENDING,
				Options: make(registration.Options, 0),
			},
		},
	}
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&regReq)
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	reg := regReq.Registration

	if !registration.ValidateStatus(reg.Status) {
		httpError(w, http.StatusUnprocessableEntity)
		return
	}

	db, err := db.NewDBSession(req.Context(), s.pool)
	if err != nil {
		serverError(w, err)
		return
	}
	defer db.Rollback(req.Context())

	reg.Id = registration.NewRegistrationId()
	reg.EventId = eventId
	reg.DateCreated = time.Now()
	reg.Version = 1

	err = db.AddRegistration(req.Context(), reg)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity)
		return
	}

	err = db.Commit(req.Context())
	if err != nil {
		serverError(w, err)
		return
	}

	w.Header().Add("ETag", getEtag(reg))

	jsonResponse(w, registrationResponse{Registration: reg})
}

func (s *Server) HandleGetRegistration(w http.ResponseWriter, req *http.Request) {
	eventId := req.PathValue("eventId")
	registrationId := req.PathValue("registrationId")

	db, err := db.NewDBSession(req.Context(), s.pool)
	if err != nil {
		serverError(w, err)
		return
	}
	defer db.Rollback(req.Context())

	reg, err := db.GetRegistration(req.Context(), eventId, registrationId, false)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, req)
		return
	}

	w.Header().Add("ETag", getEtag(reg))

	jsonResponse(w, registrationResponse{Registration: reg})
}

func (s *Server) HandleUpdateRegistration(w http.ResponseWriter, req *http.Request) {
	eventId := req.PathValue("eventId")
	registrationId := req.PathValue("registrationId")

	regReq := registrationRequest{
		Registration: registration.Registration{
			RegistrationFields: registration.RegistrationFields{
				Status:  registration.STATUS_PENDING,
				Options: make(registration.Options, 0),
			},
		},
	}
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&regReq)
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	reg := regReq.Registration

	if !registration.ValidateStatus(reg.Status) {
		httpError(w, http.StatusUnprocessableEntity)
		return
	}

	reg.Id = registrationId
	reg.EventId = eventId

	db, err := db.NewDBSession(req.Context(), s.pool)
	if err != nil {
		serverError(w, err)
		return
	}
	defer db.Rollback(req.Context())

	curReg, err := db.GetRegistration(req.Context(), eventId, registrationId, true)
	if err != nil {
		serverError(w, err)
		return
	}

	change := &batch.Change{
		Old: curReg,
		New: reg,
	}

	testRes := change.Test()

	if slices.IndexFunc(testRes.Errors, func(err batch.ChangeError) bool { return err.Code == batch.ERROR_VERSION }) != -1 {
		jsonResponseStatus(w, http.StatusConflict, registrationUpdateErrorResponse{Errors: testRes.Errors})
		return
	}

	if len(testRes.Errors) > 0 {
		jsonResponseStatus(w, http.StatusUnprocessableEntity, registrationUpdateErrorResponse{Errors: testRes.Errors})
		return
	}

	newReg, err := db.UpdateRegistration(req.Context(), reg)
	if err != nil {
		serverError(w, err)
		return
	}

	err = db.Commit(req.Context())
	if err != nil {
		serverError(w, err)
		return
	}

	w.Header().Add("ETag", getEtag(newReg))

	jsonResponse(w, registrationResponse{Registration: newReg})
}

func getEtag(r registration.Registration) string {
	return fmt.Sprintf("W/\"%d\"", r.Version)
}
