package handlers

import (
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

    var user models.User
    if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
        ctx := utils.GetGlobalContext(c)
        ctx["error"] = "Invalid username or password" 
        // Note: Ideally we pass error to template.
        tpl := pongo2.Must(pongo2.FromFile("templates/core/login.html"))
        out, err := tpl.Execute(ctx)
    if err != nil {
         c.String(http.StatusInternalServerError, "Render Error: " + err.Error())
         return
    }
    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
        return
    }

    err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
         ctx := utils.GetGlobalContext(c)
         ctx["error"] = "Invalid username or password"
         tpl := pongo2.Must(pongo2.FromFile("templates/core/login.html"))
         out, _ := tpl.Execute(ctx)
         c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
         return
    }

    session := sessions.Default(c)
    session.Set("user_id", user.ID)
    session.Save()
    
    // Redirect to 'next' if present? For now just home.
    c.Redirect(http.StatusFound, "/")
}

func RegisterPage(c *gin.Context) {
    tpl := pongo2.Must(pongo2.FromFile("templates/core/register.html"))
    out, _ := tpl.Execute(pongo2.Context{})
    c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}

func Register(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

    user := models.User{
        Username: username,
        Password: string(hashedPassword),
        Email:    username + "@example.com", 
    }

    if err := config.DB.Create(&user).Error; err != nil {
        tpl := pongo2.Must(pongo2.FromFile("templates/core/register.html"))
        ctx := pongo2.Context{"error": "Registration failed: " + err.Error()}
        out, _ := tpl.Execute(ctx)
        c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(out))
        return
    }

    session := sessions.Default(c)
    session.Set("user_id", user.ID)
    session.Save()

    c.Redirect(http.StatusFound, "/")
}

func Logout(c *gin.Context) {
    session := sessions.Default(c)
    session.Clear()
    session.Save()
    c.Redirect(http.StatusFound, "/")
}

func Profile(c *gin.Context) {
    user := c.MustGet("user").(*models.User)
    
    // Fetch user items
    var myItems []models.LostItem
    config.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&myItems)
    
    // Fetch items found by user
    var foundItems []models.LostItem
    config.DB.Where("finder_id = ? AND finder_confirmed = ?", user.ID, true).Order("updated_at desc").Find(&foundItems)

    ctx := utils.GetGlobalContext(c)
    ctx["items"] = myItems
    ctx["found_items"] = foundItems
    ctx["user"] = user
    // Assuming simple profile that lists user's posts
    // We might need a profile.html, or reuse home/dashboard?
    // Let's reuse home with a filter or a specific profile template?
    // User previously had profile logic. Let's assume there is a profile.html or we use dashboard.
    // Let's create a basic profile render.
    
    // Check if profile.html exists? If not, use home logic but filtered.
    // For this migration, let's use a dedicated simplistic render or error if template missing.
    // Or better: Re-use home template but just passing 'items' as myItems is what dashboard basically is.
    
    // Let's try loading profile.html. If it fails (panic in Must), we know.
    // Actually, let's just render a simple view.
    // Wait, the sidebar links to /profile/.
    
    tpl, err := pongo2.FromFile("templates/core/profile.html")
    if err != nil {
        // Fallback to home template with title "My Profile"
        ctx["header_title"] = "My Profile"
        tpl, err = pongo2.FromFile("templates/core/home.html")
        if err != nil {
            c.String(http.StatusInternalServerError, "Template Error (Home Fallback): " + err.Error())
            return
        }
    }
    
    out, err := tpl.Execute(ctx)
    if err != nil {
         c.String(http.StatusInternalServerError, "Render Error: " + err.Error())
         return
    }
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
