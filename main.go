package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Data
var subscriptions = []gin.H{
	{"module": "Cloud Security", "status": "Active", "next_billing": "2025-09-30", "usage": 75, "plan": "Pro"},
	{"module": "Endpoint Security", "status": "Payment Failed", "next_billing": "2025-09-15", "usage": 20, "plan": "Basic"},
	{"module": "Network Security", "status": "Active", "next_billing": "2025-09-25", "usage": 50, "plan": "Pro"},
	{"module": "Compliance", "status": "Grace Period", "next_billing": "2025-09-18", "usage": 10, "plan": "Starter"},
	{"module": "VAPT", "status": "Expired", "next_billing": "2025-08-10", "usage": 0, "plan": "Starter"},
}

var billingInfo = gin.H{
	"billing_email":  "billing@ostoclient.in",
	"payment_method": "HDFC Visa **** 4242",
	"backup_payment": "ICICI Mastercard **** 1111",
}

var invoices = []gin.H{
	{"id": "inv_001", "amount": 15000.00, "currency": "INR", "status": "Paid", "due_date": "2025-09-01"},
	{"id": "inv_002", "amount": 22000.00, "currency": "INR", "status": "Pending", "due_date": "2025-09-15"},
	{"id": "inv_003", "amount": 9500.00, "currency": "INR", "status": "Failed", "due_date": "2025-09-10"},
}

var payments = []gin.H{
	{"id": "pay_001", "invoice_id": "inv_001", "amount": 15000.00, "currency": "INR", "status": "Success", "created_at": "2025-09-01T12:30:00+05:30"},
	{"id": "pay_003", "invoice_id": "inv_003", "amount": 9500.00, "currency": "INR", "status": "Failed", "created_at": "2025-09-10T10:00:00+05:30"},
}

func getSubscriptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

func getBillingInfo(c *gin.Context) {
	c.JSON(http.StatusOK, billingInfo)
}

func updateBillingEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	billingInfo["billing_email"] = req.Email
	c.JSON(http.StatusOK, gin.H{"message": "Billing email updated", "billing_email": req.Email})
}

func getInvoices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

func payInvoice(c *gin.Context) {
	invoiceId := c.Param("invoiceId")
	c.JSON(http.StatusOK, gin.H{
		"message":    "Payment initiated",
		"invoice_id": invoiceId,
		"redirect":   "https://checkout.stripe.com/test-session", // mock Stripe Checkout
	})
}

func getPaymentHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

func main() {
	r := gin.Default()

	// Subscriptions
	r.GET("/subscriptions/:userId", getSubscriptions)

	// Billing
	r.GET("/billing/:userId", getBillingInfo)
	r.POST("/billing/:userId/email", updateBillingEmail)

	// Invoices & Payments
	r.GET("/invoices/:userId", getInvoices)
	r.POST("/invoices/:invoiceId/pay", payInvoice)
	r.GET("/invoices/:userId/history", getPaymentHistory)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)

}
