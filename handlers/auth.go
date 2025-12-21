package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"temuin/config"
	"temuin/models"
	"temuin/utils"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func LoginPage(c *gin.Context) {
	tpl, err := pongo2.FromFile("templates/core/login.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template Error: "+err.Error())
		return
	}
	ctx := utils.GetGlobalContext(c)

	// Check for banned user error from query parameter
	if c.Query("error") == "banned" {
		ctx["banned_error"] = "Akun Anda telah diblokir karena melanggar ketentuan platform (penipuan, jual beli, atau pelanggaran lainnya). Silakan hubungi admin jika ada pertanyaan."
	}

	out, err := tpl.Execute(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Render Error: "+err.Error())
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	ctx := utils.GetGlobalContext(c)

	// Validate username is not empty
	if valid, errMsg := utils.ValidateNotEmpty(username, "Username"); !valid {
		ctx["username_error"] = errMsg
		ctx["username"] = username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/login.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	// Validate password is not empty
	if valid, errMsg := utils.ValidateNotEmpty(password, "Password"); !valid {
		ctx["password_error"] = errMsg
		ctx["username"] = username // Preserve username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/login.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	// Check if user exists
	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		ctx["username_error"] = "Username tidak ditemukan"
		ctx["username"] = username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/login.html"))
		out, err := tpl.Execute(ctx)
		if err != nil {
			c.String(http.StatusInternalServerError, "Render Error: "+err.Error())
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		ctx["password_error"] = "Password tidak sesuai"
		ctx["username"] = username // Preserve username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/login.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	// Check if user is banned
	if user.IsBanned {
		ctx["banned_error"] = "Akun Anda telah diblokir karena melanggar ketentuan platform (penipuan, jual beli, atau pelanggaran lainnya). Silakan hubungi admin jika ada pertanyaan."
		ctx["username"] = username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/login.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	// Login successful
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

	c.Redirect(http.StatusFound, "/dashboard")
}

func RegisterPage(c *gin.Context) {
	tpl := pongo2.Must(pongo2.FromFile("templates/core/register.html"))
	out, _ := tpl.Execute(pongo2.Context{})
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}

func Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	ctx := utils.GetGlobalContext(c)

	// Validate username is not empty
	if valid, errMsg := utils.ValidateNotEmpty(username, "Username"); !valid {
		ctx["username_error"] = errMsg
		ctx["username"] = username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/register.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	// Validate password is not empty
	if valid, errMsg := utils.ValidateNotEmpty(password, "Password"); !valid {
		ctx["password_error"] = errMsg
		ctx["username"] = username // Preserve username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/register.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	// Validate password strength
	if valid, errMsg := utils.ValidatePassword(password); !valid {
		ctx["password_error"] = errMsg
		ctx["username"] = username // Preserve username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/register.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    username + "@example.com",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		ctx["error"] = "Username sudah digunakan atau terjadi kesalahan"
		ctx["username"] = username // Preserve username
		tpl := pongo2.Must(pongo2.FromFile("templates/core/register.html"))
		out, _ := tpl.Execute(ctx)
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(out))
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

	c.Redirect(http.StatusFound, "/dashboard")
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

func Profile(c *gin.Context) {
	ctx := utils.GetGlobalContext(c)

	user := c.MustGet("user").(*models.User)

	// ambil items & found_items (logika mu sekarang)...
	var items []models.LostItem
	var foundItems []models.LostItem
	config.DB.Where("user_id = ?", user.ID).Find(&items)
	config.DB.Where("finder_id = ?", user.ID).Find(&foundItems)

	// --- NEW: ambil riwayat topup & withdrawal
	var topups []models.TopUpTransaction
	config.DB.
		Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(20).
		Find(&topups)

	var withdrawals []models.WithdrawalRequest
	config.DB.
		Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(20).
		Find(&withdrawals)

	// Combine and sort
	type TransactionHistoryItem struct {
		Type     string // "TopUp" or "Withdraw"
		Amount   int    // Coins
		Price    int    // RP (for topup) or Amount (for withdraw)
		Method   string
		Status   string
		Date     string
		Original interface{}
	}

	var allTransactions []TransactionHistoryItem

	for _, t := range topups {
		allTransactions = append(allTransactions, TransactionHistoryItem{
			Type:     "TopUp",
			Amount:   t.Amount,
			Price:    t.Price,
			Method:   t.PaymentType,
			Status:   t.Status,
			Date:     t.CreatedAt.Format("2006-01-02 15:04:05"),
			Original: t,
		})
	}

	for _, w := range withdrawals {
		allTransactions = append(allTransactions, TransactionHistoryItem{
			Type:     "Withdraw",
			Amount:   w.Coins,
			Price:    w.Amount, // IDR value
			Method:   w.Method,
			Status:   w.Status,
			Date:     w.CreatedAt.Format("2006-01-02 15:04:05"),
			Original: w,
		})
	}

	// Sort desc by Date string (simple ISO format sort works)
	// Or simplistic bubble sort / slice sort since list is small (max 40)
	// Better: just sort by string comparison since format is YYYY-MM-DD HH:MM:SS
	// We'll use a simple manual sort or rely on client side? No, server side best.
	// Since imports might be tricky for "sort", let's do a simple bubble sort or insertion sort for this small list.
	for i := 0; i < len(allTransactions); i++ {
		for j := i + 1; j < len(allTransactions); j++ {
			if allTransactions[j].Date > allTransactions[i].Date {
				allTransactions[i], allTransactions[j] = allTransactions[j], allTransactions[i]
			}
		}
	}

	// set ctx (sesuaikan nama kunci dengan template)
	ctx["user"] = user
	ctx["items"] = items
	ctx["found_items"] = foundItems
	ctx["transactions"] = allTransactions

	tpl := pongo2.Must(pongo2.FromFile("templates/core/profile.html"))
	out, _ := tpl.Execute(ctx)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}

func TopUp(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	amountStr := c.PostForm("amount")
	amount, _ := strconv.Atoi(amountStr)

	if amount > 0 {
		user.CoinBalance += amount
		config.DB.Save(user)
	}

	c.Redirect(http.StatusFound, "/profile")
}

// generateStateOauthCookie generates a random state string and stores it in session
func generateStateOauthCookie(c *gin.Context) string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	session := sessions.Default(c)
	session.Set("oauth_state", state)
	session.Save()

	return state
}

// GoogleLogin initiates the OAuth flow by redirecting to Google
func GoogleLogin(c *gin.Context) {
	state := generateStateOauthCookie(c)
	url := config.GoogleOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleUserInfo represents the user data from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// GoogleCallback handles the OAuth callback from Google
func GoogleCallback(c *gin.Context) {
	session := sessions.Default(c)
	storedState := session.Get("oauth_state")

	// Verify state to prevent CSRF
	if c.Query("state") != storedState {
		c.String(http.StatusBadRequest, "Invalid oauth state")
		return
	}

	// Exchange authorization code for token
	code := c.Query("code")
	token, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to exchange token: "+err.Error())
		return
	}

	// Get user info from Google
	client := config.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get user info: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse user info: "+err.Error())
		return
	}

	// Check if user exists by email
	var user models.User
	result := config.DB.Where("email = ?", googleUser.Email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		user = models.User{
			Username:  googleUser.Email, // Use email as username
			Email:     googleUser.Email,
			FirstName: googleUser.GivenName,
			LastName:  googleUser.FamilyName,
			Password:  "", // No password for OAuth users
			IsActive:  true,
		}

		if err := config.DB.Create(&user).Error; err != nil {
			c.String(http.StatusInternalServerError, "Failed to create user: "+err.Error())
			return
		}
	}
	// If user exists, just login (no need to update anything)

	// Check if user is banned
	if user.IsBanned {
		c.Redirect(http.StatusFound, "/login?error=banned")
		return
	}

	// Set session
	session.Set("user_id", user.ID)
	session.Delete("oauth_state")
	session.Save()

	c.Redirect(http.StatusFound, "/dashboard")
}
