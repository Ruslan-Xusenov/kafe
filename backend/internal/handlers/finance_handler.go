package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
)

type FinanceHandler struct {
	financeRepo repository.FinanceRepository
	wsService   *service.WebsocketService
}

func NewFinanceHandler(financeRepo repository.FinanceRepository, wsService *service.WebsocketService) *FinanceHandler {
	return &FinanceHandler{financeRepo: financeRepo, wsService: wsService}
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

func (h *FinanceHandler) CloseShift(c *gin.Context) {
	// 1. Fetch current shift stats
	stats, err := h.financeRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения статистики"})
		return
	}

	// 2. Broadcast to printer
	shiftReportPayload := map[string]interface{}{
		"type":           "shift_report",
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
		"total_revenue":  stats.TotalRevenue,
		"total_expenses": stats.TotalExpenses,
		"net_profit":     stats.NetProfit,
		"cash":           stats.CashRevenue,
		"card":           stats.CardRevenue,
		"click":          stats.ClickRevenue,
		"nasiya":         stats.NasiyaRevenue,
	}
	h.wsService.BroadcastToRole("printer", shiftReportPayload)

	// 3. Send Telegram Message
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := "660122397"
	
	if botToken != "" {
		msgText := fmt.Sprintf(
			"📅 <b>ОТЧЕТ ЗА СМЕНУ ЗАКРЫТ</b>\n\n"+
				"💰 Общая выручка: <b>%.0f</b> сум\n"+
				"💸 Общие расходы: <b>%.0f</b> сум\n"+
				"📈 Чистая прибыль: <b>%.0f</b> сум\n\n"+
				"💳 <b>Способы оплаты:</b>\n"+
				"💵 Наличные: %.0f сум\n"+
				"💳 Терминал: %.0f сум\n"+
				"📲 Click/Payme: %.0f сум\n"+
				"📓 В долг: %.0f сум\n\n"+
				"🕒 Время закрытия: %s",
			stats.TotalRevenue, stats.TotalExpenses, stats.NetProfit,
			stats.CashRevenue, stats.CardRevenue, stats.ClickRevenue, stats.NasiyaRevenue,
			time.Now().Format("2006-01-02 15:04"),
		)

		payload := map[string]interface{}{
			"chat_id":    chatID,
			"text":       msgText,
			"parse_mode": "HTML",
		}
		
		jsonData, _ := json.Marshal(payload)
		http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken), "application/json", bytes.NewBuffer(jsonData))
	}

	// 4. Update last_shift_closed_at in DB
	err = h.financeRepo.CloseShift()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при закрытии смены в БД"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Смена успешно закрыта"})
}
