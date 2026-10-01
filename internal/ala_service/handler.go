package ala_service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

type createAliasRequest struct {
	AliasURL    string `json:"alias_url" binding:"required,min=3,max=64"`
	RedirectURI string `json:"redirect_uri" binding:"required,url"`
}

type createAliasResponse struct {
	AliasID  uint   `json:"alias_id"`
	AliasURL string `json:"alias_url"`
	ShortURL string `json:"short_url"`
}

type getAliasesResponse struct {
	Count  uint    `json:"count"`
	Values []Alias `json:"values"`
}

// CreateAlias handles POST /v1/aliases.
func (h *Handler) CreateAlias(c *gin.Context) {
	var req createAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alias := Alias{
		AliasURL:    req.AliasURL,
		RedirectURI: req.RedirectURI,
	}

	if err := h.DB.Create(&alias).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "alias_url already exists"})
		return
	}

	c.JSON(http.StatusCreated, createAliasResponse{
		AliasID:  alias.AliasID,
		AliasURL: alias.AliasURL,
		ShortURL: buildShortURL(c, alias.AliasURL),
	})
}

func (h *Handler) GetAliases(c *gin.Context) {
	var aliases []Alias
	if err := h.DB.Find(&aliases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to get aliases"})
		return
	}

	c.JSON(http.StatusOK, getAliasesResponse{
		Count:  uint(len(aliases)),
		Values: aliases,
	})
}

// generateShortCode returns a 6-char base64url string for the random alias fallback.
func generateShortCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:6]
}

// buildShortURL constructs the public URL for a given alias.
func buildShortURL(c *gin.Context, aliasURL string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + "/" + aliasURL
}

// Redirect handles GET /:alias_url. Looks up the alias and 302s to redirect_uri.
func (h *Handler) Redirect(c *gin.Context) {
	aliasURL := c.Param("alias_url")

	var alias Alias
	// GORM's soft-delete filter: WHERE deleted_at 	IS NULL
	if err := h.DB.Where("alias_url = ?", aliasURL).First(&alias).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "alias not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Update last_used_datetime_utc (async, fire-and-forget)
	now := time.Now().UTC()

	h.DB.Model(&alias).Update("last_used_datetime_utc",
		&now)

	c.Redirect(http.StatusFound, alias.RedirectURI)
}
