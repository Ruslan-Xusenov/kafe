package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/repository"
)

func intPtr(v int) *int {
	return &v
}

func auditActor(c *gin.Context) (*int, string) {
	var actorID *int
	if rawID, ok := c.Get("user_id"); ok {
		if id, ok := rawID.(int); ok {
			actorID = &id
		}
	}

	actorRole := ""
	if rawRole, ok := c.Get("role"); ok {
		if role, ok := rawRole.(string); ok {
			actorRole = role
		}
	}

	return actorID, actorRole
}

func writeAudit(c *gin.Context, repo *repository.AuditRepository, action, entityType string, entityID *int, details interface{}) {
	if repo == nil {
		return
	}

	actorID, actorRole := auditActor(c)
	if err := repo.Create(actorID, actorRole, action, entityType, entityID, details, c.ClientIP()); err != nil {
		log.Printf("AUDIT_LOG_ERROR: %v", err)
	}
}
