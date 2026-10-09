package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/yourname/novapromptgobackend/internal/handlers"
	custommw "github.com/yourname/novapromptgobackend/internal/middleware"
	"github.com/yourname/novapromptgobackend/internal/repository"
)

type Deps struct {
	Health       *handlers.HealthHandler
	Images       *handlers.ImageHandler
	Categories   *handlers.CategoryHandler
	Tags         *handlers.TagHandler
	Descriptions *handlers.DescriptionHandler
	AdKeys       *handlers.AdKeyHandler
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(custommw.Recoverer)
	r.Use(custommw.RequestLogger)
	r.Use(chimw.Timeout(10_000_000_000)) 
	r.Use(custommw.MaxBody(1 << 20))      

	r.Get("/health", d.Health.Live)
	r.Get("/ready", d.Health.Ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/images", func(r chi.Router) {
			r.Get("/", d.Images.List)
			r.Post("/", d.Images.Create)
			r.Get("/{id}", d.Images.Get)
			r.Put("/{id}", d.Images.Update)
			r.Delete("/{id}", d.Images.Delete)

			r.Get("/{imageID}/description", d.Descriptions.Get)
			r.Put("/{imageID}/description", d.Descriptions.Upsert)
			r.Delete("/{imageID}/description", d.Descriptions.Delete)
		})
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", d.Categories.List)
			r.Post("/", d.Categories.Create)
			r.Get("/{id}", d.Categories.Get)
			r.Put("/{id}", d.Categories.Update)
			r.Delete("/{id}", d.Categories.Delete)
		})
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", d.Tags.List)
			r.Post("/", d.Tags.Create)
			r.Put("/{id}", d.Tags.Update)
			r.Delete("/{id}", d.Tags.Delete)
		})
		r.Route("/keys", func(r chi.Router) {
			r.Get("/", d.AdKeys.List)
			r.Get("/{key}", d.AdKeys.Get)
			r.Put("/{key}", d.AdKeys.Upsert)
		})
	})

	return r
}

var _ = repository.ErrNotFound 