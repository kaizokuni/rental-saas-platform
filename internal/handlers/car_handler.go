package handlers

import (
	"encoding/json"
	"net/http"
	"rental-saas/internal/models"
	"rental-saas/internal/repository"
	"rental-saas/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CarHandler struct {
	Repo    *repository.CarRepository
	Storage *storage.MinioClient
}

func NewCarHandler(repo *repository.CarRepository, storage *storage.MinioClient) *CarHandler {
	return &CarHandler{Repo: repo, Storage: storage}
}

func (h *CarHandler) CreateCar(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form for image upload
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	make := r.FormValue("make")
	model := r.FormValue("model")
	licensePlate := r.FormValue("license_plate")
	status := r.FormValue("status")

	var imageURL string
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		// Upload to MinIO
		contentType := header.Header.Get("Content-Type")
		url, err := h.Storage.UploadFile(r.Context(), header.Filename, file, header.Size, contentType)
		if err != nil {
			http.Error(w, "Failed to upload image: "+err.Error(), http.StatusInternalServerError)
			return
		}
		imageURL = url
	}

	car := &models.Car{
		Make:         make,
		Model:        model,
		LicensePlate: licensePlate,
		Status:       models.CarStatus(status),
		ImageURL:     imageURL,
	}

	if err := h.Repo.Create(r.Context(), car); err != nil {
		http.Error(w, "Failed to create car: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   car,
	})
}

func (h *CarHandler) ListCars(w http.ResponseWriter, r *http.Request) {
	cars, err := h.Repo.List(r.Context())
	if err != nil {
		http.Error(w, "Failed to list cars: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   cars,
	})
}

func (h *CarHandler) UpdateCar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid car ID", http.StatusBadRequest)
		return
	}

	// Get current car state
	currentCar, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Car not found", http.StatusNotFound)
		return
	}

	// Parse form data
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	newStatus := models.CarStatus(r.FormValue("status"))

	// State Machine Validation
	if currentCar.Status == models.CarStatusRented && newStatus == models.CarStatusAvailable {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"code":    "ERR_INVALID_TRANSITION",
			"message": "Car cannot go from Rented to Available. Must be Inspected first.",
		})
		return
	}

	// Update fields
	if make := r.FormValue("make"); make != "" {
		currentCar.Make = make
	}
	if model := r.FormValue("model"); model != "" {
		currentCar.Model = model
	}
	if plate := r.FormValue("license_plate"); plate != "" {
		currentCar.LicensePlate = plate
	}
	if newStatus != "" {
		currentCar.Status = newStatus
	}

	// Handle Image Upload
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		url, err := h.Storage.UploadFile(r.Context(), header.Filename, file, header.Size, contentType)
		if err != nil {
			http.Error(w, "Failed to upload image: "+err.Error(), http.StatusInternalServerError)
			return
		}
		currentCar.ImageURL = url
	}

	if err := h.Repo.Update(r.Context(), currentCar); err != nil {
		http.Error(w, "Failed to update car: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   currentCar,
	})
}
