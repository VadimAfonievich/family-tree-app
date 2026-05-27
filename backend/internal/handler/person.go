package handler

import (
	"net/http"
	"time"

	"family-tree-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PersonHandler struct {
	personService *service.PersonService
	treeService   *service.TreeService
}

func NewPersonHandler(personService *service.PersonService, treeService *service.TreeService) *PersonHandler {
	return &PersonHandler{
		personService: personService,
		treeService:   treeService,
	}
}

type createPersonRequest struct {
	TreeID    string  `json:"tree_id" binding:"required"`
	FirstName string  `json:"first_name" binding:"required"`
	LastName  string  `json:"last_name"`
	Gender    string  `json:"gender" binding:"required,oneof=male female other"`
	BirthDate *string `json:"birth_date"`
	DeathDate *string `json:"death_date"`
	PhotoURL  string  `json:"photo_url"`
}

func (h *PersonHandler) Create(c *gin.Context) {
	var req createPersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	treeID, err := uuid.Parse(req.TreeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tree_id"})
		return
	}

	// Verify user owns the tree
	userID := c.MustGet("user_id").(uuid.UUID)
	_, err = h.treeService.GetTree(c.Request.Context(), treeID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tree not found"})
		return
	}

	params := service.CreatePersonParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Gender:    req.Gender,
		PhotoURL:  req.PhotoURL,
	}

	if req.BirthDate != nil {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date format, use YYYY-MM-DD"})
			return
		}
		params.BirthDate = t
	}

	if req.DeathDate != nil {
		t, err := time.Parse("2006-01-02", *req.DeathDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid death_date format, use YYYY-MM-DD"})
			return
		}
		params.DeathDate = t
	}

	person, err := h.personService.CreatePerson(c.Request.Context(), treeID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create person"})
		return
	}

	c.JSON(http.StatusCreated, person)
}

func (h *PersonHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person id"})
		return
	}

	person, err := h.personService.GetPerson(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if _, err = h.treeService.GetTree(c.Request.Context(), person.TreeID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tree not found"})
		return
	}

	var req struct {
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Gender    string  `json:"gender" binding:"omitempty,oneof=male female other"`
		BirthDate *string `json:"birth_date"`
		DeathDate *string `json:"death_date"`
		PhotoURL  string  `json:"photo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := service.UpdatePersonParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Gender:    req.Gender,
		PhotoURL:  req.PhotoURL,
	}

	if req.BirthDate != nil {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date format"})
			return
		}
		params.BirthDate = &t
	}

	if req.DeathDate != nil {
		t, err := time.Parse("2006-01-02", *req.DeathDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid death_date format"})
			return
		}
		params.DeathDate = &t
	}

	updatedPerson, err := h.personService.UpdatePerson(c.Request.Context(), id, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update person"})
		return
	}

	c.JSON(http.StatusOK, updatedPerson)
}

func (h *PersonHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person id"})
		return
	}

	person, err := h.personService.GetPerson(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if _, err = h.treeService.GetTree(c.Request.Context(), person.TreeID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tree not found"})
		return
	}

	if err := h.personService.DeletePerson(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete person"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
