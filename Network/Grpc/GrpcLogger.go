package Grpc

import (
	"context"
	"crypto/x509"
	"fmt"
	"github.com/goccy/go-json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

// SessionStats represents statistics for a gRPC session
type SessionStats struct {
	SessionID       string
	StartTime       time.Time
	EndTime         time.Time
	ClientIP        string
	Method          string
	BytesSent       int
	BytesReceived   int
	PacketsSent     int
	PacketsReceived int
	TLSCertInfo     *x509.Certificate
	PayloadEntropy  float64
	Metadata        map[string][]string
}

// StatsLog defines the interface for storing session statistics
type StatsLog interface {
	Log(stats *SessionStats) error
}

// MonitoringInterceptor provides gRPC monitoring capabilities
type MonitoringInterceptor struct {
	storage  StatsLog
	sessions sync.Map // concurrent session storage
}

func NewMonitoringInterceptor(storage StatsLog) *MonitoringInterceptor {
	return &MonitoringInterceptor{storage: storage}
}

// ServerInterceptor is the gRPC unary server interceptor
func (mi *MonitoringInterceptor) ServerInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	// Create session record
	stats := &SessionStats{
		SessionID:       generateSessionID(),
		StartTime:       time.Now(),
		Method:          info.FullMethod,
		ClientIP:        getClientIP(ctx),
		TLSCertInfo:     getPeerCert(ctx),
		PacketsSent:     1, // Each call counts as one packet
		PacketsReceived: 1, // Each call counts as one packet
		Metadata:        getMetadata(ctx),
	}

	// Record request size and entropy
	if d, ok := req.([]byte); ok {
		stats.BytesReceived = len(d)
		stats.PayloadEntropy = calculateEntropy(d)
	} else if data, err := json.Marshal(req); err == nil {
		stats.BytesReceived = len(data)
		stats.PayloadEntropy = calculateEntropy(data)
	}

	// Store session
	mi.sessions.Store(stats.SessionID, stats)
	defer mi.sessions.Delete(stats.SessionID)

	// Process request and record response
	resp, err := handler(ctx, req)

	// Update statistics
	if session, ok := mi.sessions.Load(stats.SessionID); ok {
		s := session.(*SessionStats)
		s.EndTime = time.Now()

		// Record response size
		if d, ok := resp.([]byte); ok {
			s.BytesSent = len(d)
		} else if data, err := json.Marshal(resp); err == nil {
			s.BytesSent = len(data)
		}

		// Log the session
		mi.storage.Log(s)
	}

	return resp, err
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
}

// getClientIP extracts the client IP from the context
func getClientIP(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return ""
}

// getPeerCert retrieves the client's TLS certificate if available
func getPeerCert(ctx context.Context) *x509.Certificate {
	if p, ok := peer.FromContext(ctx); ok {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			if len(tlsInfo.State.PeerCertificates) > 0 {
				return tlsInfo.State.PeerCertificates[0]
			}
		}
	}
	return nil
}

// getMetadata extracts gRPC metadata from the context
func getMetadata(ctx context.Context) map[string][]string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		return md
	}
	return nil
}

// calculateEntropy computes the information entropy of the data
func calculateEntropy(data []byte) float64 {
	freq := make([]float64, 256)
	for _, b := range data {
		freq[b]++
	}

	entropy := 0.0
	for _, f := range freq {
		if f > 0 {
			p := f / float64(len(data))
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// GrpcLogger implements StatsLog for file-based logging
type GrpcLogger struct {
	FileName string
}

func (l *GrpcLogger) Log(stats *SessionStats) error {
	file, err := os.OpenFile(l.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Prepare log entry
	logEntry := map[string]interface{}{
		"session_id":       stats.SessionID,
		"start_time":       stats.StartTime.Format(time.RFC3339),
		"end_time":         stats.EndTime.Format(time.RFC3339),
		"duration_seconds": stats.EndTime.Sub(stats.StartTime).Seconds(),
		"client_ip":        stats.ClientIP,
		"method":           stats.Method,
		"bytes_sent":       stats.BytesSent,
		"bytes_received":   stats.BytesReceived,
		"packets_sent":     stats.PacketsSent,
		"packets_received": stats.PacketsReceived,
		"payload_entropy":  stats.PayloadEntropy,
		"metadata":         stats.Metadata,
	}

	// Add TLS certificate info if available
	if stats.TLSCertInfo != nil {
		certInfo := map[string]interface{}{
			"subject":        stats.TLSCertInfo.Subject,
			"issuer":         stats.TLSCertInfo.Issuer,
			"not_before":     stats.TLSCertInfo.NotBefore.Format(time.RFC3339),
			"not_after":      stats.TLSCertInfo.NotAfter.Format(time.RFC3339),
			"serial_number":  stats.TLSCertInfo.SerialNumber.String(),
			"public_key_alg": stats.TLSCertInfo.PublicKeyAlgorithm.String(),
			"signature_alg":  stats.TLSCertInfo.SignatureAlgorithm.String(),
		}
		logEntry["tls_cert_info"] = certInfo
	}

	// Encode and write to file
	logLine, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}

	_, err = file.Write(append(logLine, '\n'))
	return err
}
