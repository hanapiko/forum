package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"forum/models"
	"forum/repository"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate category name
	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	// Validate category name length
	if len(category.Name) > 50 {
		http.Error(w, "Category name must be 50 characters or less", http.StatusBadRequest)
		return
	}

	err := h.categoryRepo.Create(&category)
	if err != nil {
		log.Printf("Failed to create category: %v", err)
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(category); err != nil {
		log.Printf("Failed to encode category response: %v", err)
	}
}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "id")
	categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := h.categoryRepo.GetByID(categoryID)
	if err != nil {
		log.Printf("Failed to retrieve category: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(category); err != nil {
		log.Printf("Failed to encode category response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.ListCategories()
	if err != nil {
		log.Printf("Failed to retrieve categories: %v", err)
		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		log.Printf("Failed to encode categories response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	// Parse category ID
	categoryIDStr := chi.URLParam(r, "id")
	categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Decode request body
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate category name
	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	// Validate category name length
	if len(category.Name) > 50 {
		http.Error(w, "Category name must be 50 characters or less", http.StatusBadRequest)
		return
	}

	// Set the ID from the URL parameter
	category.ID = categoryID

	// Update category
	err = h.categoryRepo.Update(&category)
	if err != nil {
		log.Printf("Failed to update category: %v", err)
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(category); err != nil {
		log.Printf("Failed to encode updated category response: %v", err)
	}
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	// Parse category ID
	categoryIDStr := chi.URLParam(r, "id")
	categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// Check if category exists before deletion
	_, err = h.categoryRepo.GetByID(categoryID)
	if err != nil {
		log.Printf("Category not found: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// Delete category
	err = h.categoryRepo.Delete(categoryID)
	if err != nil {
		log.Printf("Failed to delete category: %v", err)
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
