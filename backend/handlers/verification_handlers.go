package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"golang.org/x/crypto/bcrypt"
)

func otpHash(code string) string {
	secret := db.GetEnv("OTP_HASH_SECRET", "change-me-in-production")
	sum := sha256.Sum256([]byte(secret + ":" + code))
	return hex.EncodeToString(sum[:])
}

func generateOTP() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

type smtpLoginAuth struct {
	username string
	password string
	step     int
}

func (a *smtpLoginAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *smtpLoginAuth) Next(_ []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	a.step++
	if a.step == 1 {
		return []byte(a.username), nil
	}
	return []byte(a.password), nil
}

func sendEmailOTP(to, code string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USERNAME")
	if user == "" {
		user = os.Getenv("SMTP_USER")
	}
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	if host == "" || port == "" || from == "" {
		return fmt.Errorf("SMTP is not configured")
	}

	msg := []byte("From: " + from + "\r\nTo: " + to + "\r\nSubject: Your SkillSifter verification code\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nYour SkillSifter verification code is " + code + ". It expires in 10 minutes.\r\n")

	// Port 465 is implicit TLS. Exim/cPanel installations commonly authenticate
	// either with the mailbox name or the complete mailbox address. Try the
	// configured identity first, then the sender address as a safe fallback.
	users := []string{}
	if user != "" {
		users = append(users, user)
	}
	if from != "" && !strings.EqualFold(from, user) {
		users = append(users, from)
	}

	var lastErr error
	for _, authUser := range users {
		if port == "465" {
			tlsConfig := &tls.Config{
				ServerName: host,
				MinVersion: tls.VersionTLS12,
			}
			conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), tlsConfig)
			if err != nil {
				lastErr = err
				continue
			}
			client, err := smtp.NewClient(conn, host)
			if err != nil {
				_ = conn.Close()
				lastErr = err
				continue
			}

			authOK := false
			if ok, mechanisms := client.Extension("AUTH"); ok {
				mechanisms = strings.ToUpper(mechanisms)
				if strings.Contains(mechanisms, "PLAIN") {
					err = client.Auth(smtp.PlainAuth("", authUser, password, host))
					if err == nil {
						authOK = true
					} else {
						lastErr = err
					}
				}
				if !authOK && strings.Contains(mechanisms, "LOGIN") {
					auth := &smtpLoginAuth{username: authUser, password: password}
					err = client.Auth(auth)
					if err == nil {
						authOK = true
					} else {
						lastErr = err
					}
				}
			} else {
				lastErr = fmt.Errorf("SMTP server does not advertise AUTH")
			}

			if !authOK {
				_ = client.Quit()
				continue
			}
			if err = client.Mail(from); err != nil {
				lastErr = err
				_ = client.Quit()
				continue
			}
			if err = client.Rcpt(to); err != nil {
				lastErr = err
				_ = client.Quit()
				continue
			}
			writer, err := client.Data()
			if err != nil {
				lastErr = err
				_ = client.Quit()
				continue
			}
			if _, err = writer.Write(msg); err != nil {
				lastErr = err
				_ = writer.Close()
				_ = client.Quit()
				continue
			}
			if err = writer.Close(); err != nil {
				lastErr = err
				_ = client.Quit()
				continue
			}
			if err = client.Quit(); err != nil {
				lastErr = err
				continue
			}
			return nil
		}

		// Standard SMTP ports use the Go SMTP client's STARTTLS handling.
		var auth smtp.Auth
		if authUser != "" {
			auth = smtp.PlainAuth("", authUser, password, host)
		}
		if err := smtp.SendMail(net.JoinHostPort(host, port), auth, from, []string{to}, msg); err != nil {
			lastErr = err
			continue
		}
		return nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("SMTP authentication failed")
	}
	return lastErr
}
func sendSMSOTP(phone, code string) error {
	endpoint := os.Getenv("SMS_OTP_URL")
	if endpoint == "" {
		return fmt.Errorf("SMS provider is not configured")
	}
	payload := fmt.Sprintf("{\"phone\":%q,\"code\":%q}", phone, code)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("SMS_OTP_API_KEY"); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("SMS provider returned status %d", resp.StatusCode)
	}
	return nil
}

func StartRegistration(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username    string
		Email       string
		Password    string
		CompanyName string
		PlanCode    string
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.CompanyName = strings.TrimSpace(input.CompanyName)
	if input.Username == "" || input.Email == "" || input.Password == "" || input.CompanyName == "" || input.PlanCode == "" {
		respondWithError(w, http.StatusBadRequest, "Name, email, password, company and plan are required")
		return
	}

	var ok bool
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM platform_plans WHERE code=$1 AND active=TRUE)", input.PlanCode).Scan(&ok); err != nil || !ok {
		respondWithError(w, http.StatusBadRequest, "Selected plan is not available")
		return
	}
	if err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM platform_registration_registry WHERE email_id=$1) OR EXISTS(SELECT 1 FROM platform_pending_registrations WHERE email=$1 AND email_verified_at IS NULL AND expires_at>NOW())", input.Email).Scan(&ok); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not validate email")
		return
	}
	if ok {
		respondWithError(w, http.StatusConflict, "This email is already registered or has a pending verification")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not secure password")
		return
	}

	var registrationID int64
	err = db.DB.QueryRow("INSERT INTO platform_pending_registrations(username,email,company_name,password_hash,plan_code) VALUES($1,$2,$3,$4,$5) RETURNING id", input.Username, input.Email, input.CompanyName, string(passwordHash), input.PlanCode).Scan(&registrationID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create registration")
		return
	}

	code := generateOTP()
	_, _ = db.DB.Exec("DELETE FROM platform_verification_codes WHERE purpose='EMAIL_SIGNUP' AND registration_id=$1 AND consumed_at IS NULL", registrationID)
	_, err = db.DB.Exec("INSERT INTO platform_verification_codes(purpose,registration_id,destination,code_hash,expires_at) VALUES('EMAIL_SIGNUP',$1,$2,$3,NOW()+INTERVAL '10 minutes')", registrationID, input.Email, otpHash(code))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create verification code")
		return
	}

	if err := sendEmailOTP(input.Email, code); err != nil {
		_, _ = db.DB.Exec("DELETE FROM platform_pending_registrations WHERE id=$1", registrationID)
		log.Printf("registration email OTP failed: %v", err)
		respondWithError(w, http.StatusServiceUnavailable, "Email verification is temporarily unavailable")
		return
	}

	respondWithJSON(w, http.StatusAccepted, models.ApiResponse{Success: true, Message: "Verification code sent to your email", Data: map[string]interface{}{"registrationId": registrationID, "email": input.Email}})
}

func VerifyRegistrationEmail(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RegistrationID int64
		Code           string
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.RegistrationID == 0 || strings.TrimSpace(input.Code) == "" {
		respondWithError(w, http.StatusBadRequest, "registrationId and verification code are required")
		return
	}

	var username, email, company, passwordHash, planCode string
	var registrationExpires time.Time
	err := db.DB.QueryRow("SELECT username,email,company_name,password_hash,plan_code,expires_at FROM platform_pending_registrations WHERE id=$1 AND email_verified_at IS NULL", input.RegistrationID).Scan(&username, &email, &company, &passwordHash, &planCode, &registrationExpires)
	if err != nil || time.Now().After(registrationExpires) {
		respondWithError(w, http.StatusBadRequest, "Registration is invalid or expired")
		return
	}

	var verificationID int64
	var codeHash string
	var attempts int
	err = db.DB.QueryRow("SELECT id,code_hash,attempts FROM platform_verification_codes WHERE purpose='EMAIL_SIGNUP' AND registration_id=$1 AND consumed_at IS NULL AND expires_at>NOW() ORDER BY id DESC LIMIT 1", input.RegistrationID).Scan(&verificationID, &codeHash, &attempts)
	if err != nil || attempts >= 5 {
		respondWithError(w, http.StatusBadRequest, "Verification code is invalid or expired")
		return
	}
	if otpHash(strings.TrimSpace(input.Code)) != codeHash {
		_, _ = db.DB.Exec("UPDATE platform_verification_codes SET attempts=attempts+1 WHERE id=$1", verificationID)
		respondWithError(w, http.StatusBadRequest, "Incorrect verification code")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not start registration transaction")
		return
	}
	defer tx.Rollback()

	var companyExists bool
	if err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM platform_tenants WHERE company_name=$1)", company).Scan(&companyExists); err != nil || companyExists {
		respondWithError(w, http.StatusConflict, "Company already has a SkillSifter account")
		return
	}
	tenantBytes := make([]byte, 16)
	if _, err = rand.Read(tenantBytes); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create tenant identity")
		return
	}
	tenantID := "tenant_" + hex.EncodeToString(tenantBytes)

	if _, err = tx.Exec("INSERT INTO platform_tenants(tenant_id,company_name,account_status,provisioning_status,trial_started_at,trial_expires_at,data_deletion_at) VALUES($1,$2,'ACTIVE','PENDING',NOW(),NOW()+INTERVAL '2 days',NOW()+INTERVAL '9 days')", tenantID, company); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create tenant")
		return
	}

	var userLimit int
	if err = tx.QueryRow("SELECT user_limit FROM platform_plans WHERE code=$1 AND active=TRUE", planCode).Scan(&userLimit); err != nil {
		respondWithError(w, http.StatusBadRequest, "Selected plan is not available")
		return
	}
	if _, err = tx.Exec("INSERT INTO platform_subscriptions(tenant_id,plan_code,status,starts_at,ends_at,user_limit) VALUES($1,$2,'TRIAL',NOW(),NOW()+INTERVAL '2 days',$3)", tenantID, planCode, userLimit); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create trial")
		return
	}
	if _, err = tx.Exec("UPDATE platform_verification_codes SET consumed_at=NOW() WHERE id=$1", verificationID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not complete verification")
		return
	}
	if _, err = tx.Exec("UPDATE platform_pending_registrations SET email_verified_at=NOW() WHERE id=$1", input.RegistrationID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not complete registration")
		return
	}
	if _, err = tx.Exec("INSERT INTO platform_registration_registry(email_id,first_registered,last_tenant_id) VALUES($1,NOW(),$2)", email, tenantID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not record registration")
		return
	}

	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not commit registration")
		return
	}
	_, tenantUserID, err := db.ProvisionTenantDatabase(db.DB, tenantID, company, &db.TenantUser{Username: username, Email: email, Password: passwordHash, Role: "admin"})
	if err != nil {
		_, _ = db.DB.Exec("UPDATE platform_tenants SET provisioning_status='FAILED' WHERE tenant_id=$1", tenantID)
		respondWithError(w, http.StatusServiceUnavailable, "Account created but tenant provisioning failed; please retry provisioning")
		return
	}

	if _, err = db.DB.Exec(`INSERT INTO platform_user_accounts(tenant_id,user_id,email,role) VALUES($1,$2,$3,$4) ON CONFLICT (tenant_id,user_id) DO UPDATE SET email=EXCLUDED.email, role=EXCLUDED.role, updated_at=NOW()`, tenantID, tenantUserID, email, "admin"); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform account")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{Success: true, Message: "Email verified and 2-day trial started", Data: map[string]interface{}{"tenantId": tenantID, "email": email, "subscriptionStatus": "TRIAL", "planCode": planCode}})
}

func SendPhoneVerificationCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	var input struct{ Phone string }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.Phone) == "" {
		respondWithError(w, http.StatusBadRequest, "Phone number is required")
		return
	}
	phone := strings.TrimSpace(input.Phone)
	code := generateOTP()

	var accountID int64
	if err := db.DB.QueryRow("SELECT id FROM platform_user_accounts WHERE tenant_id=$1 AND user_id=$2", tenantID, userID).Scan(&accountID); err != nil {
		respondWithError(w, http.StatusForbidden, "Platform account not found")
		return
	}
	_, _ = db.DB.Exec("DELETE FROM platform_verification_codes WHERE purpose='PHONE_SUBSCRIPTION' AND platform_account_id=$1 AND consumed_at IS NULL", accountID)

	var verificationID int64
	if err := db.DB.QueryRow("INSERT INTO platform_verification_codes(purpose,platform_account_id,destination,code_hash,expires_at) VALUES('PHONE_SUBSCRIPTION',$1,$2,$3,NOW()+INTERVAL '10 minutes') RETURNING id", accountID, phone, otpHash(code)).Scan(&verificationID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create phone verification")
		return
	}
	if err := sendSMSOTP(phone, code); err != nil {
		_, _ = db.DB.Exec("DELETE FROM platform_verification_codes WHERE id=$1", verificationID)
		respondWithError(w, http.StatusServiceUnavailable, "SMS verification is temporarily unavailable")
		return
	}
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Phone verification code sent"})
}

func VerifyPhoneVerificationCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		respondWithError(w, http.StatusUnauthorized, "Tenant context missing")
		return
	}
	var input struct{ Code string }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.Code) == "" {
		respondWithError(w, http.StatusBadRequest, "Verification code is required")
		return
	}
	var accountID int64
	if err := db.DB.QueryRow("SELECT id FROM platform_user_accounts WHERE tenant_id=$1 AND user_id=$2", tenantID, userID).Scan(&accountID); err != nil {
		respondWithError(w, http.StatusForbidden, "Platform account not found")
		return
	}
	var id int64
	var hash string
	var attempts int
	err := db.DB.QueryRow("SELECT id,code_hash,attempts FROM platform_verification_codes WHERE purpose='PHONE_SUBSCRIPTION' AND platform_account_id=$1 AND consumed_at IS NULL AND expires_at>NOW() ORDER BY id DESC LIMIT 1", accountID).Scan(&id, &hash, &attempts)
	if err != nil || attempts >= 5 {
		respondWithError(w, http.StatusBadRequest, "Verification code is invalid or expired")
		return
	}
	if otpHash(strings.TrimSpace(input.Code)) != hash {
		_, _ = db.DB.Exec("UPDATE platform_verification_codes SET attempts=attempts+1 WHERE id=$1", id)
		respondWithError(w, http.StatusBadRequest, "Incorrect verification code")
		return
	}
	if _, err = db.DB.Exec("UPDATE platform_user_accounts SET phone_verified_at=NOW(), updated_at=NOW() WHERE id=$1", accountID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not verify phone")
		return
	}
	_, _ = db.DB.Exec("UPDATE platform_verification_codes SET consumed_at=NOW() WHERE id=$1", id)
	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "Phone number verified"})
}
