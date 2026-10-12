package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"gonum.org/v1/gonum/mat"
)

type FactorizeRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

type StatsResponse struct {
	PerMatrix map[string]interface{} `json:"perMatrix"`
	Global    map[string]interface{} `json:"global"`
}

func validateMatrix(a [][]float64) (int, int, error) {
	if len(a) == 0 || len(a[0]) == 0 {
		return 0, 0, fmt.Errorf("matrix must not be empty")
	}
	cols := len(a[0])
	for i, row := range a {
		if len(row) != cols {
			return 0, 0, fmt.Errorf("matrix must be rectangular: row %d has a different length", i)
		}
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return 0, 0, fmt.Errorf("matrix values must be finite numbers")
			}
		}
	}
	if len(a) < cols {
		return 0, 0, fmt.Errorf("QR factorization requires rows >= columns (m >= n)")
	}
	return len(a), cols, nil
}

func factorize(a [][]float64) ([][]float64, [][]float64, error) {
	m, n, err := validateMatrix(a)
	if err != nil {
		return nil, nil, err
	}
	flat := make([]float64, 0, m*n)
	for _, row := range a {
		flat = append(flat, row...)
	}
	dense := mat.NewDense(m, n, flat)
	var qr mat.QR
	qr.Factorize(dense)

	var q, r mat.Dense
	qr.QTo(&q)
	qr.RTo(&r)

	qrRows, qrCols := q.Dims()
	qOut := make([][]float64, qrRows)
	for i := 0; i < qrRows; i++ {
		qOut[i] = make([]float64, qrCols)
		for j := 0; j < qrCols; j++ {
			qOut[i][j] = q.At(i, j)
		}
	}
	rRows, rCols := r.Dims()
	rOut := make([][]float64, rRows)
	for i := 0; i < rRows; i++ {
		rOut[i] = make([]float64, rCols)
		for j := 0; j < rCols; j++ {
			rOut[i][j] = r.At(i, j)
		}
	}
	return qOut, rOut, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	nodeURL := os.Getenv("NODE_API_URL")
	if nodeURL == "" {
		nodeURL = "http://localhost:3000"
	}

	app := fiber.New(fiber.Config{AppName: "Coding Challenge Go API"})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "go-api"})
	})

	app.Post("/factorize", func(c *fiber.Ctx) error {
		var req FactorizeRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		q, r, err := factorize(req.Matrix)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		payload := map[string]interface{}{"matrices": map[string]interface{}{"Q": q, "R": r}}
		body, err := json.Marshal(payload)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not encode matrices"})
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(nodeURL+"/stats", "application/json", bytes.NewReader(body))
		if err != nil {
			return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": "could not reach Node.js API", "details": err.Error(), "q": q, "r": r})
		}
		defer resp.Body.Close()
		var stats StatsResponse
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": "invalid response from Node.js API"})
		}
		if resp.StatusCode >= 400 {
			return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": "Node.js API rejected the request"})
		}
		return c.JSON(fiber.Map{"q": q, "r": r, "statistics": stats})
	})

	log.Fatal(app.Listen(":" + port))
}
