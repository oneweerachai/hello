package middleware

import (
	"fmt"
	"log"
	"time"
	"user-api/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

// Logger middleware for logging HTTP requests with trace correlation
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		traceID := tracing.GetTraceID(param.Request.Context())
		spanID := tracing.GetSpanID(param.Request.Context())

		logMsg := fmt.Sprintf("[%s] %s %s %d %s %s",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)

		if traceID != "" {
			logMsg += fmt.Sprintf(" trace_id=%s", traceID)
		}
		if spanID != "" {
			logMsg += fmt.Sprintf(" span_id=%s", spanID)
		}

		log.Println(logMsg)
		return ""
	})
}

// TracingMiddleware returns OpenTelemetry tracing middleware
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// EnhancedTracingMiddleware adds additional tracing attributes
func EnhancedTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the current span
		span := trace.SpanFromContext(c.Request.Context())

		// Add request attributes
		span.SetAttributes(
			tracing.AttrHTTPMethod.String(c.Request.Method),
			tracing.AttrHTTPURL.String(c.Request.URL.String()),
			tracing.AttrHTTPUserAgent.String(c.Request.UserAgent()),
			tracing.AttrHTTPClientIP.String(c.ClientIP()),
		)

		// Add request size if available
		if c.Request.ContentLength > 0 {
			span.SetAttributes(tracing.AttrRequestSize.Int64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Add response attributes
		span.SetAttributes(
			tracing.AttrHTTPStatusCode.Int(c.Writer.Status()),
			tracing.AttrResponseSize.Int(c.Writer.Size()),
		)

		// Record error if status code indicates an error
		if c.Writer.Status() >= 400 {
			span.SetAttributes(
				tracing.AttrErrorType.String("http_error"),
				tracing.AttrErrorMessage.String(fmt.Sprintf("HTTP %d", c.Writer.Status())),
			)
		}
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// JSONContentType middleware ensures content type is application/json for POST/PUT requests
func JSONContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" {
				c.JSON(400, gin.H{
					"status":  "error",
					"message": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// Recovery middleware for handling panics with tracing
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		traceID := tracing.GetTraceID(c.Request.Context())
		spanID := tracing.GetSpanID(c.Request.Context())

		// Log panic with trace correlation
		logMsg := fmt.Sprintf("Panic recovered: %v", recovered)
		if traceID != "" {
			logMsg += fmt.Sprintf(" trace_id=%s", traceID)
		}
		if spanID != "" {
			logMsg += fmt.Sprintf(" span_id=%s", spanID)
		}
		log.Println(logMsg)

		// Record error in span
		span := trace.SpanFromContext(c.Request.Context())
		if span.IsRecording() {
			span.SetAttributes(
				tracing.AttrErrorType.String("panic"),
				tracing.AttrErrorMessage.String(fmt.Sprintf("%v", recovered)),
			)
			span.RecordError(fmt.Errorf("panic: %v", recovered))
		}

		response := gin.H{
			"status":  "error",
			"message": "Internal server error",
		}

		// Add trace ID to response if available
		if traceID != "" {
			response["trace_id"] = traceID
		}

		c.JSON(500, response)
	})
}
