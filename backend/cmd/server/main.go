package main

import (
	"log"
	"net/http"
	"time"

	"expense-tracker/backend/internal/auth"
	"expense-tracker/backend/internal/config"
	"expense-tracker/backend/internal/db"
	"expense-tracker/backend/internal/handlers"
	appmw "expense-tracker/backend/internal/middleware"
	"expense-tracker/backend/internal/repository"
	"expense-tracker/backend/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	gdb, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := db.Migrate(gdb); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := db.Seed(gdb); err != nil {
		log.Fatalf("seed: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(gdb)
	catRepo := repository.NewCategoryRepository(gdb)
	expRepo := repository.NewExpenseRepository(gdb)
	incRepo := repository.NewIncomeRepository(gdb)
	setRepo := repository.NewSettingRepository(gdb)
	tokRepo := repository.NewTokenRepository(gdb)

	// Services / auth
	tm := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	dashSvc := services.NewDashboardService(userRepo, expRepo, incRepo, setRepo)

	// Handlers
	authH := handlers.NewAuthHandler(userRepo, catRepo, tokRepo, tm, cfg.AdminEmail, cfg.CookieSecure)
	meH := handlers.NewMeHandler(userRepo)
	catH := handlers.NewCategoryHandler(catRepo)
	expH := handlers.NewExpenseHandler(expRepo, catRepo)
	incH := handlers.NewIncomeHandler(incRepo)
	dashH := handlers.NewDashboardHandler(dashSvc)
	trendsH := handlers.NewTrendsHandler(dashSvc)
	// Rate limiters (per client IP):
	//  - global: broad DoS/abuse guard across the whole API
	//  - auth:   tighter budget for unauthenticated /auth/* endpoints
	//  - login:  per-EMAIL brute-force throttle (keyed inside the handler)
	globalLimiter := appmw.NewRateLimiter(240, 60)
	authLimiter := appmw.NewRateLimiter(20, 5)
	loginLimiter := appmw.NewRateLimiter(8, 4)

	adminH := handlers.NewAdminHandler(userRepo, setRepo, tokRepo)
	authH.SetLoginLimiter(loginLimiter)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(appmw.SecureHeaders(cfg.IsProduction))
	r.Use(appmw.CORS(cfg.CORSOrigins))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(api chi.Router) {
		// Broad per-IP limiter on every API request.
		api.Use(globalLimiter.Middleware)

		// Public auth routes (extra-tight rate limit).
		api.Group(func(pub chi.Router) {
			pub.Use(authLimiter.Middleware)
			pub.Post("/auth/register", authH.Register)
			pub.Post("/auth/login", authH.Login)
			pub.Post("/auth/refresh", authH.Refresh)
			pub.Post("/auth/logout", authH.Logout)
		})

		// Authenticated routes.
		api.Group(func(priv chi.Router) {
			priv.Use(appmw.RequireAuth(tm, userRepo))

			priv.Get("/me", meH.Get)
			priv.Patch("/me", meH.UpdateIncome)

			priv.Get("/categories", catH.List)
			priv.Post("/categories", catH.Create)
			priv.Patch("/categories/{id}", catH.Update)
			priv.Delete("/categories/{id}", catH.Delete)

			priv.Get("/expenses", expH.List)
			priv.Post("/expenses", expH.Create)
			priv.Patch("/expenses/{id}", expH.Update)
			priv.Delete("/expenses/{id}", expH.Delete)

			priv.Get("/incomes", incH.List)
			priv.Post("/incomes", incH.Create)
			priv.Patch("/incomes/{id}", incH.Update)
			priv.Delete("/incomes/{id}", incH.Delete)

			priv.Get("/dashboard", dashH.Get)
			priv.Get("/trends", trendsH.Get)

			// Admin-only.
			priv.Group(func(adm chi.Router) {
				adm.Use(appmw.RequireAdmin)
				adm.Get("/admin/users", adminH.ListUsers)
				adm.Patch("/admin/users/{id}/status", adminH.UpdateUserStatus)
				adm.Delete("/admin/users/{id}", adminH.DeleteUser)
				adm.Get("/admin/settings", adminH.GetSettings)
				adm.Put("/admin/settings", adminH.UpdateSettings)
			})
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second, // mitigates Slowloris
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("expense-tracker API listening on :%s (env=%s)", cfg.Port, envLabel(cfg.IsProduction))
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func envLabel(prod bool) string {
	if prod {
		return "production"
	}
	return "development"
}
