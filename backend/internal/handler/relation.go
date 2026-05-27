package handler

import (
	"net/http"

	"family-tree-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RelationHandler struct {
	relationService *service.RelationService
	treeService     *service.TreeService
}

func NewRelationHandler(relationService *service.RelationService, treeService *service.TreeService) *RelationHandler {
	return &RelationHandler{
		relationService: relationService,
		treeService:     treeService,
	}
}

type createRelationRequest struct {
	TreeID       string `json:"tree_id" binding:"required"`
	Person1ID    string `json:"person1_id" binding:"required"`
	Person2ID    string `json:"person2_id" binding:"required"`
	RelationType string `json:"relation_type" binding:"required,oneof=parent spouse"`
}

func (h *RelationHandler) Create(c *gin.Context) {
	var req createRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	treeID, err := uuid.Parse(req.TreeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tree_id"})
		return
	}

	person1ID, err := uuid.Parse(req.Person1ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person1_id"})
		return
	}

	person2ID, err := uuid.Parse(req.Person2ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person2_id"})
		return
	}

	// Verify user owns the tree
	userID := c.MustGet("user_id").(uuid.UUID)
	_, err = h.treeService.GetTree(c.Request.Context(), treeID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tree not found"})
		return
	}

	relation, err := h.relationService.CreateRelation(c.Request.Context(), treeID, service.CreateRelationParams{
		Person1ID:    person1ID,
		Person2ID:    person2ID,
		RelationType: req.RelationType,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, relation)
}

func (h *RelationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid relation id"})
		return
	}

	relation, err := h.relationService.GetRelation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "relation not found"})
		return
	}
	userID := c.MustGet("user_id").(uuid.UUID)
	if _, err = h.treeService.GetTree(c.Request.Context(), relation.TreeID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "tree not found"})
		return
	}

	if err := h.relationService.DeleteRelation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete relation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
