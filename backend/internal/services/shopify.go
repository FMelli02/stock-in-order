package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ShopifyService maneja la integración con Shopify
type ShopifyService struct {
	// No requiere configuración global, cada tienda tiene sus propias credenciales
}

// NewShopifyService crea una nueva instancia del servicio
func NewShopifyService() *ShopifyService {
	return &ShopifyService{}
}

// ShopifyCredentials representa las credenciales de una tienda Shopify
type ShopifyCredentials struct {
	ShopName string `json:"shop_name"` // example.myshopify.com
	APIKey   string `json:"api_key"`
	Password string `json:"password"` // API Password o Access Token
}

// TestConnection verifica que las credenciales sean válidas
func (s *ShopifyService) TestConnection(creds ShopifyCredentials) error {
	url := fmt.Sprintf("https://%s/admin/api/2024-01/shop.json", creds.ShopName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(creds.APIKey, creds.Password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error connecting to Shopify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid credentials or shop not found (status: %d)", resp.StatusCode)
	}

	return nil
}

// SyncProduct sincroniza un producto con Shopify
func (s *ShopifyService) SyncProduct(creds ShopifyCredentials, productData map[string]interface{}) error {
	url := fmt.Sprintf("https://%s/admin/api/2024-01/products.json", creds.ShopName)

	payload := map[string]interface{}{
		"product": productData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.SetBasicAuth(creds.APIKey, creds.Password)
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

// GetProducts obtiene productos de Shopify
func (s *ShopifyService) GetProducts(creds ShopifyCredentials) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://%s/admin/api/2024-01/products.json", creds.ShopName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(creds.APIKey, creds.Password)
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

	var result struct {
		Products []map[string]interface{} `json:"products"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Products, nil
}
