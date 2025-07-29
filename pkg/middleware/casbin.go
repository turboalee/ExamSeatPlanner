package middleware

import (
	"ExamSeatPlanner/internal/auth"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/casbin/v2/util"
	"github.com/labstack/echo/v4"
)

var (
	enforcer     *casbin.Enforcer
	enforcerOnce sync.Once
)

// getCasbinModel returns the RBAC model as a string (previously in rbac_model.conf)
func getCasbinModel() string {
	modelStr := `[request_definition]
	r = sub, obj, act

	[policy_definition]
	p = sub, obj, act, eft

	[role_definition]
	g = _, _

	[policy_effect]
	e = some(where (p.eft == allow))

	[matchers]
	m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && r.act == p.act`
	log.Println("[DEBUG] Casbin model string loaded:")
	log.Println(modelStr)
	if len(modelStr) < 50 || !containsAllSections(modelStr) {
		panic("[FATAL] Casbin model string is empty or missing required sections!")
	}
	return modelStr
}

// containsAllSections checks for all required Casbin model sections
func containsAllSections(s string) bool {
	sections := []string{"[request_definition]", "[policy_definition]", "[role_definition]", "[policy_effect]", "[matchers]"}
	for _, sec := range sections {
		if !strings.Contains(s, sec) {
			return false
		}
	}
	return true
}

// InitCasbinEnforcer initializes the Casbin enforcer singleton with the model defined in code.
func InitCasbinEnforcer() (*casbin.Enforcer, error) {
	var err error
	enforcerOnce.Do(func() {
		// Defensive check: ensure rbac_policy.csv exists
		if _, statErr := os.Stat("rbac_policy.csv"); os.IsNotExist(statErr) {
			log.Fatalf("[FATAL] rbac_policy.csv not found: %v", statErr)
		}
		m, errM := model.NewModelFromString(getCasbinModel())
		if errM != nil {
			err = errM
			return
		}
		adapter := fileadapter.NewAdapter("rbac_policy.csv")
		enforcer, err = casbin.NewEnforcer(m, adapter)
		if err != nil || enforcer == nil {
			log.Fatalf("[FATAL] Error creating Casbin enforcer: %v", err)
		}
		// Register keyMatch function for path matching
		enforcer.AddFunction("keyMatch", util.KeyMatchFunc)
		policies, _ := enforcer.GetPolicy()
		log.Printf("Casbin enforcer created. Policy count: %d", len(policies))
	})
	return enforcer, err
}

// CasbinMiddleware enforces RBAC using Casbin for each request.
func CasbinMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*auth.JWTClaims)
		if !ok || claims == nil {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized: missing user claims"})
		}
		enf, err := InitCasbinEnforcer()
		if err != nil {
			log.Println("Casbin enforcer error:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "RBAC system error"})
		}
		role := claims.Role
		obj := c.Request().URL.Path // Use actual request path for object
		act := c.Request().Method
		log.Printf("Casbin enforce: role=%s, obj=%s, act=%s", role, obj, act)
		log.Printf("Types: role=%T, obj=%T, act=%T", role, obj, act)
		allowed, err := enf.Enforce(role, obj, act)
		if err != nil {
			log.Println("Casbin enforce error:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "RBAC system error"})
		}
		if !allowed {
			log.Printf("Casbin denied: role=%s, obj=%s, act=%s", role, obj, act)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: insufficient permissions"})
		}
		log.Printf("Casbin allowed: role=%s, obj=%s, act=%s", role, obj, act)
		return next(c)
	}
}
