package handler

import "github.com/gin-gonic/gin"

// RegisterAll mounts every resource group under /api/v1. The gateway
// entry point calls this once after wiring middleware; each resource
// file owns its own group so adding a service is one method call here.
func (h *Handler) RegisterAll(rg *gin.RouterGroup) {
	h.registerUsers(rg)
	h.registerStrategies(rg)
	h.registerFormulas(rg)
	h.registerBacktests(rg)
	h.registerMarkets(rg)
	h.registerRankings(rg)
	h.registerPortfolios(rg)
	h.registerCommunity(rg)
	h.registerAI(rg)
	h.registerBilling(rg)
	h.registerNotifications(rg)
}
