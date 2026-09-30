package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
)

// Razorpay remains the external payment system; SkillSifter stores only subscription state and opaque references.
type razorpaySubscriptionPayload struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	PlanID       string            `json:"plan_id"`
	CurrentStart int64             `json:"current_start"`
	CurrentEnd   int64             `json:"current_end"`
	Notes        map[string]string `json:"notes"`
}

type razorpayWebhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Subscription struct {
			Entity razorpaySubscriptionPayload `json:"entity"`
		} `json:"subscription"`
	} `json:"payload"`
}

func razorpayRequest(method, path string, payload []byte) ([]byte, int, error) {
	key := db.GetEnv("RAZORPAY_KEY_ID", "")
	secret := db.GetEnv("RAZORPAY_KEY_SECRET", "")
	if key == "" || secret == "" {
		return nil, http.StatusServiceUnavailable, nil
	}

	req, err := http.NewRequest(method, "https://api.razorpay.com/v1"+path, bytes.NewReader(payload))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	req.SetBasicAuth(key, secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, http.StatusBadGateway, nil
	}
	return body, http.StatusOK, nil
}

func GetSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(
		"SELECT code,name,amount_minor,currency,billing_period,billing_interval,user_limit FROM platform_plans WHERE active=TRUE ORDER BY amount_minor,code",
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not load subscription plans")
		return
	}
	defer rows.Close()

	type plan struct {
		Code            string `json:"code"`
		Name            string `json:"name"`
		AmountMinor     int64  `json:"amountMinor"`
		Currency        string `json:"currency"`
		BillingPeriod   string `json:"billingPeriod"`
		BillingInterval int    `json:"billingInterval"`
		UserLimit       int    `json:"userLimit"`
	}

	plans := []plan{}
	for rows.Next() {
		var p plan
		if err := rows.Scan(&p.Code, &p.Name, &p.AmountMinor, &p.Currency, &p.BillingPeriod, &p.BillingInterval, &p.UserLimit); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not read subscription plans")
			return
		}
		plans = append(plans, p)
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Subscription plans retrieved successfully",
		Data:    plans,
	})
}

func GetSubscriptionAccount(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	ExpireDueSubscriptions()

	var id int64
	var planCode, status, provider, ref, planName string
	var starts time.Time
	var ends *time.Time
	var userLimit int

	err := db.DB.QueryRow(
		"SELECT s.id,s.plan_code,s.status,s.starts_at,s.ends_at,COALESCE(s.provider,''),COALESCE(s.provider_subscription_ref,''),COALESCE(p.name,''),COALESCE(p.user_limit,1) FROM platform_subscriptions s LEFT JOIN platform_plans p ON p.code=s.plan_code WHERE s.tenant_id=$1 ORDER BY s.starts_at DESC,s.id DESC LIMIT 1",
		tenantID,
	).Scan(&id, &planCode, &status, &starts, &ends, &provider, &ref, &planName, &userLimit)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No subscription found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Subscription retrieved successfully",
		Data: map[string]interface{}{
			"id": id, "planCode": planCode, "planName": planName, "status": status,
			"startsAt": starts, "endsAt": ends, "provider": provider,
			"providerSubscriptionRef": ref, "userLimit": userLimit,
		},
	})
}

func StartSubscriptionCheckout(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	var input struct {
		PlanCode string `json:"planCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.PlanCode) == "" {
		respondWithError(w, http.StatusBadRequest, "planCode is required")
		return
	}

	var phoneVerified bool
	if err := db.DB.QueryRow("SELECT phone_verified_at IS NOT NULL FROM users WHERE id = (SELECT MIN(id) FROM users WHERE tenant_id=$1 AND role='admin')", tenantID).Scan(&phoneVerified); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not verify subscription phone status")
		return
	}
	if !phoneVerified {
		respondWithError(w, http.StatusPreconditionRequired, "Verify the administrator phone number before subscribing")
		return
	}

	var name, provider, providerRef string
	var totalCount int
	if err := db.DB.QueryRow(
		"SELECT name,provider,COALESCE(provider_plan_ref,''),total_count FROM platform_plans WHERE code=$1 AND active=TRUE",
		input.PlanCode,
	).Scan(&name, &provider, &providerRef, &totalCount); err != nil {
		respondWithError(w, http.StatusNotFound, "Subscription plan not found")
		return
	}
	if provider != "razorpay" || providerRef == "" {
		respondWithError(w, http.StatusServiceUnavailable, "Subscription plan is not configured with the payment provider")
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"plan_id": providerRef, "total_count": totalCount, "quantity": 1,
		"customer_notify": 1,
		"notes":           map[string]string{"tenant_id": tenantID, "plan_code": input.PlanCode},
	})
	body, status, err := razorpayRequest(http.MethodPost, "/subscriptions", payload)
	if err != nil {
		respondWithError(w, status, "Payment provider unavailable")
		return
	}
	if status != http.StatusOK {
		respondWithError(w, status, "Payment provider rejected the checkout")
		return
	}

	var sub struct {
		ID       string `json:"id"`
		ShortURL string `json:"short_url"`
	}
	if err := json.Unmarshal(body, &sub); err != nil || sub.ID == "" {
		respondWithError(w, http.StatusBadGateway, "Invalid payment provider response")
		return
	}

	_, err = db.DB.Exec(
		"INSERT INTO platform_subscription_checkouts(tenant_id,plan_code,provider,provider_subscription_ref,checkout_url,status) VALUES($1,$2,'razorpay',$3,$4,'PENDING')",
		tenantID, input.PlanCode, sub.ID, sub.ShortURL,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save checkout state")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Subscription checkout created",
		Data: map[string]string{
			"planCode": input.PlanCode, "planName": name, "provider": "razorpay",
			"subscriptionId": sub.ID, "checkoutUrl": sub.ShortURL,
		},
	})
}

func CancelSubscription(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}

	var provider, ref string
	if err := db.DB.QueryRow(
		"SELECT provider,provider_subscription_ref FROM platform_subscriptions WHERE tenant_id=$1 AND status IN ('TRIAL','ACTIVE','PAST_DUE') ORDER BY id DESC LIMIT 1",
		tenantID,
	).Scan(&provider, &ref); err != nil {
		respondWithError(w, http.StatusNotFound, "Active subscription not found")
		return
	}
	if provider != "razorpay" || ref == "" {
		respondWithError(w, http.StatusBadRequest, "Subscription provider is not configured")
		return
	}

	body, status, err := razorpayRequest(http.MethodPost, "/subscriptions/"+ref+"/cancel", []byte("{\"cancel_at_cycle_end\":1}"))
	_ = body
	if err != nil {
		respondWithError(w, status, "Payment provider unavailable")
		return
	}
	if status != http.StatusOK {
		respondWithError(w, status, "Payment provider rejected the cancellation")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Subscription cancellation requested",
		Data:    map[string]string{"provider": provider, "subscriptionId": ref},
	})
}

func RazorpaySubscriptionWebhook(w http.ResponseWriter, r *http.Request) {
	secret := db.GetEnv("RAZORPAY_WEBHOOK_SECRET", "")
	if secret == "" {
		respondWithError(w, http.StatusServiceUnavailable, "Webhook secret is not configured")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid webhook body")
		return
	}
	defer r.Body.Close()

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	if !hmac.Equal([]byte(strings.ToLower(r.Header.Get("X-Razorpay-Signature"))), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		respondWithError(w, http.StatusUnauthorized, "Invalid webhook signature")
		return
	}

	var event razorpayWebhookPayload
	if err := json.Unmarshal(body, &event); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}

	sub := event.Payload.Subscription.Entity
	if sub.ID == "" {
		respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Webhook ignored"})
		return
	}

	eventRef := r.Header.Get("X-Razorpay-Event-Id")
	if eventRef == "" {
		sum := sha256.Sum256(body)
		eventRef = hex.EncodeToString(sum[:])
	}

	var tenantID, planCode string
	err = db.DB.QueryRow(
		"SELECT tenant_id,plan_code FROM platform_subscriptions WHERE provider='razorpay' AND provider_subscription_ref=$1 ORDER BY id DESC LIMIT 1",
		sub.ID,
	).Scan(&tenantID, &planCode)
	if err != nil {
		err = db.DB.QueryRow(
			"SELECT tenant_id,plan_code FROM platform_subscription_checkouts WHERE provider='razorpay' AND provider_subscription_ref=$1 ORDER BY id DESC LIMIT 1",
			sub.ID,
		).Scan(&tenantID, &planCode)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Subscription checkout not found")
			return
		}
	}

	var exists bool
	if err := db.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM platform_subscription_events WHERE provider='razorpay' AND provider_event_ref=$1)",
		eventRef,
	).Scan(&exists); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not inspect webhook event")
		return
	}
	if exists {
		respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Webhook already processed"})
		return
	}

	status, accountStatus := "ACTIVE", "ACTIVE"
	switch event.Event {
	case "subscription.pending":
		status, accountStatus = "PAST_DUE", "SUSPENDED"
	case "subscription.halted":
		status, accountStatus = "SUSPENDED", "SUSPENDED"
	case "subscription.cancelled":
		status, accountStatus = "CANCELLED", "SUSPENDED"
	case "subscription.expired":
		status, accountStatus = "EXPIRED", "EXPIRED"
	case "subscription.resumed", "subscription.activated", "subscription.charged":
	default:
		respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Webhook event ignored"})
		return
	}

	starts := time.Now()
	var ends *time.Time
	if sub.CurrentStart > 0 {
		t := time.Unix(sub.CurrentStart, 0)
		starts = t
	}
	if sub.CurrentEnd > 0 {
		t := time.Unix(sub.CurrentEnd, 0)
		ends = &t
	}

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not start webhook transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO platform_subscriptions(tenant_id,plan_code,status,starts_at,ends_at,provider,provider_subscription_ref) VALUES($1,$2,$3,$4,$5,'razorpay',$6) ON CONFLICT(provider,provider_subscription_ref) DO UPDATE SET status=EXCLUDED.status,starts_at=EXCLUDED.starts_at,ends_at=EXCLUDED.ends_at,updated_at=NOW()",
		tenantID, planCode, status, starts, ends, sub.ID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update subscription")
		return
	}

	_, _ = tx.Exec(
		"UPDATE platform_subscription_checkouts SET status=CASE WHEN $1='ACTIVE' THEN 'COMPLETED' ELSE status END,updated_at=NOW() WHERE provider='razorpay' AND provider_subscription_ref=$2",
		status, sub.ID,
	)
	_, err = tx.Exec("UPDATE platform_tenants SET account_status=$1,updated_at=NOW() WHERE tenant_id=$2", accountStatus, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update tenant access")
		return
	}
	_, err = tx.Exec(
		"INSERT INTO platform_subscription_events(provider,provider_event_ref,tenant_id,provider_subscription_ref,event_type) VALUES('razorpay',$1,$2,$3,$4)",
		eventRef, tenantID, sub.ID, event.Event,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not record webhook event")
		return
	}
	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not commit webhook state")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Webhook processed"})
}

func ExpireDueSubscriptions() {
	_, _ = db.DB.Exec("UPDATE platform_subscriptions SET status='EXPIRED',updated_at=NOW() WHERE status IN ('TRIAL','ACTIVE') AND ends_at IS NOT NULL AND ends_at<NOW()")
	_, _ = db.DB.Exec("UPDATE platform_tenants t SET account_status='EXPIRED',updated_at=NOW() WHERE EXISTS(SELECT 1 FROM platform_subscriptions s WHERE s.tenant_id=t.tenant_id AND s.status='EXPIRED') AND NOT EXISTS(SELECT 1 FROM platform_subscriptions s WHERE s.tenant_id=t.tenant_id AND s.status IN ('TRIAL','ACTIVE'))")
}
