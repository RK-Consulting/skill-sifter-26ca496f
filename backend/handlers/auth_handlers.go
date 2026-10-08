package handlers

import (
	"net/http"

	"github.com/RK-Consulting/skill-sifter/models"
)

// GetUsers fetches tenant-local users. Platform routing metadata is not a
// substitute for the tenant user record.
func GetUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, tenantDB, err := tenantUserDB(r)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Tenant database is not ready")
		return
	}

	users := []models.User{}
	rows, err := tenantDB.Query(`
        SELECT id, username, email, role, tenant_id, created_at
        FROM users
        WHERE tenant_id=$1
        ORDER BY id
    `, tenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching users")
		return
	}
	defer rows.Close()

	companyName, _ := r.Context().Value("companyName").(string)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.TenantID, &u.CreatedAt); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error scanning user row")
			return
		}
		u.CompanyName = companyName
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error reading users")
		return
	}

	respondWithJSON(w, http.StatusOK, models.ApiResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}
