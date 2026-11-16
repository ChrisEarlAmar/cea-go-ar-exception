package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chrisamar/cea-go-ar-exception/internal/report"
)

const reportTimeout = 2 * time.Minute

type chargedSalesRequest struct {
	DateFrom  string  `json:"dateFrom"`
	DateTo    string  `json:"dateTo"`
	CompanyID *int64  `json:"companyID,omitempty"`
	StoreCode *string `json:"storeCode,omitempty"`
}

func New(runner report.Runner) http.Handler {
	return NewWithFrontend(runner, filepath.Join("frontend", "dist"))
}

func NewWithFrontend(runner report.Runner, frontendDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("POST /api/reports/charged-sales", chargedSales(runner))
	mux.Handle("/", frontend(frontendDir))
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func chargedSales(runner report.Runner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, err := decodeRequest(w, r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}

		dateFrom, err := time.Parse("2006-01-02", input.DateFrom)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "dateFrom must use YYYY-MM-DD"})
			return
		}
		dateTo, err := time.Parse("2006-01-02", input.DateTo)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "dateTo must use YYYY-MM-DD"})
			return
		}
		if dateTo.Before(dateFrom) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "dateTo cannot be before dateFrom"})
			return
		}

		filter := report.ChargedSalesFilter{
			DateFrom:  dateFrom,
			DateTo:    dateTo,
			CompanyID: input.CompanyID,
		}
		if input.StoreCode != nil {
			storeCode := strings.TrimSpace(*input.StoreCode)
			if len(storeCode) > 50 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"message": "storeCode must be 50 characters or fewer"})
				return
			}
			if storeCode != "" {
				filter.StoreCode = &storeCode
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), reportTimeout)
		defer cancel()
		result, err := runner.ChargedSales(ctx, filter)
		if err != nil {
			log.Printf("charged sales query failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Unable to generate the Charged Sales report."})
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func decodeRequest(w http.ResponseWriter, r *http.Request) (chargedSalesRequest, error) {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	var input chargedSalesRequest
	if err := decoder.Decode(&input); err != nil {
		return chargedSalesRequest{}, errors.New("request body must be valid JSON")
	}
	if input.DateFrom == "" || input.DateTo == "" {
		return chargedSalesRequest{}, errors.New("dateFrom and dateTo are required")
	}
	return input, nil
}

func frontend(directory string) http.Handler {
	files := http.FileServer(http.Dir(directory))
	index := filepath.Join(directory, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		requested := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if requested == "." {
			requested = "index.html"
		}
		if requested == ".." || strings.HasPrefix(requested, ".."+string(filepath.Separator)) {
			http.NotFound(w, r)
			return
		}

		if info, err := os.Stat(filepath.Join(directory, requested)); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, index)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}
