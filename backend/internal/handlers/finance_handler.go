package handlers

import (
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FinanceHandler struct {
	financeRepo repository.FinanceRepository
}

func NewFinanceHandler(financeRepo repository.FinanceRepository) *FinanceHandler {
	return &FinanceHandler{financeRepo: financeRepo}
}

func (h *FinanceHandler) GetStats(c *gin.Context) {
	stats, err := h.financeRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при расчете статистики"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *FinanceHandler) GetExpenses(c *gin.Context) {
	expenses, err := h.financeRepo.GetExpenses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении расходов"})
		return
	}
	if expenses == nil {
		expenses = []models.Expense{}
	}
	c.JSON(http.StatusOK, expenses)
}

func (h *FinanceHandler) CreateExpense(c *gin.Context) {
	var input struct {
		Amount      float64 `json:"amount" binding:"required"`
		Category    string  `json:"category" binding:"required"`
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	expense := &models.Expense{
		Amount:      input.Amount,
		Category:    input.Category,
		Description: input.Description,
	}

	if err := h.financeRepo.CreateExpense(expense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании расхода"})
		return
	}

	c.JSON(http.StatusCreated, expense)
}

func (h *FinanceHandler) GetWaiterSalaries(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать start_date и end_date"})
		return
	}

	salaries, err := h.financeRepo.GetWaiterSalaries(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении зарплат официантов"})
		return
	}
	
	c.JSON(http.StatusOK, salaries)
}
