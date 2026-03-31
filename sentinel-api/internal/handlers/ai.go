package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sentinel-proxy/api/internal/storage"
)

type AIHandler struct {
	storage *storage.Storage
}

func NewAIHandler(s *storage.Storage) *AIHandler {
	return &AIHandler{storage: s}
}

// Predefined question types
const (
	QuestionUnusualAccess     = "unusual_access"
	QuestionUnusualPayload    = "unusual_payload"
	QuestionHighErrorRate     = "high_error_rate"
	QuestionRateSpike         = "rate_spike"
	QuestionSuspiciousUA      = "suspicious_user_agent"
	QuestionFailedLogins      = "failed_logins"
	QuestionSensitivePaths    = "sensitive_paths"
	QuestionTopAttackers      = "top_attackers"
	QuestionRecentIncidents   = "recent_incidents"
)

type AIQuestionRequest struct {
	QuestionID string `json:"questionId"`
	TimeRange  string `json:"timeRange,omitempty"` // "1h", "24h", "7d"
}

type AIQuestionResponse struct {
	QuestionID   string      `json:"questionId"`
	Question     string      `json:"question"`
	Answer       string      `json:"answer"`
	Confidence   string      `json:"confidence"` // "high", "medium", "low"
	Data         interface{} `json:"data,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

func (h *AIHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AIQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate question ID
	if !isValidQuestionID(req.QuestionID) {
		http.Error(w, "Invalid question ID", http.StatusBadRequest)
		return
	}

	// Calculate time range
	timeRange := parseTimeRange(req.TimeRange)
	fromTime := time.Now().Add(-timeRange)

	// Route to appropriate handler
	var response AIQuestionResponse
	var err error

	switch req.QuestionID {
	case QuestionUnusualAccess:
		response, err = h.analyzeUnusualAccess(ctx, fromTime)
	case QuestionUnusualPayload:
		response, err = h.analyzeUnusualPayload(ctx, fromTime)
	case QuestionHighErrorRate:
		response, err = h.analyzeHighErrorRate(ctx, fromTime)
	case QuestionRateSpike:
		response, err = h.analyzeRateSpike(ctx, fromTime)
	case QuestionSuspiciousUA:
		response, err = h.analyzeSuspiciousUserAgent(ctx, fromTime)
	case QuestionFailedLogins:
		response, err = h.analyzeFailedLogins(ctx, fromTime)
	case QuestionSensitivePaths:
		response, err = h.analyzeSensitivePaths(ctx, fromTime)
	case QuestionTopAttackers:
		response, err = h.analyzeTopAttackers(ctx, fromTime)
	case QuestionRecentIncidents:
		response, err = h.analyzeRecentIncidents(ctx)
	default:
		http.Error(w, "Unknown question ID", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func isValidQuestionID(id string) bool {
	validIDs := []string{
		QuestionUnusualAccess,
		QuestionUnusualPayload,
		QuestionHighErrorRate,
		QuestionRateSpike,
		QuestionSuspiciousUA,
		QuestionFailedLogins,
		QuestionSensitivePaths,
		QuestionTopAttackers,
		QuestionRecentIncidents,
	}
	for _, validID := range validIDs {
		if id == validID {
			return true
		}
	}
	return false
}

func parseTimeRange(tr string) time.Duration {
	switch tr {
	case "1h":
		return 1 * time.Hour
	case "24h":
		return 24 * time.Hour
	case "7d":
		return 7 * 24 * time.Hour
	default:
		return 24 * time.Hour // Default to 24h
	}
}

func (h *AIHandler) analyzeUnusualAccess(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	// Get logs from time range
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	// Analyze IP access patterns
	ipCounts := make(map[string]int)
	ipPaths := make(map[string]map[string]int)

	for _, log := range logs {
		client, _ := log["client"].(bson.M)
		if client == nil {
			continue
		}
		ip, _ := client["ip"].(string)
		if ip == "" {
			continue
		}

		request, _ := log["request"].(bson.M)
		path := ""
		if request != nil {
			path, _ = request["path"].(string)
		}

		ipCounts[ip]++
		if ipPaths[ip] == nil {
			ipPaths[ip] = make(map[string]int)
		}
		ipPaths[ip][path]++
	}

	// Find unusual patterns (IPs accessing many different paths)
	unusualIPs := []map[string]interface{}{}
	for ip, totalRequests := range ipCounts {
		uniquePaths := len(ipPaths[ip])
		// Unusual if accessing many different paths with relatively few requests
		if uniquePaths > 10 && totalRequests < uniquePaths*2 {
			unusualIPs = append(unusualIPs, map[string]interface{}{
				"ip":           ip,
				"totalRequests": totalRequests,
				"uniquePaths":  uniquePaths,
			})
		}
	}

	answer := "No unusual access patterns detected."
	confidence := "high"
	if len(unusualIPs) > 0 {
		answer = "Found " + strconv.Itoa(len(unusualIPs)) + " IP(s) with unusual access patterns (accessing many different paths)."
		confidence = "high"
	}

	return AIQuestionResponse{
		QuestionID: QuestionUnusualAccess,
		Question:   "Is there unusual access to pages?",
		Answer:     answer,
		Confidence:  confidence,
		Data: map[string]interface{}{
			"unusualIPs": unusualIPs[:min(10, len(unusualIPs))],
			"totalIPs":   len(ipCounts),
		},
		Recommendations: h.getRecommendationsForUnusualAccess(len(unusualIPs)),
	}, nil
}

func (h *AIHandler) analyzeUnusualPayload(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	unusualPayloads := []map[string]interface{}{}
	largeBodies := 0

	for _, log := range logs {
		request, _ := log["request"].(bson.M)
		if request == nil {
			continue
		}

		bodySize, _ := request["bodySize"].(int32)
		body, _ := request["body"].(string)

		// Check for large payloads (>100KB)
		if bodySize > 100*1024 {
			largeBodies++
			client, _ := log["client"].(bson.M)
			ip := ""
			if client != nil {
				ip, _ = client["ip"].(string)
			}
			path, _ := request["path"].(string)

			unusualPayloads = append(unusualPayloads, map[string]interface{}{
				"ip":       ip,
				"path":     path,
				"bodySize": bodySize,
				"preview":  truncateString(body, 200),
			})
		}

		// Check for suspicious patterns in body
		if body != "" {
			bodyLower := strings.ToLower(body)
			suspiciousPatterns := []string{"<script", "union select", "drop table", "exec(", "eval("}
			for _, pattern := range suspiciousPatterns {
				if strings.Contains(bodyLower, pattern) {
					client, _ := log["client"].(bson.M)
					ip := ""
					if client != nil {
						ip, _ = client["ip"].(string)
					}
					path, _ := request["path"].(string)

					unusualPayloads = append(unusualPayloads, map[string]interface{}{
						"ip":       ip,
						"path":     path,
						"pattern":  pattern,
						"bodySize": bodySize,
						"preview":  truncateString(body, 200),
					})
					break
				}
			}
		}
	}

	answer := "No unusual payloads detected."
	confidence := "high"
	if len(unusualPayloads) > 0 {
		answer = "Found " + strconv.Itoa(len(unusualPayloads)) + " request(s) with unusual payloads (large bodies or suspicious patterns)."
		confidence = "high"
	}

	return AIQuestionResponse{
		QuestionID: QuestionUnusualPayload,
		Question:   "Is there unusual payload sent to requests?",
		Answer:     answer,
		Confidence:  confidence,
		Data: map[string]interface{}{
			"unusualPayloads": unusualPayloads[:min(10, len(unusualPayloads))],
			"largeBodies":     largeBodies,
		},
		Recommendations: h.getRecommendationsForUnusualPayload(len(unusualPayloads)),
	}, nil
}

func (h *AIHandler) analyzeHighErrorRate(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	totalRequests := len(logs)
	errorRequests := 0
	errorByPath := make(map[string]int)
	errorByIP := make(map[string]int)

	for _, log := range logs {
		response, _ := log["response"].(bson.M)
		if response == nil {
			continue
		}
		status, _ := response["status"].(int32)
		if status >= 400 {
			errorRequests++
			request, _ := log["request"].(bson.M)
			if request != nil {
				path, _ := request["path"].(string)
				errorByPath[path]++
			}
			client, _ := log["client"].(bson.M)
			if client != nil {
				ip, _ := client["ip"].(string)
				errorByIP[ip]++
			}
		}
	}

	errorRate := float64(errorRequests) / float64(totalRequests) * 100
	answer := "Error rate is normal."
	confidence := "high"
	if errorRate > 10 {
		answer = "High error rate detected: " + strconv.FormatFloat(errorRate, 'f', 2, 64) + "% of requests are errors."
		confidence = "high"
	}

	return AIQuestionResponse{
		QuestionID: QuestionHighErrorRate,
		Question:   "Is there a high error rate?",
		Answer:     answer,
		Confidence:  confidence,
		Data: map[string]interface{}{
			"errorRate":    errorRate,
			"totalRequests": totalRequests,
			"errorRequests": errorRequests,
			"topErrorPaths": getTopN(errorByPath, 5),
			"topErrorIPs":   getTopN(errorByIP, 5),
		},
		Recommendations: h.getRecommendationsForHighErrorRate(errorRate),
	}, nil
}

func (h *AIHandler) analyzeRateSpike(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	// Group by hour
	hourlyCounts := make(map[string]int)
	for _, log := range logs {
		var t time.Time
		timestamp := log["timestamp"]
		switch v := timestamp.(type) {
		case primitive.DateTime:
			t = time.Unix(int64(v)/1000, 0)
		case time.Time:
			t = v
		case primitive.Timestamp:
			t = time.Unix(int64(v.T), 0)
		default:
			continue
		}
		hourKey := t.Format("2006-01-02 15:00")
		hourlyCounts[hourKey]++
	}

	// Find spike (hour with significantly more requests)
	maxCount := 0
	maxHour := ""
	avgCount := 0
	if len(hourlyCounts) > 0 {
		total := 0
		for _, count := range hourlyCounts {
			total += count
			if count > maxCount {
				maxCount = count
			}
		}
		avgCount = total / len(hourlyCounts)
	}

	answer := "No significant rate spikes detected."
	confidence := "medium"
	if maxCount > avgCount*2 && avgCount > 0 {
		answer = "Rate spike detected: " + strconv.Itoa(maxCount) + " requests in " + maxHour + " (average: " + strconv.Itoa(avgCount) + ")."
		confidence = "high"
	}

	return AIQuestionResponse{
		QuestionID: QuestionRateSpike,
		Question:   "Is there a rate spike?",
		Answer:     answer,
		Confidence:  confidence,
		Data: map[string]interface{}{
			"maxCount":  maxCount,
			"maxHour":   maxHour,
			"avgCount":  avgCount,
			"hourlyData": hourlyCounts,
		},
		Recommendations: h.getRecommendationsForRateSpike(maxCount > avgCount*2),
	}, nil
}

func (h *AIHandler) analyzeSuspiciousUserAgent(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	suspiciousUAs := []map[string]interface{}{}
	suspiciousPatterns := []string{"curl", "python-requests", "wget", "scanner", "bot", "crawler", "spider"}

	uaCounts := make(map[string]int)
	uaIPs := make(map[string][]string)

	for _, log := range logs {
		client, _ := log["client"].(bson.M)
		if client == nil {
			continue
		}
		ua, _ := client["userAgent"].(string)
		ip, _ := client["ip"].(string)
		if ua == "" {
			continue
		}

		uaLower := strings.ToLower(ua)
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(uaLower, pattern) {
				uaCounts[ua]++
				if !contains(uaIPs[ua], ip) {
					uaIPs[ua] = append(uaIPs[ua], ip)
				}
				break
			}
		}
	}

	for ua, count := range uaCounts {
		suspiciousUAs = append(suspiciousUAs, map[string]interface{}{
			"userAgent": ua,
			"count":     count,
			"ips":       uaIPs[ua][:min(5, len(uaIPs[ua]))],
		})
	}

	answer := "No suspicious user agents detected."
	confidence := "high"
	if len(suspiciousUAs) > 0 {
		answer = "Found " + strconv.Itoa(len(suspiciousUAs)) + " suspicious user agent(s)."
		confidence = "medium"
	}

	return AIQuestionResponse{
		QuestionID: QuestionSuspiciousUA,
		Question:   "Are there suspicious user agents?",
		Answer:     answer,
		Confidence:  confidence,
		Data: map[string]interface{}{
			"suspiciousUAs": suspiciousUAs[:min(10, len(suspiciousUAs))],
		},
		Recommendations: h.getRecommendationsForSuspiciousUA(len(suspiciousUAs)),
	}, nil
}

func (h *AIHandler) analyzeFailedLogins(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
		"request.path": bson.M{"$regex": "(?i)(login|auth|signin)"},
		"response.status": bson.M{"$gte": 400, "$lt": 500},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	failedByIP := make(map[string]int)
	for _, log := range logs {
		client, _ := log["client"].(bson.M)
		if client != nil {
			ip, _ := client["ip"].(string)
			failedByIP[ip]++
		}
	}

	bruteForceIPs := []map[string]interface{}{}
	for ip, count := range failedByIP {
		if count >= 5 {
			bruteForceIPs = append(bruteForceIPs, map[string]interface{}{
				"ip":    ip,
				"count": count,
			})
		}
	}

	answer := "No failed login attempts detected."
	confidence := "high"
	if len(bruteForceIPs) > 0 {
		answer = "Found " + strconv.Itoa(len(bruteForceIPs)) + " IP(s) with multiple failed login attempts (potential brute force)."
		confidence = "high"
	}

	return AIQuestionResponse{
		QuestionID: QuestionFailedLogins,
		Question:   "Are there failed login attempts?",
		Answer:     answer,
		Confidence:  confidence,
		Data: map[string]interface{}{
			"bruteForceIPs": bruteForceIPs,
			"totalFailed":   len(logs),
		},
		Recommendations: h.getRecommendationsForFailedLogins(len(bruteForceIPs)),
	}, nil
}

func (h *AIHandler) analyzeSensitivePaths(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	sensitivePatterns := []string{"/admin", "/api", "/login", "/wp-admin", "/.env", "/config", "/backup"}
	accessByPath := make(map[string]int)
	accessByIP := make(map[string]map[string]int)

	for _, log := range logs {
		request, _ := log["request"].(bson.M)
		if request == nil {
			continue
		}
		path, _ := request["path"].(string)
		client, _ := log["client"].(bson.M)
		ip := ""
		if client != nil {
			ip, _ = client["ip"].(string)
		}

		for _, pattern := range sensitivePatterns {
			if strings.Contains(path, pattern) {
				accessByPath[path]++
				if accessByIP[ip] == nil {
					accessByIP[ip] = make(map[string]int)
				}
				accessByIP[ip][path]++
				break
			}
		}
	}

	sensitiveAccess := []map[string]interface{}{}
	for path, count := range accessByPath {
		sensitiveAccess = append(sensitiveAccess, map[string]interface{}{
			"path":  path,
			"count": count,
		})
	}

	answer := "No unusual access to sensitive paths detected."
	confidence := "high"
	if len(sensitiveAccess) > 0 {
		answer = "Found access to " + strconv.Itoa(len(sensitiveAccess)) + " sensitive path(s)."
		confidence = "medium"
	}

	return AIQuestionResponse{
		QuestionID: QuestionSensitivePaths,
		Question:   "Is there access to sensitive paths?",
		Answer:     answer,
		Confidence:  confidence,
		Data: map[string]interface{}{
			"sensitiveAccess": sensitiveAccess[:min(10, len(sensitiveAccess))],
		},
		Recommendations: h.getRecommendationsForSensitivePaths(len(sensitiveAccess)),
	}, nil
}

func (h *AIHandler) analyzeTopAttackers(ctx context.Context, fromTime time.Time) (AIQuestionResponse, error) {
	filter := bson.M{
		"timestamp": bson.M{"$gte": fromTime},
		"response.status": bson.M{"$gte": 400},
	}
	logs, _, err := h.storage.GetLogs(ctx, filter, 1000, "")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	ipCounts := make(map[string]int)
	for _, log := range logs {
		client, _ := log["client"].(bson.M)
		if client != nil {
			ip, _ := client["ip"].(string)
			ipCounts[ip]++
		}
	}

	topAttackers := getTopN(ipCounts, 10)

	answer := "No significant attackers detected."
	if len(topAttackers) > 0 {
		answer = "Top " + strconv.Itoa(len(topAttackers)) + " IP(s) generating errors identified."
	}

	return AIQuestionResponse{
		QuestionID: QuestionTopAttackers,
		Question:   "Who are the top attackers?",
		Answer:     answer,
		Confidence: "high",
		Data: map[string]interface{}{
			"topAttackers": topAttackers,
		},
		Recommendations: h.getRecommendationsForTopAttackers(len(topAttackers)),
	}, nil
}

func (h *AIHandler) analyzeRecentIncidents(ctx context.Context) (AIQuestionResponse, error) {
	incidents, err := h.storage.GetIncidents(ctx, "open")
	if err != nil {
		return AIQuestionResponse{}, err
	}

	recentIncidents := []map[string]interface{}{}
	now := time.Now()
	for _, incident := range incidents {
		var incidentTime time.Time
		timestamp := incident["timestamp"]
		switch v := timestamp.(type) {
		case primitive.DateTime:
			incidentTime = time.Unix(int64(v)/1000, 0)
		case time.Time:
			incidentTime = v
		case primitive.Timestamp:
			incidentTime = time.Unix(int64(v.T), 0)
		default:
			continue
		}
		if now.Sub(incidentTime) < 24*time.Hour {
			recentIncidents = append(recentIncidents, map[string]interface{}{
				"ruleName":  incident["ruleName"],
				"ip":        incident["ip"],
				"action":    incident["actionTaken"],
				"timestamp": incidentTime.Format(time.RFC3339),
			})
		}
	}

	answer := "No recent incidents."
	if len(recentIncidents) > 0 {
		answer = "Found " + strconv.Itoa(len(recentIncidents)) + " recent incident(s) in the last 24 hours."
	}

	return AIQuestionResponse{
		QuestionID: QuestionRecentIncidents,
		Question:   "What are the recent incidents?",
		Answer:     answer,
		Confidence: "high",
		Data: map[string]interface{}{
			"recentIncidents": recentIncidents,
		},
		Recommendations: h.getRecommendationsForRecentIncidents(len(recentIncidents)),
	}, nil
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getTopN(m map[string]int, n int) []map[string]interface{} {
	type kv struct {
		Key   string
		Value int
	}
	var pairs []kv
	for k, v := range m {
		pairs = append(pairs, kv{k, v})
	}
	// Simple sort (bubble sort for small n)
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].Value < pairs[j].Value {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	result := []map[string]interface{}{}
	for i := 0; i < min(n, len(pairs)); i++ {
		result = append(result, map[string]interface{}{
			"key":   pairs[i].Key,
			"value": pairs[i].Value,
		})
	}
	return result
}

// Recommendation generators
func (h *AIHandler) getRecommendationsForUnusualAccess(count int) []string {
	if count == 0 {
		return []string{"Continue monitoring access patterns."}
	}
	return []string{
		"Consider blocking IPs with unusual access patterns.",
		"Review these IPs and their access patterns.",
		"Create a rule to automatically detect similar patterns.",
	}
}

func (h *AIHandler) getRecommendationsForUnusualPayload(count int) []string {
	if count == 0 {
		return []string{"Payloads look normal."}
	}
	return []string{
		"Review suspicious payloads for potential attacks.",
		"Consider blocking IPs sending malicious payloads.",
		"Implement stricter input validation.",
	}
}

func (h *AIHandler) getRecommendationsForHighErrorRate(rate float64) []string {
	if rate < 10 {
		return []string{"Error rate is within normal range."}
	}
	return []string{
		"Investigate the root cause of high error rates.",
		"Check if endpoints are functioning correctly.",
		"Review error logs for patterns.",
	}
}

func (h *AIHandler) getRecommendationsForRateSpike(hasSpike bool) []string {
	if !hasSpike {
		return []string{"Traffic patterns are normal."}
	}
	return []string{
		"Investigate the cause of the rate spike.",
		"Consider implementing rate limiting.",
		"Monitor for potential DDoS attacks.",
	}
}

func (h *AIHandler) getRecommendationsForSuspiciousUA(count int) []string {
	if count == 0 {
		return []string{"User agents look normal."}
	}
	return []string{
		"Review suspicious user agents.",
		"Consider blocking automated tools if not needed.",
		"Implement user agent filtering if appropriate.",
	}
}

func (h *AIHandler) getRecommendationsForFailedLogins(count int) []string {
	if count == 0 {
		return []string{"No failed login attempts detected."}
	}
	return []string{
		"Block IPs with multiple failed login attempts.",
		"Implement account lockout policies.",
		"Consider implementing CAPTCHA for login pages.",
	}
}

func (h *AIHandler) getRecommendationsForSensitivePaths(count int) []string {
	if count == 0 {
		return []string{"No access to sensitive paths detected."}
	}
	return []string{
		"Review access to sensitive paths.",
		"Ensure proper authentication is in place.",
		"Consider restricting access to admin paths.",
	}
}

func (h *AIHandler) getRecommendationsForTopAttackers(count int) []string {
	if count == 0 {
		return []string{"No attackers identified."}
	}
	return []string{
		"Consider blocking top attacking IPs.",
		"Review their request patterns.",
		"Create rules to automatically block similar patterns.",
	}
}

func (h *AIHandler) getRecommendationsForRecentIncidents(count int) []string {
	if count == 0 {
		return []string{"No recent incidents."}
	}
	return []string{
		"Review recent incidents and their causes.",
		"Ensure rules are properly configured.",
		"Consider adjusting rule thresholds if needed.",
	}
}
