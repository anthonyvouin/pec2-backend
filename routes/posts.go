package routes

import (
	"pec2-backend/handlers/posts"
	"pec2-backend/handlers/posts/comment"
	"pec2-backend/handlers/posts/likes"
	"pec2-backend/handlers/posts/report"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func PostsRoutes(r *gin.Engine) { // Routes publiques
	// r.GET("/posts", posts.GetAllPosts)
	r.GET("/posts/:id", posts.GetPostByID)
	// J'ai pas trouvé la solution pour faire la vérification avec le middleware
	// J'ai l'impression qu'en SSE on peut pas envoyer de token dans le header
	// Du coup middleware = useless
	r.GET("/posts/:id/comments/sse", comment.HandleSSE)

	// Routes protégées
	postsRoutes := r.Group("/posts")
	postsRoutes.Use(middleware.JWTAuth())
	{
		postsRoutes.GET("", posts.GetAllPosts)
		// postsRoutes.GET("/:id", posts.GetPostByID)
		postsRoutes.POST("/:id/comments", comment.CreateComment)
		postsRoutes.GET("/:id/comments", comment.GetCommentsByPostID)
		postsRoutes.POST("", posts.CreatePost)
		postsRoutes.PUT("/:id", posts.UpdatePost)
		postsRoutes.DELETE("/:id", posts.DeletePost)

		postsRoutes.GET("/statistics", middleware.AdminAuth(), posts.GetPostsStatistics)

		// Routes des interactions
		postsRoutes.POST("/:id/like", likes.ToggleLike)
		postsRoutes.POST("/:id/report", report.ReportPost)
	}
}
