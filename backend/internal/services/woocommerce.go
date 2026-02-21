package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WooCommerceService maneja la integración con WooCommerce
type WooCommerceService struct {
	// No requiere configuración global, cada tienda tiene sus propias credenciales
}

// NewWooCommerceService crea una nueva instancia del servicio
func NewWooCommerceService() *WooCommerceService {
	return &WooCommerceService{}
}

// WooCommerceCredentials representa las credenciales de una tienda WooCommerce
type WooCommerceCredentials struct {
	SiteURL        string `json:"site_url"`        // https://example.com
	ConsumerKey    string `json:"consumer_key"`    // ck_xxxxx
	ConsumerSecret string `json:"consumer_secret"` // cs_xxxxx
}

// TestConnection verifica que las credenciales sean válidas
func (w *WooCommerceService) TestConnection(creds WooCommerceCredentials) error {
	url := fmt.Sprintf("%s/wp-json/wc/v3/system_status", creds.SiteURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(creds.ConsumerKey, creds.ConsumerSecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error connecting to WooCommerce: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid credentials or site not accessible (status: %d)", resp.StatusCode)
	}

	return nil
}

// SyncProduct sincroniza un producto con WooCommerce
func (w *WooCommerceService) SyncProduct(creds WooCommerceCredentials, productData map[string]interface{}) error {
	url := fmt.Sprintf("%s/wp-json/wc/v3/products", creds.SiteURL)

	jsonData, err := json.Marshal(productData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.SetBasicAuth(creds.ConsumerKey, creds.ConsumerSecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error syncing product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to sync product (status: %d)", resp.StatusCode)
	}

	return nil
}

// GetProducts obtiene productos de WooCommerce
func (w *WooCommerceService) GetProducts(creds WooCommerceCredentials) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/wp-json/wc/v3/products", creds.SiteURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(creds.ConsumerKey, creds.ConsumerSecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching products: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch products (status: %d)", resp.StatusCode)
	}

	var products []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}

	return products, nil
}

// UpdateInventory actualiza el inventario de un producto en WooCommerce
func (w *WooCommerceService) UpdateInventory(creds WooCommerceCredentials, productID string, quantity int) error {
	url := fmt.Sprintf("%s/wp-json/wc/v3/products/%s", creds.SiteURL, productID)

	payload := map[string]interface{}{
		"stock_quantity": quantity,
		"manage_stock":   true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.SetBasicAuth(creds.ConsumerKey, creds.ConsumerSecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error updating inventory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update inventory (status: %d)", resp.StatusCode)
	}

	return nil
}
