package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

	_ = godotenv.Load()
	r := gin.Default()
	r.Use(cors.Default())

	// Subscriptions enpoints
	r.GET("/subscriptions/:userId", getSubscriptions)

	// Billing endpoints
	r.GET("/billing/:userId", getBillingInfo)
	r.POST("/billing/:userId/email", updateBillingEmail)

	// Invoices endpoints
	r.GET("/invoices/:userId", getInvoices)
	r.POST("/invoices/:invoiceId/pay", payInvoice)
	r.GET("/invoices/:userId/history", getPaymentHistory)

	//Payment endpoint
	r.POST("/api/payments/initiate", func(c *gin.Context) {
		keyID := os.Getenv("RAZORPAY_KEY_ID")
		keySecret := os.Getenv("RAZORPAY_SECRET")
		callbackURL := os.Getenv("CALLBACK_URL")

		if keyID == "" || keySecret == "" || callbackURL == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing Razorpay credentials or callback"})
			return
		}
		var req struct {
			InvoiceID string  `json:"invoice_id"`
			Amount    float64 `json:"amount"`
			Method    string  `json:"method"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		payload := map[string]interface{}{
			"amount":         req.Amount * 100, // INR → paise
			"currency":       "INR",
			"accept_partial": false,
			"description":    "Invoice " + req.InvoiceID,
			"customer": map[string]string{
				"name":    "Tushar Khatri",
				"email":   "hello@tusharkhatri.in",
				"contact": "9968290156",
			},
			"notify": map[string]bool{
				"sms":   true,
				"email": true,
			},
			"callback_url":    callbackURL,
			"callback_method": "get",
		}

		body, _ := json.Marshal(payload)

		reqRazor, _ := http.NewRequest("POST", "https://api.razorpay.com/v1/payment_links", bytes.NewBuffer(body))
		reqRazor.Header.Set("Content-Type", "application/json")
		reqRazor.SetBasicAuth(keyID, keySecret)

		client := &http.Client{}
		resp, err := client.Do(reqRazor)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment link"})
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Razorpay response"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"payment_link": result["short_url"],
			"invoice_id":   req.InvoiceID,
			"amount":       req.Amount,
			"status":       "link_created",
		})
	})

	r.GET("/api/payments/callback", func(c *gin.Context) {
		paymentID := c.Query("razorpay_payment_id")
		// orderID := c.Query("razorpay_order_id")
		// signature := c.Query("razorpay_signature")
		// status := "failed"

		var heading, color string
		if paymentID != "" {
			heading = "Payment Successful! Redirecting to home page..."
			color = "green"
		} else {
			heading = "Payment Failed. Redirecting to home page..."
			color = "red"
		}

		htmlResponse := fmt.Sprintf(`
        <h1 style="color: %s;">%s</h1>
        <script>
            setTimeout(function() {
                window.location.href = 'https://osto-billing-frontend.vercel.app/'; 
            }, 3000); // Redirect after 3 seconds
        </script>
    `, color, heading)

		c.Data(http.StatusOK, "text/html", []byte(htmlResponse))

	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)

}
