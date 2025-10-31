package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MercadoLibreService maneja las interacciones con la API de Mercado Libre
type MercadoLibreService struct{}

// NewMercadoLibreService crea una nueva instancia del servicio
func NewMercadoLibreService() *MercadoLibreService {
	return &MercadoLibreService{}
}

// MLOrder representa una orden de Mercado Libre
type MLOrder struct {
	ID     int64  `json:"id"`
	Status string `json:"status"` // confirmed, payment_required, payment_in_process, partially_paid, paid, cancelled
	Buyer  struct {
		ID        int64  `json:"id"`
		Nickname  string `json:"nickname"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     struct {
			AreaCode string `json:"area_code"`
			Number   string `json:"number"`
		} `json:"phone"`
	} `json:"buyer"`
	OrderItems []struct {
		Item struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			CategoryID  string `json:"category_id"`
			VariationID int64  `json:"variation_id"`
			SellerSKU   string `json:"seller_custom_field"` // SKU del vendedor
		} `json:"item"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unit_price"`
		FullUnitPrice float64 `json:"full_unit_price"`
		SalePrice     float64 `json:"sale_fee"`
		ListingTypeID string  `json:"listing_type_id"`
	} `json:"order_items"`
	Payments []struct {
		ID     int64   `json:"id"`
		Status string  `json:"status"`
		Total  float64 `json:"transaction_amount"`
	} `json:"payments"`
	Shipping struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	} `json:"shipping"`
	DateCreated string  `json:"date_created"`
	DateClosed  string  `json:"date_closed"`
	TotalAmount float64 `json:"total_amount"`
	CurrencyID  string  `json:"currency_id"`
}

// GetOrder obtiene los detalles completos de una orden de Mercado Libre
func (s *MercadoLibreService) GetOrder(orderID int64, accessToken string) (*MLOrder, error) {
	orderURL := fmt.Sprintf("https://api.mercadolibre.com/orders/%d", orderID)

	req, err := http.NewRequest("GET", orderURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making order request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mercadolibre API error (status %d): %s", resp.StatusCode, string(body))
	}

	var order MLOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("error parsing order: %w", err)
	}

	return &order, nil
}
