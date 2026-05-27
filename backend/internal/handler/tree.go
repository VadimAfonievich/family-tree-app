package handler

import (
	"net/http"

	"family-tree-backend/internal/db"
	"family-tree-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TreeHandler struct {
	treeService    *service.TreeService
	personService  *service.PersonService
	relationService *service.RelationService
}

func NewTreeHandler(treeService *service.TreeService, personService *service.PersonService, relationService *service.RelationService) *TreeHandler {
	return &TreeHandler{
		treeService:    treeService,
		personService:  personService,
		relationService: relationService,
	}
}

type createTreeRequest struct {
	Title string `json:"title" binding:"required"`
}

func (h *TreeHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	trees, err := h.treeService.ListTrees(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list trees"})
		return
	}

	if trees == nil {
		trees = []db.Tree{}
	}

	c.JSON(http.StatusOK, trees)
}

func (h *TreeHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tree, err := h.treeService.CreateTree(c.Request.Context(), userID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tree"})
		return
	}

	c.JSON(http.StatusCreated, tree)
}

func (h *TreeHandler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	treeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tree id"})
		return
	}

	tree, err := h.treeService.GetTree(c.Request.Context(), treeID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tree not found"})
		return
	}

	// Get persons and relations for this tree
	persons, _ := h.personService.ListPersons(c.Request.Context(), treeID)
	relations, _ := h.relationService.ListRelations(c.Request.Context(), treeID)

	if persons == nil {
		persons = []db.Person{}
	}
	if relations == nil {
		relations = []db.Relation{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tree":      tree,
		"persons":   persons,
		"relations": relations,
	})
}

func (h *TreeHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	treeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tree id"})
		return
	}

	if err := h.treeService.DeleteTree(c.Request.Context(), treeID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tree not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
