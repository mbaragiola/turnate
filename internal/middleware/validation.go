package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// Security patterns for input validation
	sqlInjectionPattern     = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|javascript|onload|onerror|eval|alert|confirm|prompt)`)
	xssPattern             = regexp.MustCompile(`(?i)(<script|javascript:|onload=|onerror=|onclick=|onmouseover=|<iframe|<object|<embed|<link)`)
	commandInjectionPattern = regexp.MustCompile(`(;\s*rm\s+|;\s*cat\s+|;\s*ls\s+|;\s*wget\s+|;\s*curl\s+|&&|&|\|\||;)`)
	pathTraversalPattern   = regexp.MustCompile(`(\.\.\/|\.\.\\|%2e%2e%2f|%2e%2e%5c)`)
)

// InputValidationMiddleware validates and sanitizes input data
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if !isValidInput(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "Invalid input detected",
						"message": "Request contains potentially malicious content",
						"field":   key,
					})
					c.Abort()
					return
				}
			}
		}

		// Check form values if present
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for _, value := range values {
						if !isValidInput(value) {
							c.JSON(http.StatusBadRequest, gin.H{
								"error":   "Invalid input detected",
								"message": "Request contains potentially malicious content",
								"field":   key,
							})
							c.Abort()
							return
						}
					}
				}
			}
		}

		// Check path parameters
		for _, param := range c.Params {
			if !isValidInput(param.Value) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid path parameter",
					"message": "Path contains potentially malicious content",
					"field":   param.Key,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ContentSecurityMiddleware adds content security headers
func ContentSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		c.Header("Content-Security-Policy", 
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://code.jquery.com; "+
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
			"font-src 'self' https://cdn.jsdelivr.net; "+
			"img-src 'self' data: https:; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none';")
		
		// Additional security headers
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// isValidInput checks if input contains potentially malicious patterns
func isValidInput(input string) bool {
	input = strings.ToLower(input)
	
	// Check for SQL injection patterns
	if sqlInjectionPattern.MatchString(input) {
		return false
	}
	
	// Check for XSS patterns
	if xssPattern.MatchString(input) {
		return false
	}
	
	// Check for command injection patterns
	if commandInjectionPattern.MatchString(input) {
		return false
	}
	
	// Check for path traversal patterns
	if pathTraversalPattern.MatchString(input) {
		return false
	}
	
	return true
}

// SanitizeString removes potentially harmful characters from strings
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters except tab, newline, and carriage return
	var result strings.Builder
	for _, r := range input {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(result.String())
}

// ValidateContentType ensures the request has the correct content type for JSON APIs
func ValidateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "multipart/form-data") {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":   "Unsupported content type",
					"message": "Expected application/json or multipart/form-data",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}