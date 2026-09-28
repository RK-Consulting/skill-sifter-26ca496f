package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenantID").(string)
	companyName := r.Context().Value("companyName").(string)

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if input.Username == "" || input.Email == "" || input.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	switch input.Role {
	case "manager", "recruiter", "team_leader":
	default:
		respondWithError(w, http.StatusBadRequest, "Additional users must have role manager, recruiter, or team_leader")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not start transaction")
		return
	}
	defer tx.Rollback()

	var userLimit, userCount int
	err = tx.QueryRow(`
		SELECT s.user_limit, COUNT(u.id)
		FROM platform_subscriptions s
		LEFT JOIN users u ON u.tenant_id = s.tenant_id
		WHERE s.tenant_id = $1
		  AND s.status IN ('TRIAL', 'ACTIVE')
		  AND (s.ends_at IS NULL OR s.ends_at >= NOW())
		GROUP BY s.id, s.user_limit
		ORDER BY s.starts_at DESC
		LIMIT 1
	`, tenantID).Scan(&userLimit, &userCount)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusForbidden, "No active subscription for this tenant")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not read subscription user limit")
		return
	}
	if userCount >= userLimit {
		respondWithError(w, http.StatusConflict, "Subscription user limit reached")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	var userID int
	err = tx.QueryRow(`
		INSERT INTO users(username, email, password, role, tenant_id, company_name, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, input.Username, input.Email, hashedPassword, input.Role, tenantID, companyName, time.Now()).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	_, err = tx.Exec(`
		INSERT INTO platform_user_accounts(user_id, tenant_id, email, role)
		VALUES($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET tenant_id = EXCLUDED.tenant_id,
		    email = EXCLUDED.email,
		    role = EXCLUDED.role,
		    updated_at = NOW()
	`, userID, tenantID, input.Email, input.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create platform user account")
		return
	}

	if err = tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not commit transaction")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "User created successfully",
		Data: models.User{
			ID:          userID,
			Username:    input.Username,
			Email:       input.Email,
			Role:        input.Role,
			TenantID:    tenantID,
			CompanyName: companyName,
			CreatedAt:   time.Now(),
		},
	})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["id"]
	tenantID := r.Context().Value("tenantID").(string)

	var currentRole string
	err := db.DB.QueryRow(
		`SELECT role FROM users WHERE id = $1 AND tenant_id = $2`,
		targetID, tenantID,
	).Scan(&currentRole)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error looking up user")
		return
	}
	if currentRole == "admin" {
		respondWithError(w, http.StatusForbidden, "The tenant admin cannot be edited through this endpoint")
		return
	}

	var update struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if update.Role != "" {
		switch update.Role {
		case "manager", "recruiter", "team_leader":
		default:
			respondWithError(w, http.StatusBadRequest, "User role must be manager, recruiter, or team_leader")
			return
		}
	}

	_, err = db.DB.Exec(
		`UPDATE users SET
			username = COALESCE(NULLIF($1, ''), username),
			email = COALESCE(NULLIF($2, ''), email),
			role = COALESCE(NULLIF($3, ''), role)
		WHERE id = $4 AND tenant_id = $5`,
		update.Username, update.Email, update.Role, targetID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user")
		return
	}

	_, err = db.DB.Exec(
		`UPDATE platform_user_accounts
		SET email = COALESCE(NULLIF($1, ''), email),
		    role = COALESCE(NULLIF($2, ''), role),
		    updated_at = NOW()
		WHERE user_id = $3 AND tenant_id = $4`,
		update.Email, update.Role, targetID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating platform account")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["id"]
	tenantID := r.Context().Value("tenantID").(string)

	var targetRole string
	err := db.DB.QueryRow(
		`SELECT role FROM users WHERE id = $1 AND tenant_id = $2`,
		targetID, tenantID,
	).Scan(&targetRole)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error looking up user")
		return
	}
	if targetRole == "admin" {
		respondWithError(w, http.StatusForbidden, "The tenant admin cannot be deleted")
		return
	}

	_, err = db.DB.Exec(
		`DELETE FROM users WHERE id = $1 AND tenant_id = $2`,
		targetID, tenantID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting user")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
