package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func tenantUserDB(r *http.Request) (string, *sql.DB, error) {
	tenantID, ok := r.Context().Value("tenantID").(string)
	if !ok || tenantID == "" {
		return "", nil, sql.ErrNoRows
	}
	tenantDB := db.RequestDB(r)
	if tenantDB == nil {
		return "", nil, fmt.Errorf("tenant database is not ready")
	}
	return tenantID, tenantDB, nil
}

func validOperationalRole(role string) bool {
	switch role {
	case "manager", "recruiter", "team_leader":
		return true
	default:
		return false
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, tenantDB, err := tenantUserDB(r)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database is not ready")
		return
	}
	companyName, _ := r.Context().Value("companyName").(string)

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

	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Username == "" || input.Email == "" || input.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}
	if !validOperationalRole(input.Role) {
		respondWithError(w, http.StatusBadRequest, "Additional users must have role manager, recruiter, or team_leader")
		return
	}

	// User-limit enforcement is a cross-database operation. Serialize it per
	// tenant with a PostgreSQL advisory transaction lock so two simultaneous
	// requests cannot both observe the same free slot.
	controlTx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not start user creation")
		return
	}
	defer controlTx.Rollback()

	if _, err = controlTx.Exec("SELECT pg_advisory_xact_lock(hashtext($1)::bigint)", "skill-sifter:user-limit:"+tenantID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not lock tenant user limit")
		return
	}

	var userLimit int
	if err = controlTx.QueryRow(`
		SELECT user_limit
		FROM platform_subscriptions
		WHERE tenant_id=$1
		  AND status IN ('TRIAL','ACTIVE')
		  AND (ends_at IS NULL OR ends_at >= NOW())
		ORDER BY starts_at DESC, id DESC
		LIMIT 1
	`, tenantID).Scan(&userLimit); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusForbidden, "No active subscription for this tenant")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not read subscription user limit")
		return
	}

	var userCount int
	if err = tenantDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not count tenant users")
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
	if err = tenantDB.QueryRow(`
		INSERT INTO users(username,email,password,role,tenant_id)
		VALUES($1,$2,$3,$4,$5)
		RETURNING id
	`, input.Username, input.Email, string(hashedPassword), input.Role, tenantID).Scan(&userID); err != nil {
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not create tenant user")
		return
	}

	if _, err = controlTx.Exec(`
		INSERT INTO platform_user_accounts(tenant_id,user_id,email,role)
		VALUES($1,$2,$3,$4)
		ON CONFLICT (tenant_id,user_id) DO UPDATE
		SET email=EXCLUDED.email, role=EXCLUDED.role, updated_at=NOW()
	`, tenantID, userID, input.Email, input.Role); err != nil {
		_, _ = tenantDB.Exec("DELETE FROM users WHERE id=$1", userID)
		respondWithError(w, http.StatusInternalServerError, "Could not create platform account")
		return
	}

	if err = controlTx.Commit(); err != nil {
		_, _ = tenantDB.Exec("DELETE FROM users WHERE id=$1", userID)
		respondWithError(w, http.StatusInternalServerError, "Could not commit user creation")
		return
	}

	respondWithJSON(w, http.StatusCreated, models.ApiResponse{
		Success: true,
		Message: "User created successfully",
		Data: models.User{
			ID: userID, Username: input.Username, Email: input.Email,
			Role: input.Role, TenantID: tenantID, CompanyName: companyName,
			CreatedAt: time.Now(),
		},
	})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, tenantDB, err := tenantUserDB(r)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database is not ready")
		return
	}
	targetID := mux.Vars(r)["id"]

	var currentUsername, currentEmail, currentRole string
	if err := tenantDB.QueryRow(
		`SELECT username,email,role FROM users WHERE id=$1 AND tenant_id=$2`,
		targetID, tenantID,
	).Scan(&currentUsername, &currentEmail, &currentRole); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
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

	update.Username = strings.TrimSpace(update.Username)
	update.Email = strings.ToLower(strings.TrimSpace(update.Email))
	if update.Role != "" && !validOperationalRole(update.Role) {
		respondWithError(w, http.StatusBadRequest, "User role must be manager, recruiter, or team_leader")
		return
	}

	newUsername, newEmail, newRole := currentUsername, currentEmail, currentRole
	if update.Username != "" {
		newUsername = update.Username
	}
	if update.Email != "" {
		newEmail = update.Email
	}
	if update.Role != "" {
		newRole = update.Role
	}

	if _, err := tenantDB.Exec(`
        UPDATE users SET username=$1,email=$2,role=$3
        WHERE id=$4 AND tenant_id=$5
    `, newUsername, newEmail, newRole, targetID, tenantID); err != nil {
		if strings.Contains(err.Error(), "unique") {
			respondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not update tenant user")
		return
	}

	if _, err := db.DB.Exec(`
        UPDATE platform_user_accounts
        SET email=$1,role=$2,updated_at=NOW()
        WHERE user_id=$3 AND tenant_id=$4
    `, newEmail, newRole, targetID, tenantID); err != nil {
		_, _ = tenantDB.Exec(
			`UPDATE users SET username=$1,email=$2,role=$3 WHERE id=$4 AND tenant_id=$5`,
			currentUsername, currentEmail, currentRole, targetID, tenantID,
		)
		respondWithError(w, http.StatusInternalServerError, "Could not update platform account")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "User updated successfully"})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	tenantID, tenantDB, err := tenantUserDB(r)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database is not ready")
		return
	}
	targetID := mux.Vars(r)["id"]

	var username, email, password, role string
	var createdAt time.Time
	if err := tenantDB.QueryRow(
		`SELECT username,email,password,role,created_at FROM users WHERE id=$1 AND tenant_id=$2`,
		targetID, tenantID,
	).Scan(&username, &email, &password, &role, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error looking up user")
		return
	}
	if role == "admin" {
		respondWithError(w, http.StatusForbidden, "The tenant admin cannot be deleted")
		return
	}

	if _, err := tenantDB.Exec("DELETE FROM users WHERE id=$1 AND tenant_id=$2", targetID, tenantID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete tenant user")
		return
	}

	if _, err := db.DB.Exec(
		"DELETE FROM platform_user_accounts WHERE user_id=$1 AND tenant_id=$2",
		targetID, tenantID,
	); err != nil {
		_, _ = tenantDB.Exec(`
            INSERT INTO users(id,username,email,password,role,tenant_id,created_at)
            VALUES($1,$2,$3,$4,$5,$6,$7)
        `, targetID, username, email, password, role, tenantID, createdAt)
		respondWithError(w, http.StatusInternalServerError, "Could not delete platform account")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{Success: true, Message: "User deleted successfully"})
}
